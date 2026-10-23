package auth_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"
	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

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

// prepareLogin points env at a fake authorization server that completes a browser login.
func prepareLogin(t *testing.T, env *testutil.TestEnv) {
	t.Helper()
	prepareLoginWithToken(t, env, "access-from-login")
}

// prepareLoginWithToken is prepareLogin with the access token the server issues.
func prepareLoginWithToken(t *testing.T, env *testutil.TestEnv, accessToken string) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())

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
			"access_token":  accessToken,
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
}

const devEndpoint = "grpc.development-cloud.qdrant.io:443"

// tempConfig writes a config file for the test and returns its path; credentials.yaml goes next to it.
func tempConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	return path
}

func grantAccounts(env *testutil.TestEnv, accounts ...*accountv1.Account) {
	env.AccountServer.ListAccountsCalls.Returns(&accountv1.ListAccountsResponse{Items: accounts}, nil)
}

func TestAuthLogin_StoresToken(t *testing.T) {
	env := testutil.NewBareTestEnv(t)
	prepareLogin(t, env)
	grantAccounts(env)
	cfg := tempConfig(t, "")

	stdout, _, err := testutil.Exec(t, env, "auth", "login", "--config", cfg, "--endpoint", devEndpoint, "--client-id", "cli-app")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Logged in; no accounts were granted.")

	// The tokens are in the keychain (in-memory in tests), not in credentials.yaml.
	creds, err := os.ReadFile(filepath.Join(filepath.Dir(cfg), "credentials.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(creds), "store: keychain")
	assert.NotContains(t, string(creds), "access-from-login")

	tokenOut, _, err := testutil.Exec(t, env, "auth", "token", "--config", cfg, "--endpoint", devEndpoint)
	require.NoError(t, err)
	assert.Equal(t, "access-from-login\n", tokenOut)

	statusOut, _, err := testutil.Exec(t, env, "auth", "status", "--config", cfg, "--endpoint", devEndpoint)
	require.NoError(t, err)
	assert.Contains(t, statusOut, "logged in")

	_, _, err = testutil.Exec(t, env, "auth", "logout", "--config", cfg, "--endpoint", devEndpoint)
	require.NoError(t, err)

	_, _, err = testutil.Exec(t, env, "auth", "token", "--config", cfg, "--endpoint", devEndpoint)
	require.Error(t, err)
}

func TestAuthLogin_CreatesAndSelectsTheContextOfTheOnlyAccount(t *testing.T) {
	env := testutil.NewBareTestEnv(t)
	prepareLogin(t, env)
	grantAccounts(env, &accountv1.Account{Id: "acc-only", Name: "Only Account"})
	env.Server.ListClustersCalls.Returns(&clusterv1.ListClustersResponse{}, nil)
	cfg := tempConfig(t, "")

	stdout, _, err := testutil.Exec(t, env, "auth", "login", "--config", cfg, "--endpoint", devEndpoint, "--client-id", "cli-app")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Logged in, account Only Account (context dev-only-account).")
	assert.NotContains(t, stdout, "Switch to it")

	// The context is current: no --endpoint, no --account-id.
	_, _, err = testutil.Exec(t, env, "cluster", "list", "--config", cfg)
	require.NoError(t, err)
	req, ok := env.Server.ListClustersCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "acc-only", req.GetAccountId())

	statusOut, _, err := testutil.Exec(t, env, "auth", "status", "--config", cfg)
	require.NoError(t, err)
	assert.Contains(t, statusOut, "Context:   dev-only-account")
	assert.Contains(t, statusOut, "Account:   acc-only (from context dev-only-account)")

	listOut, _, err := testutil.Exec(t, env, "context", "list", "--config", cfg)
	require.NoError(t, err)
	assert.Regexp(t, `\*\s+dev-only-account\s+oauth`, listOut)
}

func TestAuthLogin_ShowsTheGrantedScope(t *testing.T) {
	env := testutil.NewBareTestEnv(t)
	claims := base64.RawURLEncoding.EncodeToString([]byte(
		`{"user_email":"ada@example.com","https://cloud.qdrant.io/granted_scope":"read-only"}`))
	prepareLoginWithToken(t, env, "h."+claims+".s")
	grantAccounts(env, &accountv1.Account{Id: "acc-only", Name: "Only Account"})
	cfg := tempConfig(t, "")

	stdout, _, err := testutil.Exec(t, env, "auth", "login", "--config", cfg, "--endpoint", devEndpoint, "--client-id", "cli-app")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Logged in as ada@example.com, account Only Account, read only (context dev-only-account).")

	statusOut, _, err := testutil.Exec(t, env, "auth", "status", "--config", cfg)
	require.NoError(t, err)
	assert.Contains(t, statusOut, "Access:    read only")
	assert.NotContains(t, statusOut, "Requested")
}

func TestAuthLogin_SeveralAccountsGetAContextEachAndTheFirstIsSelected(t *testing.T) {
	env := testutil.NewBareTestEnv(t)
	prepareLogin(t, env)
	grantAccounts(env,
		&accountv1.Account{Id: "a1", Name: "Jane Doe - Base Account"},
		&accountv1.Account{Id: "a2", Name: "Andrey Vasnetsov - Base Account"},
	)
	cfg := tempConfig(t, "")

	stdout, _, err := testutil.Exec(t, env, "auth", "login", "--config", cfg, "--endpoint", devEndpoint, "--client-id", "cli-app")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Contexts for your 2 accounts:")
	assert.Contains(t, stdout, "  * dev-jane-doe-base-account  (Jane Doe - Base Account)")
	assert.Contains(t, stdout, "    dev-andrey-vasnetsov-base-account  (Andrey Vasnetsov - Base Account)")
	assert.Contains(t, stdout, "Switch with: qcloud context use <name>")

	// The first account's context is current.
	env.Server.ListClustersCalls.Returns(&clusterv1.ListClustersResponse{}, nil)
	_, _, err = testutil.Exec(t, env, "cluster", "list", "--config", cfg)
	require.NoError(t, err)
	req, ok := env.Server.ListClustersCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "a1", req.GetAccountId())
}

