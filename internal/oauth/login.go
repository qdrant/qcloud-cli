package oauth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	defaultScope = "manage"
	offlineScope = "offline_access"
)

// requiredScopes are requested on every login next to read-only or manage. offline_access gives a
// refresh token. openid and email make Auth0 put email_verified in the access token, which
// Qdrant Cloud requires: Auth0 drops email_verified when an Action sets it, as it is a reserved
// OpenID Connect claim name, so only the email scope brings it in.
var requiredScopes = []string{"openid", "email", offlineScope}

// LoopbackPorts are the ports the browser login listens on, tried in order. Auth0 matches a
// loopback callback's port exactly (it does not apply the any-port rule of RFC 8252 §7.3), so
// each one is registered on the CLI's Auth0 application as http://127.0.0.1:<port>/callback.
var LoopbackPorts = []int{48271, 48272, 48273}

// WaitCodeFunc waits for an authorization code at redirectURL (loopback).
type WaitCodeFunc func(ctx context.Context, redirectURL, state string) (code string, err error)

// LoginOptions configure authorization-code or device login.
type LoginOptions struct {
	Endpoint    string
	ClientID    string
	Scope       string
	Device      bool
	LoginURL    string
	Resource    string
	Now         func() time.Time
	WaitCode    WaitCodeFunc
	RedirectURL string
	OpenURL     func(string) error
	Output      io.Writer
	Listen      func() (net.Listener, error)
}

// Login runs the OAuth flow and returns a token ready to store.
func (d Discoverer) Login(ctx context.Context, opts LoginOptions) (Token, Discovery, error) {
	if opts.Now == nil {
		opts.Now = time.Now
	}

	if opts.Scope == "" {
		opts.Scope = defaultScope
	}

	disc, err := d.Discover(ctx, opts.Endpoint)
	if err != nil {
		return Token{}, Discovery{}, err
	}

	if opts.LoginURL != "" {
		disc.Issuer = canonicalizeIssuer(opts.LoginURL)
		disc.Source = "flag --login-url"
	}

	if opts.Resource != "" {
		disc.Resource = strings.TrimRight(opts.Resource, "/")
	}

	if opts.ClientID == "" {
		id, ok := DefaultClientID(disc.Issuer)
		if !ok {
			return Token{}, disc, fmt.Errorf("no built-in OAuth client id for %s yet: pass --client-id or QDRANT_CLOUD_OAUTH_CLIENT_ID", disc.Issuer)
		}

		opts.ClientID = id
	}

	as, err := d.fetchASMetadata(ctx, disc.Issuer)
	if err != nil {
		return Token{}, disc, err
	}

	scope := joinScopes(opts.Scope)

	if opts.Device {
		tok, err := d.deviceLogin(ctx, opts, disc, as, scope)
		return tok, disc, err
	}

	tok, err := d.codeLogin(ctx, opts, disc, as, scope)
	return tok, disc, err
}

func joinScopes(scope string) string {
	parts := strings.Fields(scope)
	for _, required := range requiredScopes {
		if !slices.Contains(parts, required) {
			parts = append(parts, required)
		}
	}

	return strings.Join(parts, " ")
}

func (d Discoverer) codeLogin(ctx context.Context, opts LoginOptions, disc Discovery, as ASMetadata, scope string) (Token, error) {
	wait := opts.WaitCode
	redirectURL := opts.RedirectURL

	if wait == nil {
		ln, err := listenLoopback(opts.Listen)
		if err != nil {
			return Token{}, err
		}
		defer ln.Close()

		redirectURL = "http://" + ln.Addr().String() + "/callback"
		wait = loopbackWaiter(ln, disc.Issuer)
	} else if redirectURL == "" {
		redirectURL = "http://127.0.0.1:1/callback"
	}

	state, err := randomURLString(16)
	if err != nil {
		return Token{}, err
	}

	verifier, err := randomURLString(32)
	if err != nil {
		return Token{}, err
	}

	q := url.Values{
		"response_type":         {"code"},
		"client_id":             {opts.ClientID},
		"redirect_uri":          {redirectURL},
		"scope":                 {scope},
		"state":                 {state},
		"code_challenge":        {pkceChallenge(verifier)},
		"code_challenge_method": {"S256"},
		"resource":              {disc.Resource},
	}
	authorizeURL := as.AuthorizationEndpoint + "?" + q.Encode()

	openErr := errors.New("no browser opener")
	if opts.OpenURL != nil {
		openErr = opts.OpenURL(authorizeURL)
	}

	// The URL is only worth printing when the browser did not open: it is long, and the login
	// waits on the loopback callback either way.
	if opts.Output != nil {
		if openErr == nil {
			fmt.Fprintln(opts.Output, "Opening the browser to log in.")
		} else {
			fmt.Fprintf(opts.Output, "Could not open a browser (%v). Open this URL to log in, or use \"qcloud auth login --device\":\n%s\n", openErr, authorizeURL)
		}
	}

	code, err := wait(ctx, redirectURL, state)
	if err != nil {
		return Token{}, err
	}

	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURL},
		"client_id":     {opts.ClientID},
		"code_verifier": {verifier},
		"resource":      {disc.Resource},
	}

	base := Token{
		Endpoint: disc.Hosts.GRPCEndpoint,
		Issuer:   disc.Issuer,
		ClientID: opts.ClientID,
		Resource: disc.Resource,
		Scope:    opts.Scope,
	}

	return d.exchange(ctx, as.TokenEndpoint, form, opts.Now(), base)
}

func listenLoopback(listen func() (net.Listener, error)) (net.Listener, error) {
	if listen != nil {
		return listen()
	}

	return listenFirstFree(LoopbackPorts)
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

func loopbackWaiter(ln net.Listener, issuer string) WaitCodeFunc {
	return func(ctx context.Context, _, wantState string) (string, error) {
		codeCh := make(chan string, 1)
		errCh := make(chan error, 1)

		mux := http.NewServeMux()
		mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
			code, err := authorizationCallback(r.URL.Query(), wantState, issuer)
			if err != nil {
				writeCallbackPage(w, http.StatusBadRequest, err.Error())
				errCh <- err

				return
			}

			writeCallbackPage(w, http.StatusOK, "")
			codeCh <- code
		})

		srv := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
		go func() {
			if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
				errCh <- err
			}
		}()
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			_ = srv.Shutdown(shutdownCtx)
		}()

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case err := <-errCh:
			return "", err
		case code := <-codeCh:
			return code, nil
		}
	}
}

// authorizationCallback checks the loopback redirect. state is the client's
// CSRF value from /authorize, not the authentication server's consent-page
// redirect state. iss is checked when the authentication server sends it
// (RFC 9207). A missing iss is allowed so login still works before the server
// sends it.
func authorizationCallback(q url.Values, wantState, wantIssuer string) (string, error) {
	if errParam := q.Get("error"); errParam != "" {
		msg := q.Get("error_description")
		if msg == "" {
			msg = errParam
		}

		return "", fmt.Errorf("authorization: %s", msg)
	}

	if q.Get("state") != wantState {
		return "", fmt.Errorf("authorization state mismatch")
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
