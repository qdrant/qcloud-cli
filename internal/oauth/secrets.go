package oauth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/zalando/go-keyring"
)

// TokenStoreEnv selects where the access and refresh tokens are kept:
//   - "" or "auto": the OS keychain, or credentials.yaml (0600) when no keychain is available;
//   - "keychain": the OS keychain only; saving fails when it is not available;
//   - "file": credentials.yaml only.
const TokenStoreEnv = "QDRANT_CLOUD_TOKEN_STORE"

// Token store modes and the value of Token.Store.
const (
	StoreAuto     = "auto"
	StoreKeychain = "keychain"
	StoreFile     = "file"
)

// keychainService is the service name of qcloud's entries in the OS keychain.
const keychainService = "qcloud"

// secretStore is the OS keychain: macOS Keychain, Windows Credential Manager, or the Secret Service
// (GNOME Keyring, KWallet) on Linux.
type secretStore interface {
	Set(service, user, secret string) error
	Get(service, user string) (string, error)
	Delete(service, user string) error
}

type osKeychain struct{}

func (osKeychain) Set(service, user, secret string) error   { return keyring.Set(service, user, secret) }
func (osKeychain) Get(service, user string) (string, error) { return keyring.Get(service, user) }
func (osKeychain) Delete(service, user string) error        { return keyring.Delete(service, user) }

// defaultSecrets is the OS keychain, except in test binaries: tests must never read or write the
// keychain of the machine they run on.
var defaultSecrets = func() secretStore {
	if testing.Testing() {
		return newMemorySecrets()
	}

	return osKeychain{}
}()

// memorySecrets is an in-memory secretStore for tests.
type memorySecrets struct {
	mu      sync.Mutex
	secrets map[string]string
}

func newMemorySecrets() *memorySecrets {
	return &memorySecrets{secrets: map[string]string{}}
}

func (m *memorySecrets) Set(service, user, secret string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.secrets[service+"\x00"+user] = secret

	return nil
}

func (m *memorySecrets) Get(service, user string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	secret, ok := m.secrets[service+"\x00"+user]
	if !ok {
		return "", keyring.ErrNotFound
	}

	return secret, nil
}

func (m *memorySecrets) Delete(service, user string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := service + "\x00" + user
	if _, ok := m.secrets[key]; !ok {
		return keyring.ErrNotFound
	}

	delete(m.secrets, key)

	return nil
}

// storeMode reads TokenStoreEnv.
func storeMode() (string, error) {
	switch mode := os.Getenv(TokenStoreEnv); mode {
	case "", StoreAuto:
		return StoreAuto, nil
	case StoreKeychain, StoreFile:
		return mode, nil
	default:
		return "", fmt.Errorf("%s=%q: use auto, keychain, or file", TokenStoreEnv, mode)
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
func setSecrets(secrets secretStore, path string, tok Token) error {
	if err := secrets.Set(keychainService, keychainUser(path, tok.Endpoint, "access"), tok.AccessToken); err != nil {
		return err
	}

	if tok.RefreshToken == "" {
		return deleteSecret(secrets, path, tok.Endpoint, "refresh")
	}

	return secrets.Set(keychainService, keychainUser(path, tok.Endpoint, "refresh"), tok.RefreshToken)
}

// getSecrets fills in the access and refresh tokens of a token kept in the keychain.
func getSecrets(secrets secretStore, path string, tok Token) (Token, error) {
	access, err := secrets.Get(keychainService, keychainUser(path, tok.Endpoint, "access"))
	if err != nil {
		return Token{}, keychainReadError(tok.Endpoint, err)
	}

	refresh, err := secrets.Get(keychainService, keychainUser(path, tok.Endpoint, "refresh"))
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return Token{}, keychainReadError(tok.Endpoint, err)
	}

	tok.AccessToken = access
	tok.RefreshToken = refresh

	return tok, nil
}

func keychainReadError(endpoint string, err error) error {
	if errors.Is(err, keyring.ErrNotFound) {
		return fmt.Errorf("the OAuth tokens for %s are missing from the OS keychain (run \"qcloud auth login\")", endpoint)
	}

	return fmt.Errorf("reading the OAuth tokens for %s from the OS keychain: %w", endpoint, err)
}

// deleteSecrets removes both keychain entries of endpoint; entries that do not exist are fine.
func deleteSecrets(secrets secretStore, path, endpoint string) error {
	return errors.Join(
		deleteSecret(secrets, path, endpoint, "access"),
		deleteSecret(secrets, path, endpoint, "refresh"),
	)
}

func deleteSecret(secrets secretStore, path, endpoint, kind string) error {
	err := secrets.Delete(keychainService, keychainUser(path, endpoint, kind))
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return err
	}

	return nil
}
