package space_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spacev1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestSpaceCreateFromBackup_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.CreateSpaceFromBackupCalls.Returns(&spacev1.CreateSpaceFromBackupResponse{
		Space: &spacev1.Space{Id: "space-restored", Name: "my-restored-space"},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "create-from-backup",
		"--backup-id", "backup-abc",
		"--name", "my-restored-space",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "space-restored")
	assert.Contains(t, stdout, "created from backup")

	req, ok := env.ServerlessSpaceServer.CreateSpaceFromBackupCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "backup-abc", req.GetBackupId())
	assert.Equal(t, "my-restored-space", req.GetSpaceName())
	assert.Equal(t, 0, env.ServerlessSpaceServer.GetSpaceCalls.Count())
}

func TestSpaceCreateFromBackup_Wait(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.CreateSpaceFromBackupCalls.Returns(&spacev1.CreateSpaceFromBackupResponse{
		Space: &spacev1.Space{Id: "space-restored", Name: "my-restored-space"},
	}, nil)
	env.ServerlessSpaceServer.GetSpaceCalls.Always(func(_ context.Context, req *spacev1.GetSpaceRequest) (*spacev1.GetSpaceResponse, error) {
		return &spacev1.GetSpaceResponse{Space: &spacev1.Space{
			Id:   req.GetSpaceId(),
			Name: "my-restored-space",
			State: &spacev1.SpaceState{
				Phase:    spacev1.SpaceStatePhase_SPACE_STATE_PHASE_READY,
				Endpoint: &spacev1.SpaceEndpoint{Url: "https://space-restored.serverless.qdrant.io"},
			},
		}}, nil
	})

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "create-from-backup",
		"--backup-id", "backup-abc", "--name", "my-restored-space",
		"--wait", "--wait-poll-interval", "10ms",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "https://space-restored.serverless.qdrant.io")
}

func TestSpaceCreateFromBackup_MissingFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "missing backup id", args: []string{"--name", "my-restored-space"}},
		{name: "missing name", args: []string{"--backup-id", "backup-abc"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := testutil.NewTestEnv(t)

			args := append([]string{"serverless", "space", "create-from-backup"}, tt.args...)
			_, _, err := testutil.Exec(t, env, args...)
			require.Error(t, err)
			assert.Equal(t, 0, env.ServerlessSpaceServer.CreateSpaceFromBackupCalls.Count())
		})
	}
}

func TestSpaceCreateFromBackup_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.CreateSpaceFromBackupCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "create-from-backup",
		"--backup-id", "backup-abc", "--name", "my-restored-space",
	)
	require.Error(t, err)
}
