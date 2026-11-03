package oauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const devEndpoint = "grpc.development-cloud.qdrant.io:443"

// fakeLoginServer is the development environment's protected-resource metadata and login server.
// token answers the token endpoint.
type fakeLoginServer struct {
	mu         sync.Mutex
	tokenForms []url.Values
	deviceForm url.Values
	token      func(form url.Values) (int, map[string]any)
}

func (f *fakeLoginServer) forms() []url.Values {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]url.Values(nil), f.tokenForms...)
}

func (f *fakeLoginServer) start(t *testing.T) *http.Client {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc(wellKnownPRM, func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"resource":              "https://api.development-cloud.qdrant.io",
			"authorization_servers": []string{"https://login.development-cloud.qdrant.io/"},
		})
	})
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                        "https://login.development-cloud.qdrant.io/",
			"authorization_endpoint":        "https://login.development-cloud.qdrant.io/authorize",
			"token_endpoint":                "https://login.development-cloud.qdrant.io/oauth/token",
			"device_authorization_endpoint": "https://login.development-cloud.qdrant.io/oauth/device/code",
		})
	})
	mux.HandleFunc("/oauth/device/code", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		f.mu.Lock()
		f.deviceForm = r.PostForm
		f.mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{
			"device_code":               "dev-1",
			"user_code":                 "ABCD-EFGH",
			"verification_uri_complete": "https://login.development-cloud.qdrant.io/activate?user_code=ABCD-EFGH",
			"expires_in":                600,
			"interval":                  1,
		})
	})
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		f.mu.Lock()
		f.tokenForms = append(f.tokenForms, r.PostForm)
		f.mu.Unlock()
		status, body := f.token(r.PostForm)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return rewriteClient(t, srv.URL)
}

// newTestClient is a Client for httpClient, with an in-memory keychain in a new config directory and
// a browser login that gets the code "test-code".
func newTestClient(t *testing.T, httpClient *http.Client) *Client {
	t.Helper()
	dir := t.TempDir()

	return NewClient(Options{
		ConfigDir:   func() string { return dir },
		Secrets:     NewMemorySecrets(),
		HTTPClient:  httpClient,
		RedirectURL: "http://127.0.0.1:48271/callback",
		WaitCode: func(context.Context, string, string) (string, error) {
			return "test-code", nil
		},
	})
}

func jwt(claims string) string {
	enc := base64.RawURLEncoding.EncodeToString
	return enc([]byte(`{"alg":"RS256"}`)) + "." + enc([]byte(claims)) + ".s"
}

func TestLogin_AuthorizationCode(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		openErr error
		wantURL bool
	}{
		{name: "browser opens: no URL", openErr: nil, wantURL: false},
		{name: "no browser: the URL, and the login still waits", openErr: errors.New("no display"), wantURL: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := &fakeLoginServer{token: func(url.Values) (int, map[string]any) {
				return http.StatusOK, map[string]any{
					"access_token":  jwt(`{"user_email":"ada@example.com","https://cloud.qdrant.io/granted_scope":"read-only"}`),
					"refresh_token": "rt-1",
					"token_type":    "Bearer",
					"expires_in":    3600,
					"scope":         "manage offline_access",
				}
			}}
			c := newTestClient(t, server.start(t))

			var out strings.Builder
			var opened string
			res, err := c.Login(context.Background(), LoginOptions{
				Endpoint: devEndpoint,
				ClientID: "cli-client",
				Scope:    "manage",
				OpenURL: func(u string) error {
					opened = u
					return tc.openErr
				},
				Output: &out,
			})
			require.NoError(t, err)
			assert.Equal(t, sourceProtectedResource, res.Discovery.Source)
			assert.Equal(t, "rt-1", res.Token.RefreshToken)
			assert.Equal(t, "manage", res.Token.Scope, "the requested scope, not the scope of the response")
			assert.Equal(t, "read-only", res.Token.GrantedScope)
			assert.Equal(t, "ada@example.com", res.Token.Email)

			authorize, err := url.Parse(opened)
			require.NoError(t, err)
			q := authorize.Query()
			assert.Equal(t, "cli-client", q.Get("client_id"))
			assert.Equal(t, "https://api.development-cloud.qdrant.io", q.Get("resource"))
			assert.Equal(t, "S256", q.Get("code_challenge_method"))
			assert.Equal(t, "manage openid email offline_access", q.Get("scope"))

			forms := server.forms()
			require.Len(t, forms, 1)
			assert.Equal(t, "authorization_code", forms[0].Get("grant_type"))
			assert.Equal(t, "test-code", forms[0].Get("code"))
			assert.Equal(t, "cli-client", forms[0].Get("client_id"))
			assert.Equal(t, "https://api.development-cloud.qdrant.io", forms[0].Get("resource"))
			assert.NotEmpty(t, forms[0].Get("code_verifier"))

			stored, ok, err := c.store.load(devEndpoint)
			require.NoError(t, err)
			require.True(t, ok)
			assert.Equal(t, res.Token.AccessToken, stored.AccessToken)

			if tc.wantURL {
				assert.Contains(t, out.String(), "https://login.development-cloud.qdrant.io/authorize?")
				assert.Contains(t, out.String(), "--device")
			} else {
				assert.Equal(t, "Opening the browser to log in.\n", out.String())
			}
		})
	}
}

