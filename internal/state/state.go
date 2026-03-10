package state

import (
	"context"
	"fmt"

	"github.com/qdrant/qcloud-cli/internal/qcloudapi"
	"github.com/qdrant/qcloud-cli/internal/state/config"
)

// State holds shared dependencies for all commands.
type State struct {
	Version string
	Config  *config.Config
	client  *qcloudapi.Client
}

// New creates a new State with the given version string.
func New(version string) *State {
	return &State{Version: version}
}

// Client returns the gRPC client, creating it lazily on first call.
func (s *State) Client(ctx context.Context) (*qcloudapi.Client, error) {
	if s.client != nil {
		return s.client, nil
	}

	key := s.Config.APIKey()
	if key == "" {
		return nil, fmt.Errorf("no API Key configured — set QDRANT_CLOUD_MANAGEMENT_KEY or use --api-key")
	}

	c, err := qcloudapi.New(ctx, s.Config.Endpoint(), key)
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

// AccountID returns the configured account ID or an error.
func (s *State) AccountID() (string, error) {
	id := s.Config.AccountID()
	if id == "" {
		return "", fmt.Errorf("no account ID configured — set QDRANT_CLOUD_ACCOUNT_ID or use --account-id")
	}
	return id, nil
}
