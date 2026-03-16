package backup_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestScheduleDelete_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	var capturedScheduleID string
	var capturedDeleteBackups bool
	env.BackupServer.DeleteBackupScheduleFunc = func(_ context.Context, req *backupv1.DeleteBackupScheduleRequest) (*backupv1.DeleteBackupScheduleResponse, error) {
		assert.Equal(t, "test-account-id", req.GetAccountId())
		capturedScheduleID = req.GetBackupScheduleId()
		capturedDeleteBackups = req.GetDeleteBackups()
		return &backupv1.DeleteBackupScheduleResponse{}, nil
	}

	stdout, _, err := testutil.Exec(t, env, "backup", "schedule", "delete", "schedule-abc", "--force")
	require.NoError(t, err)
	assert.Equal(t, "schedule-abc", capturedScheduleID)
	assert.False(t, capturedDeleteBackups)
	assert.Contains(t, stdout, "schedule-abc")
	assert.Contains(t, stdout, "deleted")
}

func TestScheduleDelete_WithDeleteBackups(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	var capturedDeleteBackups bool
	env.BackupServer.DeleteBackupScheduleFunc = func(_ context.Context, req *backupv1.DeleteBackupScheduleRequest) (*backupv1.DeleteBackupScheduleResponse, error) {
		capturedDeleteBackups = req.GetDeleteBackups()
		return &backupv1.DeleteBackupScheduleResponse{}, nil
	}

	_, _, err := testutil.Exec(t, env, "backup", "schedule", "delete", "schedule-abc", "--force", "--delete-backups")
	require.NoError(t, err)
	assert.True(t, capturedDeleteBackups)
}

func TestScheduleDelete_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.BackupServer.DeleteBackupScheduleFunc = func(_ context.Context, _ *backupv1.DeleteBackupScheduleRequest) (*backupv1.DeleteBackupScheduleResponse, error) {
		return nil, assert.AnError
	}

	_, _, err := testutil.Exec(t, env, "backup", "schedule", "delete", "schedule-abc", "--force")
	require.Error(t, err)
}