func TestLogin_RequiresClientID(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, assert.AnError
	})})
	_, err := c.Login(context.Background(), LoginOptions{Endpoint: "grpc.cloud.qdrant.io:443"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "client id")
}

func TestLogin_DeviceUsesAudience(t *testing.T) {
	t.Parallel()

	polls := 0
	server := &fakeLoginServer{token: func(url.Values) (int, map[string]any) {
		polls++
		if polls == 1 {
			return http.StatusForbidden, map[string]any{"error": "authorization_pending", "error_description": "User has yet to authorize device code."}
		}

		return http.StatusOK, map[string]any{"access_token": "at-1", "refresh_token": "rt-1", "token_type": "Bearer", "expires_in": 3600}
	}}
	c := newTestClient(t, server.start(t))

	var out strings.Builder
	res, err := c.Login(context.Background(), LoginOptions{
		Endpoint: devEndpoint,
		ClientID: "cli-client",
		Device:   true,
		Output:   &out,
	})
	require.NoError(t, err)
	assert.Equal(t, "at-1", res.Token.AccessToken)
	assert.Contains(t, out.String(), "confirm the code ABCD-EFGH")

	server.mu.Lock()
	deviceForm := server.deviceForm
	server.mu.Unlock()
	assert.Equal(t, "https://api.development-cloud.qdrant.io", deviceForm.Get("audience"))
	assert.Empty(t, deviceForm.Get("resource"))

	forms := server.forms()
	require.Len(t, forms, 2, "authorization_pending keeps polling")
	assert.Equal(t, "dev-1", forms[1].Get("device_code"))
	assert.Empty(t, forms[1].Get("resource"))
	assert.Empty(t, forms[1].Get("audience"))
}

func TestLogin_DeviceStopsWhenTheConsentIsDeclined(t *testing.T) {
	t.Parallel()

	server := &fakeLoginServer{token: func(url.Values) (int, map[string]any) {
		return http.StatusForbidden, map[string]any{"error": "access_denied", "error_description": "The user declined the consent."}
	}}
	c := newTestClient(t, server.start(t))

	_, err := c.Login(context.Background(), LoginOptions{Endpoint: devEndpoint, ClientID: "cli-client", Device: true})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "The user declined the consent.")
	assert.Len(t, server.forms(), 1)
}

