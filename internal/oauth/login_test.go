package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

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

			var out strings.Builder
			d := Discoverer{HTTPClient: rewriteClient(t, srv.URL)}
			tok, disc, err := d.Login(context.Background(), LoginOptions{
				Endpoint:    "grpc.development-cloud.qdrant.io:443",
				ClientID:    "cli-client",
				Scope:       "manage",
				RedirectURL: "http://127.0.0.1:1/callback",
				WaitCode: func(context.Context, string, string) (string, error) {
					return "test-code", nil
				},
				OpenURL: func(string) error { return tc.openErr },
				Output:  &out,
			})
			require.NoError(t, err)
			assert.Equal(t, "at-1", tok.AccessToken)
			assert.Equal(t, "rt-1", tok.RefreshToken)
			assert.Equal(t, sourceProtectedResource, disc.Source)

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

	d := Discoverer{HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, assert.AnError
	})}}
	_, _, err := d.Login(context.Background(), LoginOptions{
		Endpoint: "grpc.cloud.qdrant.io:443",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "client id")
}

func TestAuthorizationCallback(t *testing.T) {
	t.Parallel()

	code, err := authorizationCallback(url.Values{
		"code":  {"abc"},
		"state": {"s"},
		"iss":   {"https://login.cloud.qdrant.io/"},
	}, "s", "https://login.cloud.qdrant.io/")
	require.NoError(t, err)
	assert.Equal(t, "abc", code)

	_, err = authorizationCallback(url.Values{"code": {"abc"}, "state": {"other"}}, "s", "https://login.cloud.qdrant.io/")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "state mismatch")

	_, err = authorizationCallback(url.Values{
		"code":  {"abc"},
		"state": {"s"},
		"iss":   {"https://evil.example/"},
	}, "s", "https://login.cloud.qdrant.io/")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "issuer mismatch")

	_, err = authorizationCallback(url.Values{"error": {"access_denied"}}, "s", "https://login.cloud.qdrant.io/")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "access_denied")
}

func TestNextDevicePollInterval(t *testing.T) {
	t.Parallel()

	next, again := nextDevicePollInterval(5*time.Second, assert.AnError)
	assert.False(t, again)
	assert.Equal(t, 5*time.Second, next)

	// Auth0's description does not contain the code; the poll goes by the code.
	next, again = nextDevicePollInterval(5*time.Second, &TokenError{Code: "authorization_pending", Description: "User has yet to authorize device code."})
	assert.True(t, again)
	assert.Equal(t, 5*time.Second, next)

	next, again = nextDevicePollInterval(5*time.Second, &TokenError{Code: "slow_down"})
	assert.True(t, again)
	assert.Equal(t, 10*time.Second, next)

	_, again = nextDevicePollInterval(5*time.Second, &TokenError{Code: "access_denied", Description: "authorization_pending"})
	assert.False(t, again, "a description that looks like a code does not keep polling")

	_, again = nextDevicePollInterval(5*time.Second, &TokenError{Code: "expired_token"})
	assert.False(t, again)
}

func TestLogin_DeviceUsesAudience(t *testing.T) {
	t.Parallel()

	var deviceForm, tokenForm url.Values
	polls := 0

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
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		deviceForm = r.Form
		_ = json.NewEncoder(w).Encode(map[string]any{
			"device_code":               "dev-1",
			"user_code":                 "ABCD-EFGH",
			"verification_uri_complete": "https://login.development-cloud.qdrant.io/activate?user_code=ABCD-EFGH",
			"expires_in":                600,
			"interval":                  1,
		})
	})
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		tokenForm = r.Form
		polls++
		if polls == 1 {
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "authorization_pending", "error_description": "User has yet to authorize device code."})
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "at-1",
			"refresh_token": "rt-1",
			"token_type":    "Bearer",
			"expires_in":    3600,
		})
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	d := Discoverer{HTTPClient: rewriteClient(t, srv.URL)}
	tok, _, err := d.Login(context.Background(), LoginOptions{
		Endpoint: "grpc.development-cloud.qdrant.io:443",
		ClientID: "cli-client",
		Device:   true,
		Now:      time.Now,
		Output:   io.Discard,
	})
	require.NoError(t, err)
	assert.Equal(t, "at-1", tok.AccessToken)
	assert.Equal(t, "https://api.development-cloud.qdrant.io", deviceForm.Get("audience"))
	assert.Empty(t, deviceForm.Get("resource"))
	assert.Equal(t, "dev-1", tokenForm.Get("device_code"))
	assert.Empty(t, tokenForm.Get("resource"))
	assert.Empty(t, tokenForm.Get("audience"))
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

			resp, err := http.Get("http://" + ln.Addr().String() + "/callback?" + tc.query)
			require.NoError(t, err)
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, resp.Body.Close())
			require.NoError(t, err)
			<-done

			assert.Equal(t, tc.status, resp.StatusCode)
			assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))
			assert.Contains(t, string(body), tc.want)
			assert.NotContains(t, string(body), "<b>not connected</b>")
		})
	}
}

func TestJoinScopes(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "manage openid email offline_access", joinScopes("manage"))
	assert.Equal(t, "read-only openid email offline_access", joinScopes("read-only openid"))
}
