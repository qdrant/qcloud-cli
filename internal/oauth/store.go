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
	Endpoint     string    `yaml:"endpoint"`
	AccessToken  string    `yaml:"access_token"`
	RefreshToken string    `yaml:"refresh_token,omitempty"`
	TokenType    string    `yaml:"token_type,omitempty"`
	Expiry       time.Time `yaml:"expiry,omitempty"`
	Scope        string    `yaml:"scope,omitempty"`
	Issuer       string    `yaml:"issuer,omitempty"`
	ClientID     string    `yaml:"client_id,omitempty"`
	Resource     string    `yaml:"resource,omitempty"`
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

// LoadToken returns the stored token for endpoint, or false if none.
func LoadToken(path, endpoint string) (Token, bool, error) {
	file, err := readCredentials(path)
	if err != nil {
		return Token{}, false, err
	}

	for _, tok := range file.Tokens {
		if tok.Endpoint == endpoint {
			return tok, true, nil
		}
	}

	return Token{}, false, nil
}

// SaveToken upserts a token for its endpoint.
func SaveToken(path string, tok Token) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("creating credentials dir: %w", err)
	}

	file, err := readCredentials(path)
	if err != nil {
		return err
	}

	replaced := false
	for i, existing := range file.Tokens {
		if existing.Endpoint == tok.Endpoint {
			file.Tokens[i] = tok
			replaced = true
			break
		}
	}

	if !replaced {
		file.Tokens = append(file.Tokens, tok)
	}

	return writeCredentials(path, file)
}

// DeleteToken removes the stored token for endpoint.
func DeleteToken(path, endpoint string) error {
	file, err := readCredentials(path)
	if err != nil {
		return err
	}

	filtered := file.Tokens[:0]
	for _, tok := range file.Tokens {
		if tok.Endpoint != endpoint {
			filtered = append(filtered, tok)
		}
	}

	file.Tokens = filtered
	return writeCredentials(path, file)
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