func TestTokenSource_RefreshesAndStoresTheToken(t *testing.T) {
	t.Parallel()

	server := &fakeLoginServer{token: func(form url.Values) (int, map[string]any) {
		if form.Get("grant_type") == "refresh_token" {
			return http.StatusOK, map[string]any{"access_token": "at-2", "refresh_token": "rt-2", "token_type": "Bearer", "expires_in": 3600}
		}

		// Expires within the refresh margin.
		return http.StatusOK, map[string]any{"access_token": "at-1", "refresh_token": "rt-1", "token_type": "Bearer", "expires_in": 10}
	}}
	c := newTestClient(t, server.start(t))
	_, err := c.Login(context.Background(), LoginOptions{Endpoint: devEndpoint, ClientID: "cli-client"})
	require.NoError(t, err)

	ts, err := c.TokenSource(context.Background(), devEndpoint)
	require.NoError(t, err)
	tok, err := ts.Token()
	require.NoError(t, err)
	assert.Equal(t, "at-2", tok.AccessToken)

	forms := server.forms()
	require.Len(t, forms, 2)
	assert.Equal(t, "refresh_token", forms[1].Get("grant_type"))
	assert.Equal(t, "rt-1", forms[1].Get("refresh_token"))
	assert.Equal(t, "cli-client", forms[1].Get("client_id"))

	// The next call reuses the refreshed token, and the store has it.
	tok, err = ts.Token()
	require.NoError(t, err)
	assert.Equal(t, "at-2", tok.AccessToken)
	assert.Len(t, server.forms(), 2)

	stored, _, err := c.store.load(devEndpoint)
	require.NoError(t, err)
	assert.Equal(t, "at-2", stored.AccessToken)
	assert.Equal(t, "rt-2", stored.RefreshToken)
}

func TestTokenSource_UsesTheTokenExpiryWithoutExpiresIn(t *testing.T) {
	t.Parallel()

	expired := jwt(fmt.Sprintf(`{"exp":%d}`, time.Now().Add(-time.Minute).Unix()))
	server := &fakeLoginServer{token: func(form url.Values) (int, map[string]any) {
		if form.Get("grant_type") == "refresh_token" {
			return http.StatusOK, map[string]any{"access_token": "at-2", "token_type": "Bearer"}
		}

		return http.StatusOK, map[string]any{"access_token": expired, "refresh_token": "rt-1", "token_type": "Bearer"}
	}}
	c := newTestClient(t, server.start(t))
	res, err := c.Login(context.Background(), LoginOptions{Endpoint: devEndpoint, ClientID: "cli-client"})
	require.NoError(t, err)
	assert.False(t, res.Token.Expiry.IsZero(), "the expiry comes from the exp claim")

	ts, err := c.TokenSource(context.Background(), devEndpoint)
	require.NoError(t, err)
	tok, err := ts.Token()
	require.NoError(t, err)
	assert.Equal(t, "at-2", tok.AccessToken)

	stored, _, err := c.store.load(devEndpoint)
	require.NoError(t, err)
	assert.Equal(t, "rt-1", stored.RefreshToken, "a refresh without a new refresh token keeps the old one")
}

func TestTokenSource_InvalidGrantEndsTheSession(t *testing.T) {
	t.Parallel()

	server := &fakeLoginServer{token: func(form url.Values) (int, map[string]any) {
		if form.Get("grant_type") == "refresh_token" {
			return http.StatusForbidden, map[string]any{"error": "invalid_grant", "error_description": "Unknown or invalid refresh token."}
		}

		return http.StatusOK, map[string]any{"access_token": "at-1", "refresh_token": "rt-1", "token_type": "Bearer", "expires_in": 10}
	}}
	c := newTestClient(t, server.start(t))
	_, err := c.Login(context.Background(), LoginOptions{Endpoint: devEndpoint, ClientID: "cli-client"})
	require.NoError(t, err)

	ts, err := c.TokenSource(context.Background(), devEndpoint)
	require.NoError(t, err)
	_, err = ts.Token()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "qcloud auth login")

	_, ok, err := c.Session(devEndpoint)
	require.NoError(t, err)
	assert.False(t, ok, "the session is removed, so the next command does not try again")

	_, err = c.TokenSource(context.Background(), devEndpoint)
	require.ErrorIs(t, err, ErrNotLoggedIn)
}

