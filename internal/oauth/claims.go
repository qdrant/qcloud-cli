package oauth

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

const (
	// grantedScopeClaim is set by the Qdrant post-login Action from the consent the user gave.
	grantedScopeClaim = "https://cloud.qdrant.io/granted_scope"
	// emailClaim is set by the Qdrant post-login Action that adds the console's claims; the standard
	// email claim is only in ID tokens.
	emailClaim = "user_email"
)

// tokenClaims are the access token claims qcloud shows to the user, and its expiry.
type tokenClaims struct {
	Email        string
	GrantedScope string
	expiry       time.Time
}

// readTokenClaims reads display claims from an access token. The signature is not checked here (the API
// does that on every call), so the values are for display and refresh timing only. Missing claims are
// empty.
func readTokenClaims(accessToken string) tokenClaims {
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		return tokenClaims{}
	}

	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(parts[1], "="))
	if err != nil {
		return tokenClaims{}
	}

	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return tokenClaims{}
	}

	claims := tokenClaims{}
	claims.Email, _ = raw[emailClaim].(string)
	claims.GrantedScope, _ = raw[grantedScopeClaim].(string)
	if exp, ok := raw["exp"].(float64); ok && exp > 0 {
		claims.expiry = time.Unix(int64(exp), 0)
	}

	return claims
}

// ScopeLabel is how a scope reads in a sentence: "read only" or "manage".
func ScopeLabel(scope string) string {
	if scope == "read-only" {
		return "read only"
	}

	return scope
}
