package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

const (
	defaultScope = "manage"
	offlineScope = "offline_access"
)

// requiredScopes are requested on every login next to read-only or manage. offline_access gives a
// refresh token. openid and email make the login server put email_verified in the access token, which
// Qdrant Cloud requires: it drops email_verified when an Action sets it, as it is a reserved
// OpenID Connect claim name, so only the email scope brings it in.
var requiredScopes = []string{"openid", "email", offlineScope}

// loopbackPorts are the ports the browser login listens on, tried in order. The login server matches
// a loopback callback's port exactly (it does not apply the any-port rule of RFC 8252 §7.3), so each
// one is registered on the CLI's application as http://127.0.0.1:<port>/callback.
var loopbackPorts = []int{48271, 48272, 48273}

// LoginOptions configure a browser or device login.
type LoginOptions struct {
	Endpoint string
	// ClientID overrides the built-in client id of the login server.
	ClientID string
	// Scope is read-only or manage (the default).
	Scope  string
	Device bool
	// LoginURL overrides the discovered login server.
	LoginURL string
	// OpenURL opens the login page in a browser.
	OpenURL func(string) error
	// Output gets the prompts: the login URL, or the device code.
	Output io.Writer
}

// LoginResult is a completed login.
type LoginResult struct {
	Token     Token
	Discovery Discovery
	// CredentialsPath is the credentials.yaml the session was written to.
	CredentialsPath string
	// KeychainErr is why the tokens went to credentials.yaml although the keychain was to be tried.
	KeychainErr error
}

// Login runs the browser (authorization code + PKCE) or device login and stores the session.
func (c *Client) Login(ctx context.Context, opts LoginOptions) (LoginResult, error) {
	if opts.Scope == "" {
		opts.Scope = defaultScope
	}

	if opts.Output == nil {
		opts.Output = io.Discard
	}

	disc, err := c.Discover(ctx, opts.Endpoint)
	if err != nil {
		return LoginResult{}, err
	}

	if opts.LoginURL != "" {
		disc.Issuer = canonicalizeIssuer(opts.LoginURL)
		disc.Source = sourceLoginURLFlag
	}

	if opts.ClientID == "" {
		id, ok := defaultClientID(disc.Issuer)
		if !ok {
			return LoginResult{}, fmt.Errorf("no built-in OAuth client id for %s yet: pass --client-id or QDRANT_CLOUD_OAUTH_CLIENT_ID", disc.Issuer)
		}

		opts.ClientID = id
	}

	as, err := c.fetchASMetadata(ctx, disc.Issuer)
	if err != nil {
		return LoginResult{}, err
	}

	cfg := c.config(as, opts.ClientID, joinScopes(opts.Scope))
	ctx = c.withHTTP(ctx)

	var t *oauth2.Token
	if opts.Device {
		t, err = c.deviceLogin(ctx, cfg, opts, disc)
	} else {
		t, err = c.codeLogin(ctx, cfg, opts, disc)
	}

	if err != nil {
		return LoginResult{}, err
	}

	tok := Token{
		Endpoint: disc.Endpoint,
		Issuer:   disc.Issuer,
		ClientID: opts.ClientID,
		Resource: disc.Resource,
		Scope:    opts.Scope,
	}.with(t)

	keychainErr, err := c.store.save(tok)
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{Token: tok, Discovery: disc, CredentialsPath: c.store.path(), KeychainErr: keychainErr}, nil
}

// with returns tok updated from a token response: the access token, the refresh token when the server
// rotated it, the expiry, and the claims qcloud shows. Without expires_in, the expiry is the access
// token's exp claim, so the token still gets refreshed.
func (tok Token) with(t *oauth2.Token) Token {
	tok.AccessToken = t.AccessToken
	if t.RefreshToken != "" {
		tok.RefreshToken = t.RefreshToken
	}

	tok.TokenType = t.Type()
	claims := readTokenClaims(t.AccessToken)
	tok.Expiry = t.Expiry
	if tok.Expiry.IsZero() {
		tok.Expiry = claims.expiry
	}

	if claims.GrantedScope != "" {
		tok.GrantedScope = claims.GrantedScope
	}

	if claims.Email != "" {
		tok.Email = claims.Email
	}

	return tok
}

func joinScopes(scope string) []string {
	parts := strings.Fields(scope)
	for _, required := range requiredScopes {
		if !slices.Contains(parts, required) {
			parts = append(parts, required)
		}
	}

	return parts
}

func (c *Client) codeLogin(ctx context.Context, cfg *oauth2.Config, opts LoginOptions, disc Discovery) (*oauth2.Token, error) {
	wait, redirectURL := c.waitCode, c.redirectURL
	if wait == nil {
		ln, err := listenFirstFree(loopbackPorts)
		if err != nil {
			return nil, err
		}
		defer ln.Close()

		redirectURL = "http://" + ln.Addr().String() + "/callback"
		wait = loopbackWaiter(ln, disc.Issuer)
	}

	cfg.RedirectURL = redirectURL
	state, err := randomURLString(16)
	if err != nil {
		return nil, err
	}

	verifier := oauth2.GenerateVerifier()
	resource := oauth2.SetAuthURLParam("resource", disc.Resource)
	authorizeURL := cfg.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier), resource)

	openErr := errors.New("no browser opener")
	if opts.OpenURL != nil {
		openErr = opts.OpenURL(authorizeURL)
	}

	// The URL is only worth printing when the browser did not open: it is long, and the login
	// waits on the loopback callback either way.
	if openErr == nil {
		fmt.Fprintln(opts.Output, "Opening the browser to log in.")
	} else {
		fmt.Fprintf(opts.Output, "Could not open a browser (%v). Open this URL to log in, or use \"qcloud auth login --device\":\n%s\n", openErr, authorizeURL)
	}

	code, err := wait(ctx, redirectURL, state)
	if err != nil {
		return nil, err
	}

	t, err := cfg.Exchange(ctx, code, oauth2.VerifierOption(verifier), resource)
	if err != nil {
		return nil, tokenError(err)
	}

	return t, nil
}