func TestAuthorizationCallback(t *testing.T) {
	t.Parallel()

	code, err := authorizationCallback(url.Values{
		"code": {"abc"},
		"iss":  {"https://login.cloud.qdrant.io/"},
	}, "https://login.cloud.qdrant.io/")
	require.NoError(t, err)
	assert.Equal(t, "abc", code)

	_, err = authorizationCallback(url.Values{
		"code": {"abc"},
		"iss":  {"https://evil.example/"},
	}, "https://login.cloud.qdrant.io/")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "issuer mismatch")

	_, err = authorizationCallback(url.Values{"error": {"access_denied"}}, "https://login.cloud.qdrant.io/")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "access_denied")
}

func TestListenFirstFree_SkipsBusyPort(t *testing.T) {
	t.Parallel()

	busy, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer busy.Close()

	free, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	freePort := free.Addr().(*net.TCPAddr).Port
	require.NoError(t, free.Close())

	ln, err := listenFirstFree([]int{busy.Addr().(*net.TCPAddr).Port, freePort})
	require.NoError(t, err)
	defer ln.Close()

	assert.Equal(t, freePort, ln.Addr().(*net.TCPAddr).Port)
}

func TestListenFirstFree_AllBusy(t *testing.T) {
	t.Parallel()

	busy, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer busy.Close()

	_, err = listenFirstFree([]int{busy.Addr().(*net.TCPAddr).Port})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--device")
}

// callback sends the browser redirect to the loopback listener and returns the status and the page.
func callback(t *testing.T, ln net.Listener, query string) (int, string) {
	t.Helper()

	resp, err := http.Get("http://" + ln.Addr().String() + "/callback?" + query)
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, resp.Body.Close())
	require.NoError(t, err)

	return resp.StatusCode, string(body)
}

func TestLoopbackWaiter_Pages(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, query, want string
		status            int
	}{
		{"success", "code=abc&state=s", "Login Complete", http.StatusOK},
		{"denied", "error=access_denied&error_description=" + url.QueryEscape("<b>not connected</b>") + "&state=s", "&lt;b&gt;not connected&lt;/b&gt;", http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ln, err := net.Listen("tcp", "127.0.0.1:0")
			require.NoError(t, err)

			wait := loopbackWaiter(ln, "https://login.cloud.qdrant.io/")
			done := make(chan struct{})
			go func() {
				defer close(done)
				_, _ = wait(context.Background(), "", "s")
			}()

			status, body := callback(t, ln, tc.query)
			<-done

			assert.Equal(t, tc.status, status)
			assert.Contains(t, body, tc.want)
			assert.NotContains(t, body, "<b>not connected</b>")
		})
	}
}

func TestLoopbackWaiter_KeepsWaitingAfterAnotherState(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	wait := loopbackWaiter(ln, "https://login.cloud.qdrant.io/")
	type result struct {
		code string
		err  error
	}
	results := make(chan result, 1)
	go func() {
		code, err := wait(context.Background(), "", "s")
		results <- result{code, err}
	}()

	// A prefetch, or the callback of an old tab, then the real one.
	status, body := callback(t, ln, "code=old&state=other")
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, body, "not from the current login")
	status, _ = callback(t, ln, "error=access_denied")
	assert.Equal(t, http.StatusBadRequest, status, "an error without the login's state does not end it either")

	status, _ = callback(t, ln, "code=abc&state=s")
	assert.Equal(t, http.StatusOK, status)

	r := <-results
	require.NoError(t, r.err)
	assert.Equal(t, "abc", r.code)
}

func TestJoinScopes(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{"manage", "openid", "email", "offline_access"}, joinScopes("manage"))
	assert.Equal(t, []string{"read-only", "openid", "email", "offline_access"}, joinScopes("read-only openid"))
}
