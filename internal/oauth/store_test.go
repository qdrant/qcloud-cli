package oauth

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenStore_RoundTrip(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "credentials.yaml")
	tok := Token{
		Endpoint:     "grpc.development-cloud.qdrant.io:443",
		AccessToken:  "access",
		RefreshToken: "refresh",
		Scope:        "manage",
		Expiry:       time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC),
	}

	require.NoError(t, SaveToken(path, tok))

	got, ok, err := LoadToken(path, tok.Endpoint)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, "access", got.AccessToken)
	assert.Equal(t, "refresh", got.RefreshToken)

	require.NoError(t, DeleteToken(path, tok.Endpoint))
	_, ok, err = LoadToken(path, tok.Endpoint)
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
