package space_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spacebackupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestSpaceBackupScheduleDelete_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.DeleteBackupScheduleCalls.Returns(&spacebackupv1.DeleteBackupScheduleResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "delete", "sched-1", "--force")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Backup schedule sched-1 deleted.")

	req, ok := env.ServerlessBackupServer.DeleteBackupScheduleCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "sched-1", req.GetBackupScheduleId())
	assert.Nil(t, req.DeleteBackups)
}

func TestSpaceBackupScheduleDelete_DeleteBackups(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.DeleteBackupScheduleCalls.Returns(&spacebackupv1.DeleteBackupScheduleResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "delete", "sched-1",
		"--delete-backups", "--force")
	require.NoError(t, err)

	req, ok := env.ServerlessBackupServer.DeleteBackupScheduleCalls.Last()
	require.True(t, ok)
	require.NotNil(t, req.DeleteBackups)
	assert.True(t, req.GetDeleteBackups())
}

func TestSpaceBackupScheduleDelete_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.DeleteBackupScheduleCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "delete", "sched-1", "--force")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete backup schedule")
}

func TestSpaceBackupScheduleDelete_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "delete", "--force")
	require.Error(t, err)
	assert.Equal(t, 0, env.ServerlessBackupServer.DeleteBackupScheduleCalls.Count())
}
