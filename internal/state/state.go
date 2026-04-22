package state

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/qdrant/qcloud-cli/internal/cmdexec"
	"github.com/qdrant/qcloud-cli/internal/qcloudapi"
	"github.com/qdrant/qcloud-cli/internal/selfupgrade"
	"github.com/qdrant/qcloud-cli/internal/state/config"
)

var (
	errNoAPIKey    = errors.New("no API Key configured — set QDRANT_CLOUD_API_KEY, use --api-key, or run \"qcloud context set\" to save credentials")
	errNoAccountID = errors.New("no account ID configured — set QDRANT_CLOUD_ACCOUNT_ID, use --account-id, or run \"qcloud context set\" to save credentials")
)

// Updater checks for and applies CLI updates.
type Updater interface {
	DetectLatest(ctx context.Context) (*selfupgrade.ReleaseInfo, bool, error)
	UpdateSelf(ctx context.Context, currentVersion string) (*selfupgrade.ReleaseInfo, error)
}

// State holds shared dependencies for all commands.
type State struct {
	Version   string
	Config    *config.Config
	Logger    *slog.Logger
	cmdRunner cmdexec.Runner
	client    *qcloudapi.Client
	updater   Updater
}

// New creates a new State with the given version string.
func New(version string) *State {
	return &State{
		Version:   version,
		Config:    config.New(),
		Logger:    slog.New(slog.DiscardHandler),
		cmdRunner: cmdexec.ExecRunner{},
	}
}

// CmdRunner returns the command runner.
func (s *State) CmdRunner() cmdexec.Runner {
	return s.cmdRunner
}

// SetCmdRunner injects a command runner implementation.
func (s *State) SetCmdRunner(r cmdexec.Runner) {
	s.cmdRunner = r
}

// Client returns the gRPC client, creating it lazily on first call.
func (s *State) Client(ctx context.Context) (*qcloudapi.Client, error) {
	if s.client != nil {
		return s.client, nil
	}

	key := s.Config.APIKey()
	if key == "" {
		return nil, errNoAPIKey
	}

	c, err := qcloudapi.New(ctx, s.Config.Endpoint(), key, s.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Qdrant Cloud API: %w", err)
	}

	s.client = c
	return s.client, nil
}

// SetClient injects a pre-built client, bypassing lazy creation.
func (s *State) SetClient(c *qcloudapi.Client) {
	s.client = c
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
