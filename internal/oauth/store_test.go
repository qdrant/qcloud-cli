package oauth

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testToken() Token {
	return Token{
		Endpoint:     "grpc.development-cloud.qdrant.io:443",
		AccessToken:  "access",
		RefreshToken: "refresh",
		Scope:        "manage",
		Expiry:       time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC),
	}
}

// failingSecrets is a machine without a keychain.
type failingSecrets struct{}

var errNoKeychain = errors.New("no keychain")

func (failingSecrets) Set(string, string, string) error   { return errNoKeychain }
func (failingSecrets) Get(string, string) (string, error) { return "", errNoKeychain }
func (failingSecrets) Delete(string, string) error        { return errNoKeychain }

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	return string(data)
}

func TestTokenStore_KeepsTheTokensInTheKeychain(t *testing.T) {
	t.Parallel()

	secrets := newMemorySecrets()
	path := filepath.Join(t.TempDir(), "credentials.yaml")
	tok := testToken()

	saved, err := saveToken(secrets, StoreAuto, path, tok)
	require.NoError(t, err)
	assert.Equal(t, StoreKeychain, saved.Store)
	require.NoError(t, saved.KeychainErr)

	file := readFile(t, path)
	assert.NotContains(t, file, "access_token")
	assert.NotContains(t, file, "refresh_token")
	assert.Contains(t, file, "store: keychain")

	got, ok, err := loadToken(secrets, path, tok.Endpoint)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, "access", got.AccessToken)
	assert.Equal(t, "refresh", got.RefreshToken)
	assert.Equal(t, "manage", got.Scope)

	info, ok, err := LoadTokenInfo(path, tok.Endpoint)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Empty(t, info.AccessToken, "LoadTokenInfo does not read the keychain")
	assert.Equal(t, tok.Expiry, info.Expiry)

	require.NoError(t, deleteToken(secrets, path, tok.Endpoint))
	_, ok, err = loadToken(secrets, path, tok.Endpoint)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Empty(t, secrets.secrets, "logout removes the keychain entries")
}

func TestTokenStore_FallsBackToTheFileWithoutAKeychain(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "credentials.yaml")
	tok := testToken()

	saved, err := saveToken(failingSecrets{}, StoreAuto, path, tok)
	require.NoError(t, err)
	assert.Equal(t, StoreFile, saved.Store)
	require.ErrorIs(t, saved.KeychainErr, errNoKeychain)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	got, ok, err := loadToken(failingSecrets{}, path, tok.Endpoint)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, "access", got.AccessToken)
	assert.Equal(t, "refresh", got.RefreshToken)

	require.NoError(t, deleteToken(failingSecrets{}, path, tok.Endpoint), "no keychain is not an error on logout")
}

func TestTokenStore_KeychainModeFailsWithoutAKeychain(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "credentials.yaml")

	_, err := saveToken(failingSecrets{}, StoreKeychain, path, testToken())
	require.ErrorIs(t, err, errNoKeychain)
	assert.NoFileExists(t, path)
}

func TestTokenStore_FileModeSkipsTheKeychain(t *testing.T) {
	t.Parallel()

	secrets := newMemorySecrets()
	path := filepath.Join(t.TempDir(), "credentials.yaml")

	saved, err := saveToken(secrets, StoreFile, path, testToken())
	require.NoError(t, err)
	assert.Equal(t, StoreFile, saved.Store)
	require.NoError(t, saved.KeychainErr)
	assert.Empty(t, secrets.secrets)
	assert.Contains(t, readFile(t, path), "refresh_token: refresh")
}

func TestTokenStore_MovesAFileTokenToTheKeychainOnTheNextSave(t *testing.T) {
	t.Parallel()

	secrets := newMemorySecrets()
	path := filepath.Join(t.TempDir(), "credentials.yaml")
	tok := testToken()

	// A session saved before qcloud used the keychain.
	_, err := saveToken(secrets, StoreFile, path, tok)
	require.NoError(t, err)

	got, ok, err := loadToken(secrets, path, tok.Endpoint)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, "access", got.AccessToken)

	// The next refresh saves it again.
	got.AccessToken = "refreshed"
	_, err = saveToken(secrets, StoreAuto, path, got)
	require.NoError(t, err)
	assert.NotContains(t, readFile(t, path), "refresh_token")

	got, _, err = loadToken(secrets, path, tok.Endpoint)
	require.NoError(t, err)
	assert.Equal(t, "refreshed", got.AccessToken)
	assert.Equal(t, "refresh", got.RefreshToken)
}

func TestTokenStore_MovingBackToTheFileClearsTheKeychain(t *testing.T) {
	t.Parallel()

	secrets := newMemorySecrets()
	path := filepath.Join(t.TempDir(), "credentials.yaml")

	_, err := saveToken(secrets, StoreAuto, path, testToken())
	require.NoError(t, err)
	_, err = saveToken(secrets, StoreFile, path, testToken())
	require.NoError(t, err)

	assert.Empty(t, secrets.secrets)
}

func TestTokenStore_MissingKeychainEntryAsksToLogIn(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "credentials.yaml")
	tok := testToken()
	_, err := saveToken(newMemorySecrets(), StoreAuto, path, tok)
	require.NoError(t, err)

	// Another machine's keychain, or an entry deleted by hand.
	_, _, err = loadToken(newMemorySecrets(), path, tok.Endpoint)
	require.ErrorContains(t, err, "qcloud auth login")
}

func TestTokenStore_ConfigDirectoriesHaveSeparateKeychainEntries(t *testing.T) {
	t.Parallel()

	secrets := newMemorySecrets()
	a := filepath.Join(t.TempDir(), "credentials.yaml")
	b := filepath.Join(t.TempDir(), "credentials.yaml")
	tok := testToken()

	_, err := saveToken(secrets, StoreAuto, a, tok)
	require.NoError(t, err)

	tok.AccessToken = "other"
	_, err = saveToken(secrets, StoreAuto, b, tok)
	require.NoError(t, err)

	got, _, err := loadToken(secrets, a, tok.Endpoint)
	require.NoError(t, err)
	assert.Equal(t, "access", got.AccessToken)
}

func TestStoreMode(t *testing.T) {
	for value, want := range map[string]string{"": StoreAuto, "auto": StoreAuto, "keychain": StoreKeychain, "file": StoreFile} {
		t.Setenv(TokenStoreEnv, value)
		got, err := storeMode()
		require.NoError(t, err)
		assert.Equal(t, want, got)
	}

	t.Setenv(TokenStoreEnv, "vault")
	_, err := storeMode()
	require.Error(t, err)
}

func TestToken_Expired(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.True(t, Token{Expiry: now.Add(10 * time.Second)}.Expired(now))
	assert.False(t, Token{Expiry: now.Add(2 * time.Minute)}.Expired(now))
	assert.False(t, Token{}.Expired(now))
}
