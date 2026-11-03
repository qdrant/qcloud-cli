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
	// Store is StoreKeychain when the access and refresh tokens are in the OS keychain; credentials.yaml
	// then holds only the fields that are not secret. Empty means both are in credentials.yaml.
	Store        string    `yaml:"store,omitempty"`
	AccessToken  string    `yaml:"access_token,omitempty"`
	RefreshToken string    `yaml:"refresh_token,omitempty"`
	TokenType    string    `yaml:"token_type,omitempty"`
	Expiry       time.Time `yaml:"expiry,omitempty"`
	// Scope is the requested scope: what qcloud asked the login server for (read-only or manage).
	Scope string `yaml:"requested_scope,omitempty"`
	// GrantedScope is what the user granted on the consent page (read-only or manage), which can be
	// narrower than the requested scope. Qdrant Cloud enforces it on every call.
	GrantedScope string `yaml:"granted_scope,omitempty"`
	// Email is the user the session belongs to, for display.
	Email    string `yaml:"email,omitempty"`
	Issuer   string `yaml:"issuer,omitempty"`
	ClientID string `yaml:"client_id,omitempty"`
	Resource string `yaml:"resource,omitempty"`
}

// expiryDelta is how long before its expiry an access token is refreshed.
const expiryDelta = 30 * time.Second

// Expired reports whether the access token should be refreshed.
func (t Token) Expired(now time.Time) bool {
	if t.Expiry.IsZero() {
		return false
	}

	return !t.Expiry.After(now.Add(expiryDelta))
}

// tokenStore keeps the sessions of all endpoints: credentials.yaml next to the config file, and the
// access and refresh tokens in the OS keychain unless the mode says otherwise.
type tokenStore struct {
	dir     func() string
	mode    func() string
	secrets SecretStore
}

func (s tokenStore) path() string {
	return filepath.Join(s.dir(), credentialsFileName)
}

type credentialsFile struct {
	Tokens []Token `yaml:"tokens,omitempty"`
}

// load returns the stored token for endpoint, with its access and refresh tokens, or false if none.
func (s tokenStore) load(endpoint string) (Token, bool, error) {
	tok, ok, err := s.info(endpoint)
	if err != nil || !ok || tok.Store != StoreKeychain {
		return tok, ok, err
	}

	tok, err = getSecrets(s.secrets, s.path(), tok)
	if err != nil {
		return Token{}, false, err
	}

	return tok, true, nil
}

// info returns the stored token for endpoint like load, but without reading the OS keychain: when the
// tokens are kept there, AccessToken and RefreshToken are empty.
func (s tokenStore) info(endpoint string) (Token, bool, error) {
	file, err := readCredentials(s.path())
	if err != nil {
		return Token{}, false, err
	}

	tok, ok := file.find(endpoint)

	return tok, ok, nil
}

// save upserts a token for its endpoint. The access and refresh tokens go to the OS keychain and the
// rest to credentials.yaml, as the mode says. A token that was in credentials.yaml moves to the keychain
// on its next save. keychainErr is why the tokens went to credentials.yaml although the keychain was
// to be tried.
func (s tokenStore) save(tok Token) (keychainErr error, err error) {
	mode, err := parseStoreMode(s.mode())
	if err != nil {
		return nil, err
	}

	path := s.path()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("creating credentials dir: %w", err)
	}

	file, err := readCredentials(path)
	if err != nil {
		return nil, err
	}

	previous, _ := file.find(tok.Endpoint)
	tok.Store = ""

	if mode != StoreFile {
		err := setSecrets(s.secrets, path, tok)
		switch {
		case err == nil:
			tok.Store = StoreKeychain
			tok.AccessToken = ""
			tok.RefreshToken = ""
		case mode == StoreKeychain:
			return nil, fmt.Errorf("storing the OAuth tokens in the OS keychain: %w (set QDRANT_CLOUD_TOKEN_STORE=file to use credentials.yaml)", err)
		default:
			keychainErr = err
		}
	}

	// Tokens that move to the file must not stay behind in the keychain.
	if tok.Store != StoreKeychain && previous.Store == StoreKeychain {
		_ = deleteSecrets(s.secrets, path, tok.Endpoint)
	}

	file.upsert(tok)

	return keychainErr, writeCredentials(path, file)
}

// delete removes the stored token for endpoint, from credentials.yaml and the OS keychain. The session
// is gone once credentials.yaml is written; keychainErr is why its keychain entries could not be removed.
func (s tokenStore) delete(endpoint string) (keychainErr error, err error) {
	path := s.path()
	file, err := readCredentials(path)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	// Without a keychain entry there may be no keychain at all (headless Linux, CI); only a failure to
	// remove tokens that are known to be there is reported.
	if err := deleteSecrets(s.secrets, path, endpoint); err != nil && previous.Store == StoreKeychain {
		return fmt.Errorf("could not remove the OAuth tokens from the OS keychain: %w", err), nil
	}

	return nil, nil
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
