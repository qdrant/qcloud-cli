package oauth

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadTokenClaims(t *testing.T) {
	t.Parallel()

	payload := base64.RawURLEncoding.EncodeToString(
		[]byte(`{"user_email":"ada@example.com","https://cloud.qdrant.io/granted_scope":"read-only"}`))
	assert.Equal(t, TokenClaims{Email: "ada@example.com", GrantedScope: "read-only"}, ReadTokenClaims("h."+payload+".s"))
	assert.Equal(t, TokenClaims{}, ReadTokenClaims("opaque-token"))
	assert.Equal(t, TokenClaims{}, ReadTokenClaims("h."+base64.RawURLEncoding.EncodeToString([]byte(`{}`))+".s"))
}

func TestScopeLabel(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "read only", ScopeLabel("read-only"))
	assert.Equal(t, "manage", ScopeLabel("manage"))
}
