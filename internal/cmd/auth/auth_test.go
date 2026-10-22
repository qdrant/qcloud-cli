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

func TestAuthLogin_StoresToken(t *testing.T) {
	env := testutil.NewBareTestEnv(t)
	prepareLogin(t, env)
	env.AccountServer.ListAccountsCalls.Returns(&accountv1.ListAccountsResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "auth", "login", "--endpoint", devEndpoint, "--client-id", "cli-app")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Logged in")

	tokenOut, _, err := testutil.Exec(t, env, "auth", "token", "--endpoint", devEndpoint)
	require.NoError(t, err)
	assert.Equal(t, "access-from-login\n", tokenOut)

	statusOut, _, err := testutil.Exec(t, env, "auth", "status", "--endpoint", devEndpoint)
	require.NoError(t, err)
	assert.Contains(t, statusOut, "logged in")

	_, _, err = testutil.Exec(t, env, "auth", "logout", "--endpoint", devEndpoint)
	require.NoError(t, err)

	_, _, err = testutil.Exec(t, env, "auth", "token", "--endpoint", devEndpoint)
	require.Error(t, err)
}

func TestAuthLogin_RemembersTheOnlyGrantedAccount(t *testing.T) {
	env := testutil.NewBareTestEnv(t)
	prepareLogin(t, env)
	env.AccountServer.ListAccountsCalls.Returns(&accountv1.ListAccountsResponse{
		Items: []*accountv1.Account{{Id: "acc-only", Name: "Only Account"}},
	}, nil)
	env.Server.ListClustersCalls.Returns(&clusterv1.ListClustersResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "auth", "login", "--endpoint", devEndpoint, "--client-id", "cli-app")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Logged in, account Only Account.")

	statusOut, _, err := testutil.Exec(t, env, "auth", "status", "--endpoint", devEndpoint)
	require.NoError(t, err)
	assert.Contains(t, statusOut, "Account:   acc-only (from the login)")

	// No --account-id: the command uses the account the login was granted.
	_, _, err = testutil.Exec(t, env, "cluster", "list", "--endpoint", devEndpoint)
	require.NoError(t, err)
	req, ok := env.Server.ListClustersCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "acc-only", req.GetAccountId())
}

func TestAuthLogin_ShowsTheGrantedScope(t *testing.T) {
	env := testutil.NewBareTestEnv(t)
	claims := base64.RawURLEncoding.EncodeToString([]byte(
		`{"user_email":"ada@example.com","https://cloud.qdrant.io/granted_scope":"read-only"}`))
	prepareLoginWithToken(t, env, "h."+claims+".s")
	env.AccountServer.ListAccountsCalls.Returns(&accountv1.ListAccountsResponse{
		Items: []*accountv1.Account{{Id: "acc-only", Name: "Only Account"}},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "auth", "login", "--endpoint", devEndpoint, "--client-id", "cli-app")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Logged in as ada@example.com, account Only Account, read only.")

	statusOut, _, err := testutil.Exec(t, env, "auth", "status", "--endpoint", devEndpoint)
	require.NoError(t, err)
	assert.Contains(t, statusOut, "Granted:   read-only")
	assert.Contains(t, statusOut, "Requested: manage")
}

func TestAuthStatus_ShowsWhereTheAccountComesFrom(t *testing.T) {
	env := testutil.NewBareTestEnv(t)
	prepareLogin(t, env)
	env.AccountServer.ListAccountsCalls.Returns(&accountv1.ListAccountsResponse{
		Items: []*accountv1.Account{{Id: "acc-only", Name: "Only Account"}},
	}, nil)

	// A context that also sets an account, active through current_context.
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(`current_context: dev
contexts:
  - name: dev
    endpoint: `+devEndpoint+`
    account_id: ctx-acc
`), 0o600))

	_, _, err := testutil.Exec(t, env, "auth", "login", "--config", configPath, "--client-id", "cli-app")
	require.NoError(t, err)

	statusOut, _, err := testutil.Exec(t, env, "auth", "status", "--config", configPath)
	require.NoError(t, err)
	assert.Contains(t, statusOut, "Context:   dev")
	assert.Contains(t, statusOut, "Account:   ctx-acc (from context dev)")

	statusOut, _, err = testutil.Exec(t, env, "auth", "status", "--config", configPath, "--account-id", "flag-acc")
	require.NoError(t, err)
	assert.Contains(t, statusOut, "Account:   flag-acc (from --account-id)")

	t.Setenv("QDRANT_CLOUD_ACCOUNT_ID", "env-acc")
	statusOut, _, err = testutil.Exec(t, env, "auth", "status", "--config", configPath)
	require.NoError(t, err)
	assert.Contains(t, statusOut, "Account:   env-acc (from QDRANT_CLOUD_ACCOUNT_ID)")
}

func TestAuthLogin_SeveralGrantedAccountsAreNotChosen(t *testing.T) {
	env := testutil.NewBareTestEnv(t)
	prepareLogin(t, env)
	env.AccountServer.ListAccountsCalls.Returns(&accountv1.ListAccountsResponse{
		Items: []*accountv1.Account{{Id: "a1", Name: "One"}, {Id: "a2", Name: "Two"}},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "auth", "login", "--endpoint", devEndpoint, "--client-id", "cli-app")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Logged in; 2 accounts available, pass --account-id.")

	_, _, err = testutil.Exec(t, env, "cluster", "list", "--endpoint", devEndpoint)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no account ID configured")
}

func TestAuthLogin_AccountLookupFailureDoesNotFailLogin(t *testing.T) {
	env := testutil.NewBareTestEnv(t)
	prepareLogin(t, env)
	env.AccountServer.ListAccountsCalls.Returns(nil, status.Error(codes.Unavailable, "down"))

	stdout, stderr, err := testutil.Exec(t, env, "auth", "login", "--endpoint", devEndpoint, "--client-id", "cli-app")
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
