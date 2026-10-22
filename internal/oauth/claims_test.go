package oauth

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenEmail(t *testing.T) {
	t.Parallel()

	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"user_email":"ada@example.com"}`))
	assert.Equal(t, "ada@example.com", TokenEmail("h."+payload+".s"))
	assert.Empty(t, TokenEmail("opaque-token"))
	assert.Empty(t, TokenEmail("h."+base64.RawURLEncoding.EncodeToString([]byte(`{}`))+".s"))
}
