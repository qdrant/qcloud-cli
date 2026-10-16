package auth_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qdrant/qcloud-cli/internal/state"
	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func newEnv(t *testing.T) *testutil.TestEnv {
	t.Helper()
	return &testutil.TestEnv{
		State:   state.New("test"),
		Cleanup: func() {},
	}
}

func TestAuthDiscover_InfersLoginFromEndpoint(t *testing.T) {
	env := newEnv(t)
	env.State.SetHTTPClient(&http.Client{Transport: errRoundTrip{}})

	stdout, _, err := testutil.Exec(t, env,
		"auth", "discover",
		"--endpoint", "grpc.development-cloud.qdrant.io:443",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "grpc.development-cloud.qdrant.io:443")
	assert.Contains(t, stdout, "https://api.development-cloud.qdrant.io")
	assert.Contains(t, stdout, "https://login.development-cloud.qdrant.io/")
	assert.Contains(t, stdout, "inferred from endpoint")
}

func TestAuthDiscover_UsesGatewayMetadata(t *testing.T) {
	env := newEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/.well-known/oauth-protected-resource", r.URL.Path)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"resource":              "https://api.staging-cloud.qdrant.io",
			"authorization_servers": []string{"https://login.staging-cloud.qdrant.io/"},
			"scopes_supported":      []string{"read-only", "manage"},
		})
	}))
	t.Cleanup(srv.Close)
	env.State.SetHTTPClient(rewriteClient(t, srv.URL))

	stdout, _, err := testutil.Exec(t, env,
		"auth", "discover",
		"--endpoint", "grpc.staging-cloud.qdrant.io:443",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "https://login.staging-cloud.qdrant.io/")
	assert.Contains(t, stdout, "protected-resource metadata")
}

func TestAuthLogin_StoresToken(t *testing.T) {
	env := newEnv(t)
	home := t.TempDir()
	t.Setenv("HOME", home)

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/oauth-protected-resource", func(w http.ResponseWriter, _ *http.Request) {
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

		assert.Equal(t, "test-code", r.Form.Get("code"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "access-from-login",
			"refresh_token": "refresh-from-login",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"scope":         "manage",
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	env.State.SetHTTPClient(rewriteClient(t, srv.URL))
	env.State.SetBrowserOpener(func(string) error { return nil })
	env.State.SetWaitAuthCode(func(context.Context, string, string) (string, error) {
		return "test-code", nil
	}, "http://127.0.0.1:1/callback")

	stdout, _, err := testutil.Exec(t, env,
		"auth", "login",
		"--endpoint", "grpc.development-cloud.qdrant.io:443",
		"--client-id", "cli-app",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "Logged in")

	tokenOut, _, err := testutil.Exec(t, env,
		"auth", "token",
		"--endpoint", "grpc.development-cloud.qdrant.io:443",
	)
	require.NoError(t, err)
	assert.Equal(t, "access-from-login\n", tokenOut)

	statusOut, _, err := testutil.Exec(t, env,
		"auth", "status",
		"--endpoint", "grpc.development-cloud.qdrant.io:443",
	)
	require.NoError(t, err)
	assert.Contains(t, statusOut, "logged in")

	_, _, err = testutil.Exec(t, env,
		"auth", "logout",
		"--endpoint", "grpc.development-cloud.qdrant.io:443",
	)
	require.NoError(t, err)

	_, _, err = testutil.Exec(t, env,
		"auth", "token",
		"--endpoint", "grpc.development-cloud.qdrant.io:443",
	)
	require.Error(t, err)
}

func TestAuthLogin_MissingClientID(t *testing.T) {
	env := newEnv(t)
	t.Setenv("QDRANT_CLOUD_OAUTH_CLIENT_ID", "")
	env.State.SetHTTPClient(&http.Client{Transport: errRoundTrip{}})

	_, _, err := testutil.Exec(t, env,
		"auth", "login",
		"--endpoint", "grpc.cloud.qdrant.io:443",
		"--client-id", "",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "client id")
}

type errRoundTrip struct{}

func (errRoundTrip) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, io.ErrUnexpectedEOF
}

func rewriteClient(t *testing.T, target string) *http.Client {
	t.Helper()
	u, err := url.Parse(target)
	require.NoError(t, err)
	return &http.Client{Transport: rewriteTransport{target: u}}
}

type rewriteTransport struct {
	target *url.URL
}

func (r rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = r.target.Scheme
	clone.URL.Host = r.target.Host
	clone.Host = r.target.Host
	return http.DefaultTransport.RoundTrip(clone)
}
