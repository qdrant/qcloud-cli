package backup_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestBackupDelete_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	var capturedBackupID string
	env.BackupServer.DeleteBackupFunc = func(_ context.Context, req *backupv1.DeleteBackupRequest) (*backupv1.DeleteBackupResponse, error) {
		assert.Equal(t, "test-account-id", req.GetAccountId())
		capturedBackupID = req.GetBackupId()
		return &backupv1.DeleteBackupResponse{}, nil
	}

	stdout, _, err := testutil.Exec(t, env, "backup", "delete", "backup-abc", "--force")
	require.NoError(t, err)
	assert.Equal(t, "backup-abc", capturedBackupID)
	assert.Contains(t, stdout, "backup-abc")
	assert.Contains(t, stdout, "deleted")
}

func TestBackupDelete_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.BackupServer.DeleteBackupFunc = func(_ context.Context, _ *backupv1.DeleteBackupRequest) (*backupv1.DeleteBackupResponse, error) {
		return nil, assert.AnError
	}

	_, _, err := testutil.Exec(t, env, "backup", "delete", "backup-abc", "--force")
	require.Error(t, err)
}
