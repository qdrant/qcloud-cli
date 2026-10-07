package space_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spacebackupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestSpaceBackupDelete_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.DeleteBackupCalls.Returns(&spacebackupv1.DeleteBackupResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "delete", "backup-1", "--force")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Backup backup-1 deleted.")

	req, ok := env.ServerlessBackupServer.DeleteBackupCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "backup-1", req.GetBackupId())
}

func TestSpaceBackupDelete_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.DeleteBackupCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "delete", "backup-1", "--force")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete backup")
}

func TestSpaceBackupDelete_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "delete", "--force")
	require.Error(t, err)
	assert.Equal(t, 0, env.ServerlessBackupServer.DeleteBackupCalls.Count())
}
