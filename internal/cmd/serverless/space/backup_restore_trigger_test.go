package space_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spacebackupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestSpaceBackupRestoreTrigger_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.RestoreBackupCalls.Returns(&spacebackupv1.RestoreBackupResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "restore", "trigger", "backup-1", "--force")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Restore of backup backup-1 started.")
	assert.Contains(t, stdout, "qcloud serverless space backup restore list")

	req, ok := env.ServerlessBackupServer.RestoreBackupCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "backup-1", req.GetBackupId())
}

func TestSpaceBackupRestoreTrigger_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.RestoreBackupCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "restore", "trigger", "backup-1", "--force")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to restore backup")
}

func TestSpaceBackupRestoreTrigger_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "restore", "trigger", "--force")
	require.Error(t, err)
	assert.Equal(t, 0, env.ServerlessBackupServer.RestoreBackupCalls.Count())
}
