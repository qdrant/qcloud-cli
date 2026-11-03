package testutil

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
)

// FakeOAuthServer is the OAuth setup of one Qdrant Cloud environment: the API's protected-resource
// metadata and the login server (OpenID configuration, token, device authorization and revocation
// endpoints). With WithOAuthServer, every OAuth request of the CLI goes to it, whatever its host.
// The browser login gets the code "test-code" without a browser.
type FakeOAuthServer struct {
	// Cluster names the environment: the API is api.<Cluster>.qdrant.io, the login server
	// login.<Cluster>.qdrant.io. The default is development-cloud.
	Cluster string
	// AccessToken and RefreshToken are what a login issues.
	AccessToken  string
	RefreshToken string
	// RefreshedAccessToken is what a refresh issues.
	RefreshedAccessToken string
	// ExpiresIn is the expires_in of token responses; 0 leaves it out.
	ExpiresIn int
	// RefreshError, when set, is the error code a refresh gets.
	RefreshError string

	mu          sync.Mutex
	tokenForms  []url.Values
	deviceForms []url.Values
	revoked     []string
	srv         *httptest.Server
}

// NewFakeOAuthServer returns a FakeOAuthServer for the development environment that issues
// "access-from-login" and "refresh-from-login", valid for an hour.
func NewFakeOAuthServer() *FakeOAuthServer {
	return &FakeOAuthServer{
		Cluster:              "development-cloud",
		AccessToken:          "access-from-login",
		RefreshToken:         "refresh-from-login",
		RefreshedAccessToken: "access-from-refresh",
		ExpiresIn:            3600,
	}
}

// TokenForms returns the forms posted to the token endpoint, in order.
func (f *FakeOAuthServer) TokenForms() []url.Values {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]url.Values(nil), f.tokenForms...)
}

// DeviceForms returns the forms posted to the device authorization endpoint, in order.
func (f *FakeOAuthServer) DeviceForms() []url.Values {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]url.Values(nil), f.deviceForms...)
}

// Revoked returns the refresh tokens that were revoked, in order.
func (f *FakeOAuthServer) Revoked() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.revoked...)
}

// FakeJWT returns an unsigned JWT with claims, shaped like the access tokens of the login server.
func FakeJWT(t *testing.T, claims map[string]any) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}

	enc := base64.RawURLEncoding.EncodeToString
	return enc([]byte(`{"alg":"RS256","typ":"JWT"}`)) + "." + enc(payload) + ".signature"
}

func (f *FakeOAuthServer) start(t *testing.T) *http.Client {
	t.Helper()
	login := "https://login." + f.Cluster + ".qdrant.io"

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/oauth-protected-resource", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"resource":              "https://api." + f.Cluster + ".qdrant.io",
			"authorization_servers": []string{login + "/"},
			"scopes_supported":      []string{"read-only", "manage"},
		})
	})
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"issuer":                        login + "/",
			"authorization_endpoint":        login + "/authorize",
			"token_endpoint":                login + "/oauth/token",
			"device_authorization_endpoint": login + "/oauth/device/code",
			"revocation_endpoint":           login + "/oauth/revoke",
		})
	})
	mux.HandleFunc("/oauth/device/code", func(w http.ResponseWriter, r *http.Request) {
		form := f.record(r, &f.deviceForms)
		writeJSON(w, http.StatusOK, map[string]any{
			"device_code":               "device-" + form.Get("client_id"),
			"user_code":                 "ABCD-EFGH",
			"verification_uri":          login + "/activate",
			"verification_uri_complete": login + "/activate?user_code=ABCD-EFGH",
			"expires_in":                600,
			"interval":                  1,
		})
	})
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		form := f.record(r, &f.tokenForms)
		access := f.AccessToken
		switch form.Get("grant_type") {
		case "authorization_code":
			if form.Get("code") != "test-code" {
				writeJSON(w, http.StatusForbidden, map[string]any{"error": "invalid_grant"})
				return
			}
		case "refresh_token":
			if f.RefreshError != "" {
				writeJSON(w, http.StatusForbidden, map[string]any{"error": f.RefreshError, "error_description": "Unknown or invalid refresh token."})
				return
			}

			access = f.RefreshedAccessToken
		}

		resp := map[string]any{"access_token": access, "refresh_token": f.RefreshToken, "token_type": "Bearer"}
		if f.ExpiresIn != 0 {
			resp["expires_in"] = f.ExpiresIn
		}

		writeJSON(w, http.StatusOK, resp)
	})
	mux.HandleFunc("/oauth/revoke", func(w http.ResponseWriter, r *http.Request) {
		form := f.record(r, nil)
		f.mu.Lock()
		f.revoked = append(f.revoked, form.Get("token"))
		f.mu.Unlock()
		w.WriteHeader(http.StatusOK)
	})

	f.srv = httptest.NewServer(mux)
	t.Cleanup(f.srv.Close)

	target, _ := url.Parse(f.srv.URL)
	return &http.Client{Transport: redirectTransport{target: target}}
}

func (f *FakeOAuthServer) record(r *http.Request, into *[]url.Values) url.Values {
	_ = r.ParseForm()
	if into != nil {
		f.mu.Lock()
		*into = append(*into, r.PostForm)
		f.mu.Unlock()
	}

	return r.PostForm
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// redirectTransport sends every request to target, whatever its host.
type redirectTransport struct {
	target *url.URL
}

func (r redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = r.target.Scheme
	clone.URL.Host = r.target.Host
	clone.Host = r.target.Host
	return http.DefaultTransport.RoundTrip(clone)
}

// noNetwork fails every request, so that tests without a FakeOAuthServer never reach the network.
type noNetwork struct{}

func (noNetwork) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, io.ErrUnexpectedEOF
}