func (c *Client) deviceLogin(ctx context.Context, cfg *oauth2.Config, opts LoginOptions, disc Discovery) (*oauth2.Token, error) {
	if cfg.Endpoint.DeviceAuthURL == "" {
		return nil, fmt.Errorf("authorization server does not advertise device_authorization_endpoint")
	}

	// The login server takes RFC 8707 resource on /authorize and refresh, but audience on the device
	// authorization endpoint. The value is the same API identifier discovery resolved.
	da, err := cfg.DeviceAuth(ctx, oauth2.SetAuthURLParam("audience", disc.Resource))
	if err != nil {
		return nil, fmt.Errorf("device authorization: %w", tokenError(err))
	}

	verify := da.VerificationURIComplete
	if verify == "" {
		verify = da.VerificationURI
		fmt.Fprintf(opts.Output, "Open %s and enter the code %s.\n", verify, da.UserCode)
	} else {
		fmt.Fprintf(opts.Output, "Open %s and confirm the code %s.\n", verify, da.UserCode)
	}

	if opts.OpenURL != nil && verify != "" {
		_ = opts.OpenURL(verify)
	}

	// x/oauth2 polls at the server's interval, slows down on slow_down, keeps going on
	// authorization_pending, and stops on anything else (access_denied from a declined consent,
	// expired_token) or when the code expires.
	t, err := cfg.DeviceAccessToken(ctx, da)
	if err != nil {
		return nil, tokenError(err)
	}

	return t, nil
}

// tokenError makes an OAuth error response readable: the server's description, or its error code.
func tokenError(err error) error {
	re, ok := errors.AsType[*oauth2.RetrieveError](err)
	if !ok {
		return err
	}

	msg := re.ErrorDescription
	if msg == "" {
		msg = re.ErrorCode
	}

	if msg == "" {
		return err
	}

	return fmt.Errorf("token endpoint: %s: %w", msg, err)
}

// listenFirstFree listens on the first of ports that is free on 127.0.0.1.
func listenFirstFree(ports []int) (net.Listener, error) {
	var errs []error
	for _, port := range ports {
		ln, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
		if err == nil {
			return ln, nil
		}

		errs = append(errs, err)
	}

	return nil, fmt.Errorf("listen on loopback: none of ports %v is free (or use --device): %w", ports, errors.Join(errs...))
}

type callbackResult struct {
	code string
	err  error
}

// loopbackWaiter serves the browser redirect on ln. Only a callback with the login's state ends the
// wait: a request with another state (a prefetch, an old tab) gets 400 and the wait goes on.
func loopbackWaiter(ln net.Listener, issuer string) WaitCodeFunc {
	return func(ctx context.Context, _, wantState string) (string, error) {
		results := make(chan callbackResult, 1)
		var once sync.Once
		finish := func(r callbackResult) {
			once.Do(func() { results <- r })
		}

		mux := http.NewServeMux()
		mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if q.Get("state") != wantState {
				writeCallbackPage(w, http.StatusBadRequest, "This page is not from the current login; finish the login in the window qcloud opened.")

				return
			}

			code, err := authorizationCallback(q, issuer)
			if err != nil {
				writeCallbackPage(w, http.StatusBadRequest, err.Error())
				finish(callbackResult{err: err})

				return
			}

			writeCallbackPage(w, http.StatusOK, "")
			finish(callbackResult{code: code})
		})

		srv := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
		go func() {
			if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
				finish(callbackResult{err: err})
			}
		}()
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
			defer cancel()
			_ = srv.Shutdown(shutdownCtx)
		}()

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case r := <-results:
			return r.code, r.err
		}
	}
}

// authorizationCallback reads the loopback redirect of the login (its state already checked): the
// error the login server sent, or the code. iss is checked when the server sends it (RFC 9207); a
// missing iss is allowed so login still works before the server sends it.
func authorizationCallback(q url.Values, wantIssuer string) (string, error) {
	if errParam := q.Get("error"); errParam != "" {
		msg := q.Get("error_description")
		if msg == "" {
			msg = errParam
		}

		return "", fmt.Errorf("authorization: %s", msg)
	}

	if iss := q.Get("iss"); iss != "" && canonicalizeIssuer(iss) != canonicalizeIssuer(wantIssuer) {
		return "", fmt.Errorf("authorization issuer mismatch")
	}

	code := q.Get("code")
	if code == "" {
		return "", fmt.Errorf("authorization response missing code")
	}

	return code, nil
}

func randomURLString(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}
