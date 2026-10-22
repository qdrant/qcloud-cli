package oauth

import (
	"encoding/base64"
	"encoding/json"
	"strings"
)

// TokenEmail returns the user_email claim of an access token, for display only: the signature is not
// checked here (the API does that on every call). Empty when the token carries no email.
func TokenEmail(accessToken string) string {
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		return ""
	}

	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(parts[1], "="))
	if err != nil {
		return ""
	}

	var claims struct {
		UserEmail string `json:"user_email"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return ""
	}

	return claims.UserEmail
}
