package oauth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

// httpTimeout bounds every request to the gateway's metadata and to the login server.
const httpTimeout = 10 * time.Second

// ErrNotLoggedIn means there is no stored session for the endpoint.
var ErrNotLoggedIn = errors.New("not logged in")

// WaitCodeFunc waits for the authorization code of a browser login, sent to redirectURL with state.
type WaitCodeFunc func(ctx context.Context, redirectURL, state string) (code string, err error)

// Options configure a Client. Only ConfigDir is required.
type Options struct {
	// ConfigDir returns the directory of the config file; credentials.yaml goes next to it.
	ConfigDir func() string
	// TokenStore returns the token store mode: StoreAuto (the default), StoreKeychain, or StoreFile.
	TokenStore func() string
	// Secrets keeps the access and refresh tokens. The default is the OS keychain.
	Secrets SecretStore
	// HTTPClient is used for every OAuth request. The default times out after 10 seconds.
	HTTPClient *http.Client
	// WaitCode replaces the loopback listener of the browser login, with the redirect URL it sends.
	// For tests.
	WaitCode    WaitCodeFunc
	RedirectURL string
}

// Client logs in to Qdrant Cloud with OAuth and keeps the session of each endpoint: discovery of the
// login server, the browser and device logins, refresh, and logout.
type Client struct {
	http        *http.Client
	store       tokenStore
	waitCode    WaitCodeFunc
	redirectURL string
}

// NewClient returns a Client for opts.
func NewClient(opts Options) *Client {
	c := &Client{
		http:        opts.HTTPClient,
		waitCode:    opts.WaitCode,
		redirectURL: opts.RedirectURL,
		store: tokenStore{
			dir:     opts.ConfigDir,
			mode:    opts.TokenStore,
			secrets: opts.Secrets,
		},
	}
	if c.http == nil {
		c.http = &http.Client{Timeout: httpTimeout}
	}

	if c.store.mode == nil {
		c.store.mode = func() string { return StoreAuto }
	}

	if c.store.secrets == nil {
		c.store.secrets = OSKeychain{}
	}

	return c
}

// Session returns the stored session for endpoint without its access and refresh tokens (it does not
// read the OS keychain), or false when there is none.
func (c *Client) Session(endpoint string) (Token, bool, error) {
	return c.store.info(endpoint)
}

// withHTTP makes x/oauth2 use the client's HTTP client.
func (c *Client) withHTTP(ctx context.Context) context.Context {
	return context.WithValue(ctx, oauth2.HTTPClient, c.http)
}

// config is the x/oauth2 configuration of the CLI's public client at the login server.
func (c *Client) config(as asMetadata, clientID string, scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID: clientID,
		Scopes:   scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:       as.AuthorizationEndpoint,
			DeviceAuthURL: as.DeviceAuthorizationEndpoint,
			TokenURL:      as.TokenEndpoint,
			// A public client: the client id goes in the form, there is no secret.
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}
