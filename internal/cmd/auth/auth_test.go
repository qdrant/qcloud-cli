package auth_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"
	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/oauth"
	"github.com/qdrant/qcloud-cli/internal/testutil"
)

const devEndpoint = "grpc.development-cloud.qdrant.io:443"

// tempConfig writes a config file for the test and returns its path; credentials.yaml goes next to it.
func tempConfig(t *testing.T, contexts map[string]map[string]string, current string) string {
	t.Helper()
	return testutil.WriteContextConfigFile(t, t.TempDir(), current, contexts)
}

func grantAccounts(env *testutil.TestEnv, accounts ...*accountv1.Account) {
	env.AccountServer.ListAccountsCalls.Returns(&accountv1.ListAccountsResponse{Items: accounts}, nil)
}

// login runs "auth login" against the development endpoint with cfg.
func login(t *testing.T, env *testutil.TestEnv, cfg string, args ...string) (string, string) {
	t.Helper()
	stdout, stderr, err := testutil.Exec(t, env, append([]string{"auth", "login", "--config", cfg, "--endpoint", devEndpoint, "--client-id", "cli-app"}, args...)...)
	require.NoError(t, err)

	return stdout, stderr
}

func TestAuthDiscover_InfersLoginFromEndpoint(t *testing.T) {
	env := testutil.NewTestEnv(t)

	stdout, _, err := testutil.Exec(t, env, "auth", "discover", "--endpoint", devEndpoint)
	require.NoError(t, err)
	assert.Contains(t, stdout, "Endpoint:     grpc.development-cloud.qdrant.io:443")
	assert.Contains(t, stdout, "API resource: https://api.development-cloud.qdrant.io")
	assert.Contains(t, stdout, "Login URL:    https://login.development-cloud.qdrant.io/")
	assert.Contains(t, stdout, "Source:       inferred from endpoint")
	assert.Contains(t, stdout, "Metadata:     not used")
}

func TestAuthDiscover_UsesGatewayMetadata(t *testing.T) {
	server := testutil.NewFakeOAuthServer()
	server.Cluster = "staging-cloud"
	env := testutil.NewTestEnv(t, testutil.WithOAuthServer(server))

	stdout, _, err := testutil.Exec(t, env, "auth", "discover", "--endpoint", "grpc.staging-cloud.qdrant.io:443")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Login URL:    https://login.staging-cloud.qdrant.io/")
	assert.Contains(t, stdout, "Source:       protected-resource metadata")
	assert.Contains(t, stdout, "Scopes:       read-only, manage")
}

func TestAuthDiscover_JSON(t *testing.T) {
	env := testutil.NewTestEnv(t)

	stdout, _, err := testutil.Exec(t, env, "auth", "discover", "--endpoint", devEndpoint, "--json")
	require.NoError(t, err)

	var got oauth.Discovery
	require.NoError(t, json.Unmarshal([]byte(stdout), &got))
	assert.Equal(t, devEndpoint, got.Endpoint)
	assert.Equal(t, "https://api.development-cloud.qdrant.io", got.Resource)
	assert.Equal(t, "https://login.development-cloud.qdrant.io/", got.Issuer)
	assert.Equal(t, "inferred from endpoint", got.Source)
	assert.NotEmpty(t, got.MetadataError)
}

func TestAuthLogin_StoresToken(t *testing.T) {
	env := testutil.NewBareTestEnv(t, testutil.WithOAuthServer(testutil.NewFakeOAuthServer()))
	grantAccounts(env)
	cfg := tempConfig(t, nil, "")

	stdout, _ := login(t, env, cfg)
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
	assert.Contains(t, statusOut, "Status:    logged in")

	_, _, err = testutil.Exec(t, env, "auth", "logout", "--config", cfg, "--endpoint", devEndpoint)
	require.NoError(t, err)

	_, _, err = testutil.Exec(t, env, "auth", "token", "--config", cfg, "--endpoint", devEndpoint)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not logged in")
}

func TestAuthLogin_Device(t *testing.T) {
	server := testutil.NewFakeOAuthServer()
	env := testutil.NewBareTestEnv(t, testutil.WithOAuthServer(server))
	grantAccounts(env)
	cfg := tempConfig(t, nil, "")

	_, stderr := login(t, env, cfg, "--device", "--scope", "read-only")
	assert.Contains(t, stderr, "confirm the code ABCD-EFGH")

	forms := server.DeviceForms()
	require.Len(t, forms, 1)
	assert.Equal(t, "https://api.development-cloud.qdrant.io", forms[0].Get("audience"))
	assert.Equal(t, "read-only openid email offline_access", forms[0].Get("scope"))
}

