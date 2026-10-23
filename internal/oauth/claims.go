package oauth

import (
	"encoding/base64"
	"encoding/json"
	"strings"
)

// grantedScopeClaim is set by the Qdrant post-login Action from the consent the user gave.
const grantedScopeClaim = "https://cloud.qdrant.io/granted_scope"

// TokenClaims are the access token claims qcloud shows to the user.
type TokenClaims struct {
	Email        string
	GrantedScope string
}

// ReadTokenClaims reads display claims from an access token. The signature is not checked here (the API
// does that on every call), so the values are for display only. Missing claims are empty.
func ReadTokenClaims(accessToken string) TokenClaims {
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		return TokenClaims{}
	}

	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(parts[1], "="))
	if err != nil {
		return TokenClaims{}
	}

	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return TokenClaims{}
	}

	email, _ := raw["user_email"].(string)
	granted, _ := raw[grantedScopeClaim].(string)

	return TokenClaims{Email: email, GrantedScope: granted}
}

// ScopeLabel is how a scope reads in a sentence: "read only" or "manage".
func ScopeLabel(scope string) string {
	if scope == "read-only" {
		return "read only"
	}

	return scope
}
