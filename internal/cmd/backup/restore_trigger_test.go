package backup_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestRestoreTrigger_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	var capturedBackupID string
	env.BackupServer.RestoreBackupFunc = func(_ context.Context, req *backupv1.RestoreBackupRequest) (*backupv1.RestoreBackupResponse, error) {
		assert.Equal(t, "test-account-id", req.GetAccountId())
		capturedBackupID = req.GetBackupId()
		return &backupv1.RestoreBackupResponse{}, nil
	}

	stdout, _, err := testutil.Exec(t, env, "backup", "restore", "trigger", "backup-abc", "--force")
	require.NoError(t, err)
	assert.Equal(t, "backup-abc", capturedBackupID)
	assert.Contains(t, stdout, "Restore of backup backup-abc started.")
	assert.Contains(t, stdout, "Run 'qcloud backup restore list' to track progress.")
}

func TestRestoreTrigger_Aborted(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.BackupServer.RestoreBackupFunc = func(_ context.Context, _ *backupv1.RestoreBackupRequest) (*backupv1.RestoreBackupResponse, error) {
		panic("RestoreBackup must not be called when aborted")
	}

	stdout, _, err := testutil.Exec(t, env, "backup", "restore", "trigger", "backup-abc")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Aborted.")
}

func TestRestoreTrigger_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.BackupServer.RestoreBackupFunc = func(_ context.Context, _ *backupv1.RestoreBackupRequest) (*backupv1.RestoreBackupResponse, error) {
		return nil, assert.AnError
	}

	_, _, err := testutil.Exec(t, env, "backup", "restore", "trigger", "backup-abc", "--force")
	require.Error(t, err)
}
