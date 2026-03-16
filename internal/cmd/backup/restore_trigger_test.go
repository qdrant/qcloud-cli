package backup_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestRestoreTrigger_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.RestoreBackupCalls.Returns(&backupv1.RestoreBackupResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "backup", "restore", "trigger", "backup-abc", "--force")
	require.NoError(t, err)
	req, _ := env.BackupServer.RestoreBackupCalls.Last()
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "backup-abc", req.GetBackupId())
	assert.Contains(t, stdout, "Restore of backup backup-abc started.")
	assert.Contains(t, stdout, "Run 'qcloud backup restore list' to track progress.")
}

func TestRestoreTrigger_Aborted(t *testing.T) {
	env := testutil.NewTestEnv(t)

	stdout, _, err := testutil.Exec(t, env, "backup", "restore", "trigger", "backup-abc")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Aborted.")
	assert.Equal(t, 0, env.BackupServer.RestoreBackupCalls.Count())
}

func TestRestoreTrigger_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.RestoreBackupCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "backup", "restore", "trigger", "backup-abc", "--force")
	require.Error(t, err)
}
