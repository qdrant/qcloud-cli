package oauth

import (
	"errors"
	"os"
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

// newStore is a token store in a new directory, with secrets and mode.
func newStore(t *testing.T, secrets SecretStore, mode string) tokenStore {
	t.Helper()
	dir := t.TempDir()

	return tokenStore{dir: func() string { return dir }, mode: func() string { return mode }, secrets: secrets}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	return string(data)
}

func TestTokenStore_KeepsTheTokensInTheKeychain(t *testing.T) {
	t.Parallel()

	secrets := NewMemorySecrets()
	store := newStore(t, secrets, StoreAuto)
	tok := testToken()

	keychainErr, err := store.save(tok)
	require.NoError(t, err)
	require.NoError(t, keychainErr)

	file := readFile(t, store.path())
	assert.NotContains(t, file, "access_token")
	assert.NotContains(t, file, "refresh_token")
	assert.Contains(t, file, "store: keychain")

	got, ok, err := store.load(tok.Endpoint)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, "access", got.AccessToken)
	assert.Equal(t, "refresh", got.RefreshToken)
	assert.Equal(t, "manage", got.Scope)

	info, ok, err := store.info(tok.Endpoint)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Empty(t, info.AccessToken, "info does not read the keychain")
	assert.Equal(t, tok.Expiry, info.Expiry)

	keychainErr, err = store.delete(tok.Endpoint)
	require.NoError(t, err)
	require.NoError(t, keychainErr)
	_, ok, err = store.load(tok.Endpoint)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Empty(t, secrets.secrets, "logout removes the keychain entries")
}

func TestTokenStore_FallsBackToTheFileWithoutAKeychain(t *testing.T) {
	t.Parallel()

	store := newStore(t, failingSecrets{}, StoreAuto)
	tok := testToken()

	keychainErr, err := store.save(tok)
	require.NoError(t, err)
	require.ErrorIs(t, keychainErr, errNoKeychain)

	info, err := os.Stat(store.path())
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	got, ok, err := store.load(tok.Endpoint)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, "access", got.AccessToken)
	assert.Equal(t, "refresh", got.RefreshToken)

	keychainErr, err = store.delete(tok.Endpoint)
	require.NoError(t, err)
	require.NoError(t, keychainErr, "no keychain is not an error on logout")
}

func TestTokenStore_KeychainModeFailsWithoutAKeychain(t *testing.T) {
	t.Parallel()

	store := newStore(t, failingSecrets{}, StoreKeychain)

	_, err := store.save(testToken())
	require.ErrorIs(t, err, errNoKeychain)
	assert.NoFileExists(t, store.path())
}

func TestTokenStore_FileModeSkipsTheKeychain(t *testing.T) {
	t.Parallel()

	secrets := NewMemorySecrets()
	store := newStore(t, secrets, StoreFile)

	keychainErr, err := store.save(testToken())
	require.NoError(t, err)
	require.NoError(t, keychainErr)
	assert.Empty(t, secrets.secrets)
	assert.Contains(t, readFile(t, store.path()), "refresh_token: refresh")
}

func TestTokenStore_MovesAFileTokenToTheKeychainOnTheNextSave(t *testing.T) {
	t.Parallel()

	secrets := NewMemorySecrets()
	fileStore := newStore(t, secrets, StoreFile)
	store := fileStore
	store.mode = func() string { return StoreAuto }
	tok := testToken()

	// A session saved before qcloud used the keychain.
	_, err := fileStore.save(tok)
	require.NoError(t, err)

	got, ok, err := store.load(tok.Endpoint)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, "access", got.AccessToken)

	// The next refresh saves it again.
	got.AccessToken = "refreshed"
	_, err = store.save(got)
	require.NoError(t, err)
	assert.NotContains(t, readFile(t, store.path()), "refresh_token")

	got, _, err = store.load(tok.Endpoint)
	require.NoError(t, err)
	assert.Equal(t, "refreshed", got.AccessToken)
	assert.Equal(t, "refresh", got.RefreshToken)
}

func TestTokenStore_MovingBackToTheFileClearsTheKeychain(t *testing.T) {
	t.Parallel()

	secrets := NewMemorySecrets()
	store := newStore(t, secrets, StoreAuto)
	fileStore := store
	fileStore.mode = func() string { return StoreFile }

	_, err := store.save(testToken())
	require.NoError(t, err)
	_, err = fileStore.save(testToken())
	require.NoError(t, err)

	assert.Empty(t, secrets.secrets)
}

func TestTokenStore_MissingKeychainEntryAsksToLogIn(t *testing.T) {
	t.Parallel()

	store := newStore(t, NewMemorySecrets(), StoreAuto)
	tok := testToken()
	_, err := store.save(tok)
	require.NoError(t, err)

	// Another machine's keychain, or an entry deleted by hand.
	store.secrets = NewMemorySecrets()
	_, _, err = store.load(tok.Endpoint)
	require.ErrorContains(t, err, "qcloud auth login")
}

func TestTokenStore_ConfigDirectoriesHaveSeparateKeychainEntries(t *testing.T) {
	t.Parallel()

	secrets := NewMemorySecrets()
	a := newStore(t, secrets, StoreAuto)
	b := newStore(t, secrets, StoreAuto)
	tok := testToken()

	_, err := a.save(tok)
	require.NoError(t, err)

	tok.AccessToken = "other"
	_, err = b.save(tok)
	require.NoError(t, err)

	got, _, err := a.load(tok.Endpoint)
	require.NoError(t, err)
	assert.Equal(t, "access", got.AccessToken)
}

func TestParseStoreMode(t *testing.T) {
	t.Parallel()

	for value, want := range map[string]string{"": StoreAuto, "auto": StoreAuto, "keychain": StoreKeychain, "file": StoreFile} {
		got, err := parseStoreMode(value)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	}

	_, err := parseStoreMode("vault")
	require.Error(t, err)
}

func TestTokenStore_DeleteReportsAKeychainThatFails(t *testing.T) {
	t.Parallel()

	store := newStore(t, NewMemorySecrets(), StoreAuto)
	tok := testToken()
	_, err := store.save(tok)
	require.NoError(t, err)

	store.secrets = failingSecrets{}
	keychainErr, err := store.delete(tok.Endpoint)
	require.NoError(t, err, "the session is removed from credentials.yaml")
	require.ErrorIs(t, keychainErr, errNoKeychain)

	_, ok, err := store.info(tok.Endpoint)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestToken_Expired(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.True(t, Token{Expiry: now.Add(10 * time.Second)}.Expired(now))
	assert.False(t, Token{Expiry: now.Add(2 * time.Minute)}.Expired(now))
	assert.False(t, Token{}.Expired(now))
}
