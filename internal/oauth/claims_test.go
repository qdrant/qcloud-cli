package oauth

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadTokenClaims(t *testing.T) {
	t.Parallel()

	enc := base64.RawURLEncoding.EncodeToString
	// The claims of an access token for the generic Cloud API after the consent page, as the
	// development tenant issues them.
	payload := enc([]byte(`{
		"https://cloud.qdrant.io/email_verified": true,
		"https://cloud.qdrant.io/granted_scope": "read-only",
		"https://cloud.qdrant.io/grant_id": "8f6a1c2e-3b4d-4e5f-9a0b-1c2d3e4f5a6b",
		"user_email": "ada@example.com",
		"iss": "https://login.development-cloud.qdrant.io/",
		"sub": "auth0|0123456789abcdef",
		"aud": ["https://api.development-cloud.qdrant.io", "https://development-qdrant.eu.auth0.com/userinfo"],
		"iat": 1793600000,
		"exp": 1793603600,
		"scope": "openid email manage offline_access",
		"azp": "CjeXCZKv1plzIF9h3vZBGAhM6MLE68fI"
	}`))
	assert.Equal(t, tokenClaims{
		Email:        "ada@example.com",
		GrantedScope: "read-only",
		expiry:       time.Unix(1793603600, 0),
	}, readTokenClaims(enc([]byte(`{"alg":"RS256","typ":"JWT"}`))+"."+payload+".signature"))

	assert.Equal(t, tokenClaims{}, readTokenClaims("opaque-token"))
	assert.Equal(t, tokenClaims{}, readTokenClaims("h."+enc([]byte(`{}`))+".s"))
}

func TestScopeLabel(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "read only", ScopeLabel("read-only"))
	assert.Equal(t, "manage", ScopeLabel("manage"))
}
