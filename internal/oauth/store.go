package oauth

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const credentialsFileName = "credentials.yaml"

// Token is a stored OAuth session for one gRPC endpoint.
type Token struct {
	Endpoint string `yaml:"endpoint"`
	// Store is "keychain" when the access and refresh tokens are in the OS keychain; credentials.yaml
	// then holds only the fields that are not secret. Empty means both are in credentials.yaml.
	Store        string    `yaml:"store,omitempty"`
	AccessToken  string    `yaml:"access_token,omitempty"`
	RefreshToken string    `yaml:"refresh_token,omitempty"`
	TokenType    string    `yaml:"token_type,omitempty"`
	Expiry       time.Time `yaml:"expiry,omitempty"`
	// Scope is the requested scope: what qcloud asked Auth0 for, and what the token's scope claim says.
	Scope string `yaml:"requested_scope,omitempty"`
	// GrantedScope is what the user granted on the consent page (read-only or manage), which can be
	// narrower than the requested scope. iam-service enforces it on every call.
	GrantedScope string `yaml:"granted_scope,omitempty"`
	Issuer       string `yaml:"issuer,omitempty"`
	ClientID     string `yaml:"client_id,omitempty"`
	Resource     string `yaml:"resource,omitempty"`
}

// Expired reports whether the access token should be refreshed.
func (t Token) Expired(now time.Time) bool {
	if t.Expiry.IsZero() {
		return false
	}

	return !t.Expiry.After(now.Add(30 * time.Second))
}

type credentialsFile struct {
	Tokens []Token `yaml:"tokens,omitempty"`
}

// CredentialsPath is credentials.yaml next to the CLI config file.
func CredentialsPath(configFilePath string) string {
	dir := filepath.Join(os.Getenv("HOME"), ".config", "qcloud")
	if configFilePath != "" {
		dir = filepath.Dir(configFilePath)
	} else if home, err := os.UserHomeDir(); err == nil {
		dir = filepath.Join(home, ".config", "qcloud")
	}

	return filepath.Join(dir, credentialsFileName)
}

// LoadToken returns the stored token for endpoint, with its access and refresh tokens, or false if
// none.
func LoadToken(path, endpoint string) (Token, bool, error) {
	return loadToken(defaultSecrets, path, endpoint)
}

// LoadTokenInfo returns the stored token for endpoint like LoadToken, but without reading the OS
// keychain: when the tokens are kept there, AccessToken and RefreshToken are empty.
func LoadTokenInfo(path, endpoint string) (Token, bool, error) {
	file, err := readCredentials(path)
	if err != nil {
		return Token{}, false, err
	}

	tok, ok := file.find(endpoint)

	return tok, ok, nil
}

func loadToken(secrets secretStore, path, endpoint string) (Token, bool, error) {
	tok, ok, err := LoadTokenInfo(path, endpoint)
	if err != nil || !ok || tok.Store != StoreKeychain {
		return tok, ok, err
	}

	tok, err = getSecrets(secrets, path, tok)
	if err != nil {
		return Token{}, false, err
	}

	return tok, true, nil
}

// Saved says where SaveToken put the access and refresh tokens.
type Saved struct {
	// Store is StoreKeychain or StoreFile.
	Store string
	// KeychainErr is why the tokens went to credentials.yaml although the keychain was to be tried.
	KeychainErr error
}

// SaveToken upserts a token for its endpoint. The access and refresh tokens go to the OS keychain and
// the rest to credentials.yaml; see TokenStoreEnv. A token that was in credentials.yaml moves to the
// keychain on its next save.
func SaveToken(path string, tok Token) (Saved, error) {
	mode, err := storeMode()
	if err != nil {
		return Saved{}, err
	}

	return saveToken(defaultSecrets, mode, path, tok)
}

func saveToken(secrets secretStore, mode, path string, tok Token) (Saved, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return Saved{}, fmt.Errorf("creating credentials dir: %w", err)
	}

	file, err := readCredentials(path)
	if err != nil {
		return Saved{}, err
	}

	previous, _ := file.find(tok.Endpoint)
	saved := Saved{Store: StoreFile}
	tok.Store = ""

	if mode != StoreFile {
		err := setSecrets(secrets, path, tok)
		switch {
		case err == nil:
			saved.Store = StoreKeychain
			tok.Store = StoreKeychain
			tok.AccessToken = ""
			tok.RefreshToken = ""
		case mode == StoreKeychain:
			return Saved{}, fmt.Errorf("storing the OAuth tokens in the OS keychain: %w (set %s=file to use credentials.yaml)", err, TokenStoreEnv)
		default:
			saved.KeychainErr = err
		}
	}

	// Tokens that move to the file must not stay behind in the keychain.
	if saved.Store == StoreFile && previous.Store == StoreKeychain {
		_ = deleteSecrets(secrets, path, tok.Endpoint)
	}

	file.upsert(tok)

	return saved, writeCredentials(path, file)
}

// DeleteToken removes the stored token for endpoint, from credentials.yaml and the OS keychain.
func DeleteToken(path, endpoint string) error {
	return deleteToken(defaultSecrets, path, endpoint)
}

func deleteToken(secrets secretStore, path, endpoint string) error {
	file, err := readCredentials(path)
	if err != nil {
		return err
	}

	previous, _ := file.find(endpoint)
	filtered := file.Tokens[:0]
	for _, tok := range file.Tokens {
		if tok.Endpoint != endpoint {
			filtered = append(filtered, tok)
		}
	}

	file.Tokens = filtered
	if err := writeCredentials(path, file); err != nil {
		return err
	}

	// Without a keychain entry there may be no keychain at all (headless Linux, CI); only a failure to
	// remove tokens that are known to be there is an error.
	if err := deleteSecrets(secrets, path, endpoint); err != nil && previous.Store == StoreKeychain {
		return fmt.Errorf("removing the OAuth tokens from the OS keychain: %w", err)
	}

	return nil
}

func (f credentialsFile) find(endpoint string) (Token, bool) {
	for _, tok := range f.Tokens {
		if tok.Endpoint == endpoint {
			return tok, true
		}
	}

	return Token{}, false
}

func (f *credentialsFile) upsert(tok Token) {
	for i, existing := range f.Tokens {
		if existing.Endpoint == tok.Endpoint {
			f.Tokens[i] = tok
			return
		}
	}

	f.Tokens = append(f.Tokens, tok)
}

func readCredentials(path string) (credentialsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return credentialsFile{}, nil
		}

		return credentialsFile{}, fmt.Errorf("reading credentials: %w", err)
	}

	var file credentialsFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return credentialsFile{}, fmt.Errorf("parsing credentials: %w", err)
	}

	return file, nil
}

func writeCredentials(path string, file credentialsFile) error {
	data, err := yaml.Marshal(file)
	if err != nil {
		return fmt.Errorf("marshaling credentials: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing credentials: %w", err)
	}

	return nil
}
