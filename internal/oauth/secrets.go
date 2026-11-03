package oauth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/zalando/go-keyring"
)

// Token store modes (config token_store, env QDRANT_CLOUD_TOKEN_STORE), and the value of Token.Store:
//   - "" or StoreAuto: the OS keychain, or credentials.yaml (0600) when no keychain is available;
//   - StoreKeychain: the OS keychain only; saving fails when it is not available;
//   - StoreFile: credentials.yaml only.
const (
	StoreAuto     = "auto"
	StoreKeychain = "keychain"
	StoreFile     = "file"
)

// keychainService is the service name of qcloud's entries in the OS keychain.
const keychainService = "qcloud"

// SecretStore keeps the access and refresh tokens: the OS keychain (macOS Keychain, Windows Credential
// Manager, or the Secret Service on Linux), or MemorySecrets in tests.
type SecretStore interface {
	Set(service, user, secret string) error
	Get(service, user string) (string, error)
	Delete(service, user string) error
}

// ErrSecretNotFound is the error a SecretStore returns for an entry that does not exist.
var ErrSecretNotFound = keyring.ErrNotFound

// OSKeychain is the SecretStore of the operating system.
type OSKeychain struct{}

// Set stores secret in the keychain.
func (OSKeychain) Set(service, user, secret string) error { return keyring.Set(service, user, secret) }

// Get reads a secret from the keychain, or returns ErrSecretNotFound.
func (OSKeychain) Get(service, user string) (string, error) { return keyring.Get(service, user) }

// Delete removes a secret from the keychain, or returns ErrSecretNotFound.
func (OSKeychain) Delete(service, user string) error { return keyring.Delete(service, user) }

// MemorySecrets is an in-memory SecretStore, so that tests never read or write the keychain of the
// machine they run on.
type MemorySecrets struct {
	mu      sync.Mutex
	secrets map[string]string
}

// NewMemorySecrets returns an empty MemorySecrets.
func NewMemorySecrets() *MemorySecrets {
	return &MemorySecrets{secrets: map[string]string{}}
}

// Set stores secret.
func (m *MemorySecrets) Set(service, user, secret string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.secrets[service+"\x00"+user] = secret

	return nil
}

// Get returns a secret, or ErrSecretNotFound.
func (m *MemorySecrets) Get(service, user string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	secret, ok := m.secrets[service+"\x00"+user]
	if !ok {
		return "", ErrSecretNotFound
	}

	return secret, nil
}

// Delete removes a secret, or returns ErrSecretNotFound.
func (m *MemorySecrets) Delete(service, user string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := service + "\x00" + user
	if _, ok := m.secrets[key]; !ok {
		return ErrSecretNotFound
	}

	delete(m.secrets, key)

	return nil
}

// parseStoreMode validates the configured token store mode.
func parseStoreMode(mode string) (string, error) {
	switch mode {
	case "", StoreAuto:
		return StoreAuto, nil
	case StoreKeychain, StoreFile:
		return mode, nil
	default:
		return "", fmt.Errorf("token store %q: use auto, keychain, or file (QDRANT_CLOUD_TOKEN_STORE)", mode)
	}
}

// keychainUser names a token's keychain entry: the endpoint, the kind of token ("access" or
// "refresh"), and a short hash of the credentials file, so that two config directories keep separate
// sessions.
func keychainUser(path, endpoint, kind string) string {
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}

	sum := sha256.Sum256([]byte(path))

	return fmt.Sprintf("%s %s %s", endpoint, kind, hex.EncodeToString(sum[:4]))
}

// setSecrets stores the access and refresh tokens in the keychain, as two entries: Windows limits an
// entry to 2.5 KB, which a JWT and a refresh token together can exceed.
func setSecrets(secrets SecretStore, path string, tok Token) error {
	if err := secrets.Set(keychainService, keychainUser(path, tok.Endpoint, "access"), tok.AccessToken); err != nil {
		return err
	}

	if tok.RefreshToken == "" {
		return deleteSecret(secrets, path, tok.Endpoint, "refresh")
	}

	return secrets.Set(keychainService, keychainUser(path, tok.Endpoint, "refresh"), tok.RefreshToken)
}

// getSecrets fills in the access and refresh tokens of a token kept in the keychain.
func getSecrets(secrets SecretStore, path string, tok Token) (Token, error) {
	access, err := secrets.Get(keychainService, keychainUser(path, tok.Endpoint, "access"))
	if err != nil {
		return Token{}, keychainReadError(tok.Endpoint, err)
	}

	refresh, err := secrets.Get(keychainService, keychainUser(path, tok.Endpoint, "refresh"))
	if err != nil && !errors.Is(err, ErrSecretNotFound) {
		return Token{}, keychainReadError(tok.Endpoint, err)
	}

	tok.AccessToken = access
	tok.RefreshToken = refresh

	return tok, nil
}

func keychainReadError(endpoint string, err error) error {
	if errors.Is(err, ErrSecretNotFound) {
		return fmt.Errorf("the OAuth tokens for %s are missing from the OS keychain (run \"qcloud auth login\")", endpoint)
	}

	return fmt.Errorf("reading the OAuth tokens for %s from the OS keychain: %w", endpoint, err)
}

// deleteSecrets removes both keychain entries of endpoint; entries that do not exist are fine.
func deleteSecrets(secrets SecretStore, path, endpoint string) error {
	return errors.Join(
		deleteSecret(secrets, path, endpoint, "access"),
		deleteSecret(secrets, path, endpoint, "refresh"),
	)
}

func deleteSecret(secrets SecretStore, path, endpoint, kind string) error {
	err := secrets.Delete(keychainService, keychainUser(path, endpoint, kind))
	if err != nil && !errors.Is(err, ErrSecretNotFound) {
		return err
	}

	return nil
}