func TestAuthLogin_JSON(t *testing.T) {
	server := testutil.NewFakeOAuthServer()
	server.AccessToken = testutil.FakeJWT(t, map[string]any{
		"user_email":                            "ada@example.com",
		"https://cloud.qdrant.io/granted_scope": "manage",
	})
	env := testutil.NewBareTestEnv(t, testutil.WithOAuthServer(server))
	grantAccounts(env,
		&accountv1.Account{Id: "a1", Name: "First"},
		&accountv1.Account{Id: "a2", Name: "Second"},
	)
	cfg := tempConfig(t, nil, "")

	stdout, stderr := login(t, env, cfg, "--json")
	assert.Contains(t, stderr, "Opening the browser to log in.", "the prompts go to stderr")

	var got struct {
		Endpoint        string `json:"endpoint"`
		Email           string `json:"email"`
		GrantedScope    string `json:"granted_scope"`
		CurrentContext  string `json:"current_context"`
		SelectedContext string `json:"selected_context"`
		AccountsListed  bool   `json:"accounts_listed"`
		Contexts        []struct {
			Name        string `json:"name"`
			AccountID   string `json:"account_id"`
			AccountName string `json:"account_name"`
		} `json:"contexts"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &got))
	assert.Equal(t, devEndpoint, got.Endpoint)
	assert.Equal(t, "ada@example.com", got.Email)
	assert.Equal(t, "manage", got.GrantedScope)
	assert.True(t, got.AccountsListed)
	assert.Equal(t, "dev-first", got.CurrentContext)
	assert.Equal(t, "dev-first", got.SelectedContext)
	require.Len(t, got.Contexts, 2)
	assert.Equal(t, "dev-second", got.Contexts[1].Name)
	assert.Equal(t, "a2", got.Contexts[1].AccountID)
	assert.Equal(t, "Second", got.Contexts[1].AccountName)
}

func TestAuthLogin_CreatesAndSelectsTheContextOfTheOnlyAccount(t *testing.T) {
	env := testutil.NewBareTestEnv(t, testutil.WithOAuthServer(testutil.NewFakeOAuthServer()))
	grantAccounts(env, &accountv1.Account{Id: "acc-only", Name: "Only Account"})
	env.Server.ListClustersCalls.Returns(&clusterv1.ListClustersResponse{}, nil)
	cfg := tempConfig(t, nil, "")

	stdout, _ := login(t, env, cfg)
	assert.Contains(t, stdout, "Logged in, account Only Account (context dev-only-account).")
	assert.NotContains(t, stdout, "Switch to it")

	// The context is current: no --endpoint, no --account-id.
	_, _, err := testutil.Exec(t, env, "cluster", "list", "--config", cfg)
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
	server := testutil.NewFakeOAuthServer()
	server.AccessToken = testutil.FakeJWT(t, map[string]any{
		"user_email":                            "ada@example.com",
		"https://cloud.qdrant.io/granted_scope": "read-only",
		"https://cloud.qdrant.io/grant_id":      "8f6a1c2e-3b4d-4e5f-9a0b-1c2d3e4f5a6b",
		"iss":                                   "https://login.development-cloud.qdrant.io/",
		"aud":                                   []string{"https://api.development-cloud.qdrant.io"},
		"scope":                                 "openid email manage offline_access",
	})
	env := testutil.NewBareTestEnv(t, testutil.WithOAuthServer(server))
	grantAccounts(env, &accountv1.Account{Id: "acc-only", Name: "Only Account"})
	cfg := tempConfig(t, nil, "")

	stdout, _ := login(t, env, cfg)
	assert.Contains(t, stdout, "Logged in as ada@example.com, account Only Account, read only (context dev-only-account).")

	statusOut, _, err := testutil.Exec(t, env, "auth", "status", "--config", cfg)
	require.NoError(t, err)
	assert.Contains(t, statusOut, "Status:    logged in as ada@example.com")
	assert.Contains(t, statusOut, "Access:    read only")
}

func TestAuthLogin_SeveralAccountsGetAContextEachAndTheFirstIsSelected(t *testing.T) {
	env := testutil.NewBareTestEnv(t, testutil.WithOAuthServer(testutil.NewFakeOAuthServer()))
	grantAccounts(env,
		&accountv1.Account{Id: "a1", Name: "Jane Doe - Base Account"},
		&accountv1.Account{Id: "a2", Name: "John Roe - Base Account"},
	)
	cfg := tempConfig(t, nil, "")

	stdout, _ := login(t, env, cfg)
	assert.Contains(t, stdout, "Contexts for your 2 accounts:")
	assert.Contains(t, stdout, "  * dev-jane-doe-base-account  (Jane Doe - Base Account)")
	assert.Contains(t, stdout, "    dev-john-roe-base-account  (John Roe - Base Account)")
	assert.Contains(t, stdout, "Switch with: qcloud context use <name>")

	// The first account's context is current.
	env.Server.ListClustersCalls.Returns(&clusterv1.ListClustersResponse{}, nil)
	_, _, err := testutil.Exec(t, env, "cluster", "list", "--config", cfg)
	require.NoError(t, err)
	req, ok := env.Server.ListClustersCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "a1", req.GetAccountId())
}

func TestAuthLogin_UpdatesOnlyItsOwnContexts(t *testing.T) {
	env := testutil.NewBareTestEnv(t, testutil.WithOAuthServer(testutil.NewFakeOAuthServer()))
	// A current API key context with the name the login would pick, and an OAuth context of an account
	// that is no longer granted.
	cfg := tempConfig(t, map[string]map[string]string{
		"dev-only-account": {"endpoint": devEndpoint, "api_key": "secret", "account_id": "key-acc"},
		"dev-old-account":  {"endpoint": devEndpoint, "account_id": "old-acc", "auth": "oauth"},
	}, "dev-only-account")
	grantAccounts(env, &accountv1.Account{Id: "acc-only", Name: "Only Account"})

	// The API key of the current context does not stop the login from creating its contexts.
	stdout, stderr := login(t, env, cfg)
	assert.NotContains(t, stderr, "could not list your accounts")
	assert.NotContains(t, stderr, "API key is set")
	assert.Contains(t, stdout, "account Only Account (context dev-only-account-2).")
	assert.Contains(t, stdout, "Switch to it with: qcloud context use dev-only-account-2")

	listOut, _, err := testutil.Exec(t, env, "context", "list", "--config", cfg)
	require.NoError(t, err)
	assert.Regexp(t, `\*\s+dev-only-account\s+api key`, listOut, "the API key context is untouched and still current")
	assert.Regexp(t, `dev-only-account-2\s+oauth`, listOut)
	assert.NotContains(t, listOut, "dev-old-account", "the context of an account no longer granted is removed")
}

func TestAuthLogin_WarnsAboutAnAPIKeyFromTheEnvironment(t *testing.T) {
	env := testutil.NewBareTestEnv(t, testutil.WithOAuthServer(testutil.NewFakeOAuthServer()))
	grantAccounts(env, &accountv1.Account{Id: "acc-only", Name: "Only Account"})
	cfg := tempConfig(t, nil, "")
	t.Setenv("QDRANT_CLOUD_API_KEY", "env-key")

	stdout, stderr := login(t, env, cfg)
	assert.Contains(t, stderr, "Warning: an API key is set with --api-key or QDRANT_CLOUD_API_KEY")
	assert.Contains(t, stdout, "(context dev-only-account)", "the contexts are created anyway")
}

func TestAuthLogin_AccountLookupFailureDoesNotFailLogin(t *testing.T) {
	env := testutil.NewBareTestEnv(t, testutil.WithOAuthServer(testutil.NewFakeOAuthServer()))
	env.AccountServer.ListAccountsCalls.Returns(nil, status.Error(codes.Unavailable, "down"))
	cfg := tempConfig(t, nil, "")

	stdout, stderr := login(t, env, cfg)
	assert.Contains(t, stdout, "Logged in.")
	assert.Contains(t, stderr, "Warning: could not list your accounts")
}

func TestAuthLogin_MissingClientID(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAPIKey(""), testutil.WithOAuthServer(testutil.NewFakeOAuthServer()))
	t.Setenv("QDRANT_CLOUD_OAUTH_CLIENT_ID", "")

	// A login server without a built-in client id.
	_, _, err := testutil.Exec(t, env, "auth", "login", "--endpoint", "grpc.example.com:443")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no built-in OAuth client id for https://login.example.com/")
}

func TestAuthLogin_ClientIDFromTheEnvironment(t *testing.T) {
	server := testutil.NewFakeOAuthServer()
	env := testutil.NewBareTestEnv(t, testutil.WithOAuthServer(server))
	grantAccounts(env)
	cfg := tempConfig(t, nil, "")
	t.Setenv("QDRANT_CLOUD_OAUTH_CLIENT_ID", "env-client")

	_, _, err := testutil.Exec(t, env, "auth", "login", "--config", cfg, "--endpoint", devEndpoint)
	require.NoError(t, err)

	forms := server.TokenForms()
	require.Len(t, forms, 1)
	assert.Equal(t, "env-client", forms[0].Get("client_id"))
}

func TestAuthLogout_RevokesAndRemovesTheOAuthContexts(t *testing.T) {
	server := testutil.NewFakeOAuthServer()
	env := testutil.NewBareTestEnv(t, testutil.WithOAuthServer(server))
	grantAccounts(env, &accountv1.Account{Id: "acc-only", Name: "Only Account"})
	cfg := tempConfig(t, nil, "")
	login(t, env, cfg)

	stdout, stderr, err := testutil.Exec(t, env, "auth", "logout", "--config", cfg, "--endpoint", devEndpoint)
	require.NoError(t, err)
	assert.Empty(t, stderr)
	assert.Contains(t, stdout, "Logged out; removed contexts dev-only-account.")
	assert.Equal(t, []string{"refresh-from-login"}, server.Revoked())

	listOut, _, err := testutil.Exec(t, env, "context", "list", "--config", cfg)
	require.NoError(t, err)
	assert.NotContains(t, listOut, "dev-only-account")
}

// brokenKeychain stores the tokens of the login, then fails, as a keychain that was locked or reset.
type brokenKeychain struct {
	*oauth.MemorySecrets
	broken bool
}

var errKeychain = errors.New("keychain is locked")

func (k *brokenKeychain) Get(service, user string) (string, error) {
	if k.broken {
		return "", errKeychain
	}

	return k.MemorySecrets.Get(service, user)
}

func (k *brokenKeychain) Delete(service, user string) error {
	if k.broken {
		return errKeychain
	}

	return k.MemorySecrets.Delete(service, user)
}

func TestAuthLogout_WorksWhenTheKeychainFails(t *testing.T) {
	keychain := &brokenKeychain{MemorySecrets: oauth.NewMemorySecrets()}
	server := testutil.NewFakeOAuthServer()
	env := testutil.NewBareTestEnv(t, testutil.WithOAuthServer(server), testutil.WithSecretStore(keychain))
	grantAccounts(env, &accountv1.Account{Id: "acc-only", Name: "Only Account"})
	cfg := tempConfig(t, nil, "")
	login(t, env, cfg)

	keychain.broken = true
	stdout, stderr, err := testutil.Exec(t, env, "auth", "logout", "--config", cfg, "--endpoint", devEndpoint)
	require.NoError(t, err)
	assert.Contains(t, stderr, "Warning: could not read the refresh token, so it was not revoked")
	assert.Contains(t, stderr, "Warning: could not remove the OAuth tokens from the OS keychain")
	assert.Contains(t, stdout, "Logged out; removed contexts dev-only-account.")
	assert.Empty(t, server.Revoked())

	statusOut, _, err := testutil.Exec(t, env, "auth", "status", "--config", cfg, "--endpoint", devEndpoint)
	require.NoError(t, err)
	assert.Contains(t, statusOut, "Status:    logged out")
}

func TestAuthStatus_ShowsWhereTheAccountComesFrom(t *testing.T) {
	env := testutil.NewBareTestEnv(t, testutil.WithOAuthServer(testutil.NewFakeOAuthServer()))
	grantAccounts(env, &accountv1.Account{Id: "acc-only", Name: "Only Account"})
	cfg := tempConfig(t, nil, "")
	login(t, env, cfg)

	statusOut, _, err := testutil.Exec(t, env, "auth", "status", "--config", cfg, "--account-id", "flag-acc")
	require.NoError(t, err)
	assert.Contains(t, statusOut, "Account:   flag-acc (from --account-id)")

	t.Setenv("QDRANT_CLOUD_ACCOUNT_ID", "env-acc")
	statusOut, _, err = testutil.Exec(t, env, "auth", "status", "--config", cfg)
	require.NoError(t, err)
	assert.Contains(t, statusOut, "Account:   env-acc (from QDRANT_CLOUD_ACCOUNT_ID)")
}

func TestAuthStatus_JSON(t *testing.T) {
	server := testutil.NewFakeOAuthServer()
	server.AccessToken = testutil.FakeJWT(t, map[string]any{
		"user_email":                            "ada@example.com",
		"https://cloud.qdrant.io/granted_scope": "read-only",
	})
	env := testutil.NewBareTestEnv(t, testutil.WithOAuthServer(server))
	grantAccounts(env, &accountv1.Account{Id: "acc-only", Name: "Only Account"})
	cfg := tempConfig(t, nil, "")
	login(t, env, cfg)

	stdout, _, err := testutil.Exec(t, env, "auth", "status", "--config", cfg, "--json")
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &got))
	assert.Equal(t, "dev-only-account", got["context"])
	assert.Equal(t, devEndpoint, got["endpoint"])
	assert.Equal(t, "https://login.development-cloud.qdrant.io/", got["login_url"])
	assert.Equal(t, true, got["logged_in"])
	assert.Equal(t, false, got["expired"])
	assert.Equal(t, "ada@example.com", got["email"])
	assert.Equal(t, "read-only", got["granted_scope"])
	assert.Equal(t, "acc-only", got["account_id"])
	assert.Equal(t, "context dev-only-account", got["account_source"])
	assert.Equal(t, false, got["api_key_set"])
	expiry, err := time.Parse(time.RFC3339, got["expiry"].(string))
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now().Add(time.Hour), expiry, time.Minute)
}

func TestAuthStatus_LoggedOutJSON(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAPIKey(""))

	stdout, _, err := testutil.Exec(t, env, "auth", "status", "--endpoint", devEndpoint, "--json")
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &got))
	assert.Equal(t, false, got["logged_in"])
	assert.NotContains(t, got, "expiry")
}

func TestAuthToken_JSON(t *testing.T) {
	env := testutil.NewBareTestEnv(t, testutil.WithOAuthServer(testutil.NewFakeOAuthServer()))
	grantAccounts(env)
	cfg := tempConfig(t, nil, "")
	login(t, env, cfg, "--scope", "read-only")

	stdout, _, err := testutil.Exec(t, env, "auth", "token", "--config", cfg, "--endpoint", devEndpoint, "--json")
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &got))
	assert.Equal(t, "access-from-login", got["access_token"])
	assert.Equal(t, "Bearer", got["token_type"])
	assert.Equal(t, "read-only", got["scope"])
	assert.NotEmpty(t, got["expiry"])
}

func TestAuthToken_RefreshesAnExpiredToken(t *testing.T) {
	server := testutil.NewFakeOAuthServer()
	server.ExpiresIn = 10 // within the refresh margin
	env := testutil.NewBareTestEnv(t, testutil.WithOAuthServer(server))
	grantAccounts(env)
	cfg := tempConfig(t, nil, "")
	login(t, env, cfg)

	stdout, _, err := testutil.Exec(t, env, "auth", "token", "--config", cfg, "--endpoint", devEndpoint)
	require.NoError(t, err)
	assert.Equal(t, "access-from-refresh\n", stdout)

	forms := server.TokenForms()
	require.Len(t, forms, 2)
	assert.Equal(t, "refresh_token", forms[1].Get("grant_type"))
}

func TestAuthToken_AnEndedLoginAsksToLogInAgain(t *testing.T) {
	server := testutil.NewFakeOAuthServer()
	server.ExpiresIn = 10
	server.RefreshError = "invalid_grant"
	env := testutil.NewBareTestEnv(t, testutil.WithOAuthServer(server))
	grantAccounts(env)
	cfg := tempConfig(t, nil, "")
	login(t, env, cfg)

	_, _, err := testutil.Exec(t, env, "auth", "token", "--config", cfg, "--endpoint", devEndpoint)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "qcloud auth login")

	statusOut, _, err := testutil.Exec(t, env, "auth", "status", "--config", cfg, "--endpoint", devEndpoint)
	require.NoError(t, err)
	assert.Contains(t, statusOut, "Status:    logged out")
}
