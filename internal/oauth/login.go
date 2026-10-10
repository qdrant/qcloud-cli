package oauth

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"
)

const (
	defaultScope = "manage"
	offlineScope = "offline_access"
)

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
		return Token{}, disc, fmt.Errorf("OAuth client id is required (--client-id or QDRANT_CLOUD_OAUTH_CLIENT_ID) once the Qdrant Cloud CLI OAuth app exists for this tenant")
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
	hasOffline := slices.Contains(parts, offlineScope)

	if !hasOffline {
		parts = append(parts, offlineScope)
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
		wait = loopbackWaiter(ln, opts.Output)
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

	if opts.OpenURL != nil {
		if err := opts.OpenURL(authorizeURL); err != nil {
			return Token{}, fmt.Errorf("open browser: %w", err)
		}
	}

	if opts.Output != nil {
		fmt.Fprintf(opts.Output, "Opening browser for login.\nIf it does not open, visit:\n%s\n", authorizeURL)
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

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("listen on loopback: %w", err)
	}

	return ln, nil
}

func loopbackWaiter(ln net.Listener, out io.Writer) WaitCodeFunc {
	return func(ctx context.Context, _, wantState string) (string, error) {
		codeCh := make(chan string, 1)
		errCh := make(chan error, 1)

		mux := http.NewServeMux()
		mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if errParam := q.Get("error"); errParam != "" {
				msg := q.Get("error_description")
				if msg == "" {
					msg = errParam
				}

				http.Error(w, msg, http.StatusBadRequest)
				errCh <- fmt.Errorf("authorization: %s", msg)
				return
			}

			if q.Get("state") != wantState {
				http.Error(w, "state mismatch", http.StatusBadRequest)
				errCh <- fmt.Errorf("authorization state mismatch")
				return
			}

			code := q.Get("code")
			if code == "" {
				http.Error(w, "missing code", http.StatusBadRequest)
				errCh <- fmt.Errorf("authorization response missing code")
				return
			}

			fmt.Fprint(w, "You can close this window and return to the CLI.")
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

		if out != nil {
			fmt.Fprintf(out, "Waiting for login to complete on %s\n", ln.Addr().String())
		}

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
