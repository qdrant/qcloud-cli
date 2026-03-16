package backup_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestScheduleDelete_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.DeleteBackupScheduleCalls.Returns(&backupv1.DeleteBackupScheduleResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "backup", "schedule", "delete", "schedule-abc", "--force")
	require.NoError(t, err)
	req, _ := env.BackupServer.DeleteBackupScheduleCalls.Last()
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "schedule-abc", req.GetBackupScheduleId())
	assert.False(t, req.GetDeleteBackups())
	assert.Contains(t, stdout, "schedule-abc")
	assert.Contains(t, stdout, "deleted")
}

func TestScheduleDelete_WithDeleteBackups(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.DeleteBackupScheduleCalls.Returns(&backupv1.DeleteBackupScheduleResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "backup", "schedule", "delete", "schedule-abc", "--force", "--delete-backups")
	require.NoError(t, err)
	req, _ := env.BackupServer.DeleteBackupScheduleCalls.Last()
	assert.True(t, req.GetDeleteBackups())
}

func TestScheduleDelete_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.DeleteBackupScheduleCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "backup", "schedule", "delete", "schedule-abc", "--force")
	require.Error(t, err)
}
