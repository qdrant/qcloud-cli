package state

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"

	"github.com/qdrant/qcloud-cli/internal/oauth"
	"github.com/qdrant/qcloud-cli/internal/qcloudapi"
	"github.com/qdrant/qcloud-cli/internal/selfupgrade"
	"github.com/qdrant/qcloud-cli/internal/state/config"
)

var (
	errNoAPIKey    = errors.New("no API Key configured — set QDRANT_CLOUD_API_KEY, use --api-key, run \"qcloud auth login\", or run \"qcloud context set\" to save credentials")
	errNoAccountID = errors.New("no account ID configured — set QDRANT_CLOUD_ACCOUNT_ID, use --account-id, or run \"qcloud context set\" to save credentials")
)

// Updater checks for and applies CLI updates.
type Updater interface {
	DetectLatest(ctx context.Context) (*selfupgrade.ReleaseInfo, bool, error)
	UpdateSelf(ctx context.Context, currentVersion string) (*selfupgrade.ReleaseInfo, error)
}

// OAuth is the OAuth login: discovery of the login server, the browser and device logins, and the
// stored session of each endpoint, which it refreshes and removes.
type OAuth interface {
	Discover(ctx context.Context, endpoint string) (oauth.Discovery, error)
	Login(ctx context.Context, opts oauth.LoginOptions) (oauth.LoginResult, error)
	// Session returns the stored session without its tokens, or false when there is none.
	Session(endpoint string) (oauth.Token, bool, error)
	// Token returns the session with a fresh access token.
	Token(ctx context.Context, endpoint string) (oauth.Token, error)
	TokenSource(ctx context.Context, endpoint string) (oauth2.TokenSource, error)
	Logout(ctx context.Context, endpoint string) (warnings []error, err error)
}

// BrowserOpener opens a URL in the user's default browser.
type BrowserOpener func(url string) error

// State holds shared dependencies for all commands.
type State struct {
	Version               string
	Config                *config.Config
	Logger                *slog.Logger
	client                *qcloudapi.Client
	unAuthenticatedClient *qcloudapi.Client
	oauthClient           *qcloudapi.Client
	updater               Updater
	openBrowser           BrowserOpener
	oauth                 OAuth
}

// New creates a new State with the given version string.
func New(version string) *State {
	return &State{
		Version: version,
		Config:  config.New(),
		Logger:  slog.New(slog.DiscardHandler),
	}
}

// Client returns the gRPC client, creating it lazily on first call. It uses the configured API key, or
// else the OAuth login of the endpoint.
func (s *State) Client(ctx context.Context) (*qcloudapi.Client, error) {
	if s.client != nil {
		return s.client, nil
	}

	key := s.Config.APIKey()
	if key == "" {
		c, err := s.OAuthClient(ctx)
		if err != nil {
			return nil, err
		}

		s.client = c
		return s.client, nil
	}

	c, err := qcloudapi.New(ctx, s.Config.Endpoint(), key, s.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Qdrant Cloud API: %w", err)
	}

	s.client = c
	return s.client, nil
}

// HasCredentials reports whether Client can authenticate: an API key is configured, or there is an OAuth
// login for the endpoint. An expired login still counts, Client refreshes it.
func (s *State) HasCredentials() (bool, error) {
	if s.Config.APIKey() != "" {
		return true, nil
	}

	_, ok, err := s.OAuth().Session(s.Config.Endpoint())
	return ok, err
}

// OAuthClient returns the gRPC client that uses the OAuth login of the endpoint, whatever API key is
// configured, creating it lazily on first call.
func (s *State) OAuthClient(ctx context.Context) (*qcloudapi.Client, error) {
	if s.oauthClient != nil {
		return s.oauthClient, nil
	}

	ts, err := s.OAuth().TokenSource(ctx, s.Config.Endpoint())
	if errors.Is(err, oauth.ErrNotLoggedIn) {
		return nil, errNoAPIKey
	}

	if err != nil {
		return nil, err
	}

	c, err := qcloudapi.NewWithTokenSource(ctx, s.Config.Endpoint(), ts, s.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Qdrant Cloud API: %w", err)
	}

	s.oauthClient = c
	return s.oauthClient, nil
}

// UnAuthenticatedClient returns the gRPC client without authentication, creating it lazily on first call.
func (s *State) UnAuthenticatedClient(ctx context.Context) (*qcloudapi.Client, error) {
	if s.unAuthenticatedClient != nil {
		return s.unAuthenticatedClient, nil
	}

	c, err := qcloudapi.New(ctx, s.Config.Endpoint(), "", s.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Qdrant Cloud API: %w", err)
	}

	s.unAuthenticatedClient = c
	return s.unAuthenticatedClient, nil
}

// SetClient injects a pre-built client, bypassing lazy creation.
func (s *State) SetClient(c *qcloudapi.Client) {
	s.client = c
}

// SetUnAuthenticatedClient injects a pre-built unauthenticated client, bypassing lazy creation.
func (s *State) SetUnAuthenticatedClient(c *qcloudapi.Client) {
	s.unAuthenticatedClient = c
}

// SetOAuthClient injects a pre-built OAuth client, bypassing lazy creation.
func (s *State) SetOAuthClient(c *qcloudapi.Client) {
	s.oauthClient = c
}

// OpenBrowser opens the given URL in the user's default browser.
func (s *State) OpenBrowser(url string) error {
	if s.openBrowser != nil {
		return s.openBrowser(url)
	}

	return browser.OpenURL(url)
}

// SetBrowserOpener injects a browser opener, bypassing the default. For testing.
func (s *State) SetBrowserOpener(fn BrowserOpener) {
	s.openBrowser = fn
}

// OAuth returns the OAuth login, creating it lazily on first call: the OS keychain, and
// credentials.yaml next to the config file.
func (s *State) OAuth() OAuth {
	if s.oauth == nil {
		s.oauth = oauth.NewClient(oauth.Options{
			ConfigDir:  s.ConfigDir,
			TokenStore: s.Config.TokenStore,
		})
	}

	return s.oauth
}

// SetOAuth injects an OAuth implementation.
func (s *State) SetOAuth(o OAuth) {
	s.oauth = o
}

// ConfigDir is the directory of the config file, or of the default one when none was loaded.
func (s *State) ConfigDir() string {
	path := s.Config.ConfigFilePath()
	if path == "" {
		path = config.DefaultConfigPath()
	}

	return filepath.Dir(path)
}

// Updater returns the CLI updater, creating it lazily on first call.
func (s *State) Updater() (Updater, error) {
	if s.updater != nil {
		return s.updater, nil
	}

	u, err := selfupgrade.NewGitHubUpdater()
	if err != nil {
		return nil, err
	}

	s.updater = u
	return s.updater, nil
}

// SetUpdater injects an Updater implementation.
func (s *State) SetUpdater(u Updater) {
	s.updater = u
}

// AccountID returns the configured account ID or an error.
func (s *State) AccountID() (string, error) {
	id := s.Config.AccountID()
	if id == "" {
		return "", errNoAccountID
	}

	return id, nil
}
