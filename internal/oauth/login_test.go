package oauth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogin_AuthorizationCode(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc(wellKnownPRM, func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"resource":              "https://api.development-cloud.qdrant.io",
			"authorization_servers": []string{"https://login.development-cloud.qdrant.io/"},
		})
	})
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                 "https://login.development-cloud.qdrant.io/",
			"authorization_endpoint": "https://login.development-cloud.qdrant.io/authorize",
			"token_endpoint":         "https://login.development-cloud.qdrant.io/oauth/token",
		})
	})
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		assert.Equal(t, "authorization_code", r.Form.Get("grant_type"))
		assert.Equal(t, "test-code", r.Form.Get("code"))
		assert.Equal(t, "cli-client", r.Form.Get("client_id"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "at-1",
			"refresh_token": "rt-1",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"scope":         "manage offline_access",
		})
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	d := Discoverer{HTTPClient: rewriteClient(t, srv.URL)}
	tok, disc, err := d.Login(context.Background(), LoginOptions{
		Endpoint:    "grpc.development-cloud.qdrant.io:443",
		ClientID:    "cli-client",
		Scope:       "manage",
		RedirectURL: "http://127.0.0.1:1/callback",
		WaitCode: func(context.Context, string, string) (string, error) {
			return "test-code", nil
		},
		OpenURL: func(string) error { return nil },
		Output:  io.Discard,
	})
	require.NoError(t, err)
	assert.Equal(t, "at-1", tok.AccessToken)
	assert.Equal(t, "rt-1", tok.RefreshToken)
	assert.Equal(t, sourceProtectedResource, disc.Source)
}

func TestLogin_RequiresClientID(t *testing.T) {
	t.Parallel()

	d := Discoverer{HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, assert.AnError
	})}}
	_, _, err := d.Login(context.Background(), LoginOptions{
		Endpoint: "grpc.cloud.qdrant.io:443",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "client id")
}
