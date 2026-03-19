package backup_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestScheduleDelete_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.EXPECT().DeleteBackupSchedule(mock.Anything, mock.MatchedBy(func(req *backupv1.DeleteBackupScheduleRequest) bool {
		assert.Equal(t, "test-account-id", req.GetAccountId())
		assert.Equal(t, "schedule-abc", req.GetBackupScheduleId())
		assert.False(t, req.GetDeleteBackups())
		return true
	})).Return(&backupv1.DeleteBackupScheduleResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "backup", "schedule", "delete", "schedule-abc", "--force")
	require.NoError(t, err)
	assert.Contains(t, stdout, "schedule-abc")
	assert.Contains(t, stdout, "deleted")
}

func TestScheduleDelete_WithDeleteBackups(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.EXPECT().DeleteBackupSchedule(mock.Anything, mock.MatchedBy(func(req *backupv1.DeleteBackupScheduleRequest) bool {
		assert.True(t, req.GetDeleteBackups())
		return true
	})).Return(&backupv1.DeleteBackupScheduleResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "backup", "schedule", "delete", "schedule-abc", "--force", "--delete-backups")
	require.NoError(t, err)
}

func TestScheduleDelete_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.EXPECT().DeleteBackupSchedule(mock.Anything, mock.Anything).Return(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "backup", "schedule", "delete", "schedule-abc", "--force")
	require.Error(t, err)
}
