package backup_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestBackupDelete_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.DeleteBackupCalls.Returns(&backupv1.DeleteBackupResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "backup", "delete", "backup-abc", "--force")
	require.NoError(t, err)

	req, ok := env.BackupServer.DeleteBackupCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "backup-abc", req.GetBackupId())
	assert.Contains(t, stdout, "backup-abc")
	assert.Contains(t, stdout, "deleted")
}

func TestBackupDelete_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.DeleteBackupCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "backup", "delete", "backup-abc", "--force")
	require.Error(t, err)
}
