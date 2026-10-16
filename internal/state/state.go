package state

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/pkg/browser"

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

// BrowserOpener opens a URL in the user's default browser.
type BrowserOpener func(url string) error

// State holds shared dependencies for all commands.
type State struct {
	Version               string
	Config                *config.Config
	Logger                *slog.Logger
	client                *qcloudapi.Client
	unAuthenticatedClient *qcloudapi.Client
	updater               Updater
	openBrowser           BrowserOpener
	httpClient            *http.Client
	waitAuthCode          oauth.WaitCodeFunc
	oauthRedirectURL      string
}

// New creates a new State with the given version string.
func New(version string) *State {
	return &State{
		Version: version,
		Config:  config.New(),
		Logger:  slog.New(slog.DiscardHandler),
	}
}

// Client returns the gRPC client, creating it lazily on first call.
func (s *State) Client(ctx context.Context) (*qcloudapi.Client, error) {
	if s.client != nil {
		return s.client, nil
	}

	key := s.Config.APIKey()
	if key != "" {
		c, err := qcloudapi.New(ctx, s.Config.Endpoint(), key, s.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to Qdrant Cloud API: %w", err)
		}

		s.client = c
		return s.client, nil
	}

	tok, err := s.oauthAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	c, err := qcloudapi.NewWithBearer(ctx, s.Config.Endpoint(), tok, s.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Qdrant Cloud API: %w", err)
	}

	s.client = c
	return s.client, nil
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

// HTTPClient returns the HTTP client used for OAuth discovery and token calls.
func (s *State) HTTPClient() *http.Client {
	if s.httpClient != nil {
		return s.httpClient
	}

	return http.DefaultClient
}

// SetHTTPClient injects an HTTP client. For testing.
func (s *State) SetHTTPClient(c *http.Client) {
	s.httpClient = c
}

// SetWaitAuthCode injects a function that supplies the OAuth authorization code. For testing.
func (s *State) SetWaitAuthCode(fn oauth.WaitCodeFunc, redirectURL string) {
	s.waitAuthCode = fn
	s.oauthRedirectURL = redirectURL
}

// OAuthDiscoverer returns a Discoverer bound to this state's HTTP client.
func (s *State) OAuthDiscoverer() oauth.Discoverer {
	return oauth.Discoverer{HTTPClient: s.HTTPClient()}
}

// WaitAuthCode is the injected authorization-code waiter, or nil for loopback.
func (s *State) WaitAuthCode() oauth.WaitCodeFunc {
	return s.waitAuthCode
}

// OAuthRedirectURL is the injected loopback redirect used with WaitAuthCode.
func (s *State) OAuthRedirectURL() string {
	return s.oauthRedirectURL
}

func (s *State) oauthAccessToken(ctx context.Context) (string, error) {
	path := oauth.CredentialsPath(s.Config.ConfigFilePath())
	tok, ok, err := oauth.LoadToken(path, s.Config.Endpoint())
	if err != nil {
		return "", err
	}

	if !ok || tok.AccessToken == "" {
		return "", errNoAPIKey
	}

	if tok.Expired(time.Now()) {
		refreshed, err := s.OAuthDiscoverer().Refresh(ctx, tok, time.Now())
		if err != nil {
			return "", fmt.Errorf("refresh OAuth token: %w (run \"qcloud auth login\")", err)
		}

		if err := oauth.SaveToken(path, refreshed); err != nil {
			return "", err
		}

		tok = refreshed
	}

	return tok.AccessToken, nil
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