func TestAuthLogin_UpdatesOnlyItsOwnContexts(t *testing.T) {
	env := testutil.NewBareTestEnv(t)
	prepareLogin(t, env)
	// An API key context with the name the login would pick, current; and an OAuth context of an
	// account that is no longer granted.
	cfg := tempConfig(t, `current_context: dev-only-account
contexts:
  - name: dev-only-account
    endpoint: `+devEndpoint+`
    api_key: secret
    account_id: key-acc
  - name: dev-old-account
    endpoint: `+devEndpoint+`
    account_id: old-acc
    auth: oauth
`)
	grantAccounts(env, &accountv1.Account{Id: "acc-only", Name: "Only Account"})

	// --context picks the OAuth context for this run: with the API key context active, the login would use
	// the key and look nothing up.
	stdout, stderr, err := testutil.Exec(t, env, "auth", "login", "--config", cfg, "--endpoint", devEndpoint,
		"--client-id", "cli-app", "--context", "dev-old-account")
	require.NoError(t, err)
	assert.NotContains(t, stderr, "Could not list your accounts")
	assert.Contains(t, stdout, "account Only Account (context dev-only-account-2).")
	assert.Contains(t, stdout, "Switch to it with: qcloud context use dev-only-account-2")

	listOut, _, err := testutil.Exec(t, env, "context", "list", "--config", cfg)
	require.NoError(t, err)
	assert.Regexp(t, `\*\s+dev-only-account\s+api key`, listOut, "the API key context is untouched and still current")
	assert.Regexp(t, `dev-only-account-2\s+oauth`, listOut)
	assert.NotContains(t, listOut, "dev-old-account", "the context of an account no longer granted is removed")
}

func TestAuthLogout_RemovesTheOAuthContexts(t *testing.T) {
	env := testutil.NewBareTestEnv(t)
	prepareLogin(t, env)
	grantAccounts(env, &accountv1.Account{Id: "acc-only", Name: "Only Account"})
	cfg := tempConfig(t, "")

	_, _, err := testutil.Exec(t, env, "auth", "login", "--config", cfg, "--endpoint", devEndpoint, "--client-id", "cli-app")
	require.NoError(t, err)

	stdout, _, err := testutil.Exec(t, env, "auth", "logout", "--config", cfg, "--endpoint", devEndpoint)
	require.NoError(t, err)
	assert.Contains(t, stdout, "Logged out; removed contexts dev-only-account.")

	listOut, _, err := testutil.Exec(t, env, "context", "list", "--config", cfg)
	require.NoError(t, err)
	assert.NotContains(t, listOut, "dev-only-account")
}

func TestAuthStatus_ShowsWhereTheAccountComesFrom(t *testing.T) {
	env := testutil.NewBareTestEnv(t)
	prepareLogin(t, env)
	grantAccounts(env, &accountv1.Account{Id: "acc-only", Name: "Only Account"})
	cfg := tempConfig(t, "")

	_, _, err := testutil.Exec(t, env, "auth", "login", "--config", cfg, "--endpoint", devEndpoint, "--client-id", "cli-app")
	require.NoError(t, err)

	statusOut, _, err := testutil.Exec(t, env, "auth", "status", "--config", cfg, "--account-id", "flag-acc")
	require.NoError(t, err)
	assert.Contains(t, statusOut, "Account:   flag-acc (from --account-id)")

	t.Setenv("QDRANT_CLOUD_ACCOUNT_ID", "env-acc")
	statusOut, _, err = testutil.Exec(t, env, "auth", "status", "--config", cfg)
	require.NoError(t, err)
	assert.Contains(t, statusOut, "Account:   env-acc (from QDRANT_CLOUD_ACCOUNT_ID)")
}

func TestAuthLogin_AccountLookupFailureDoesNotFailLogin(t *testing.T) {
	env := testutil.NewBareTestEnv(t)
	prepareLogin(t, env)
	env.AccountServer.ListAccountsCalls.Returns(nil, status.Error(codes.Unavailable, "down"))
	cfg := tempConfig(t, "")

	stdout, stderr, err := testutil.Exec(t, env, "auth", "login", "--config", cfg, "--endpoint", devEndpoint, "--client-id", "cli-app")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Logged in")
	assert.Contains(t, stderr, "Could not list your accounts")
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
