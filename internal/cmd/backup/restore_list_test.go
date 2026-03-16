package backup_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestRestoreList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.ListBackupRestoresCalls.Returns(
		&backupv1.ListBackupRestoresResponse{
			Items: []*backupv1.BackupRestore{
				{
					Id:        "restore-1",
					BackupId:  "backup-abc",
					ClusterId: "cluster-123",
					Status:    backupv1.BackupRestoreStatus_BACKUP_RESTORE_STATUS_SUCCEEDED,
					CreatedAt: timestamppb.Now(),
				},
			},
		},
		nil,
	)

	stdout, _, err := testutil.Exec(t, env, "backup", "restore", "list")
	require.NoError(t, err)
	req, _ := env.BackupServer.ListBackupRestoresCalls.Last()
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Contains(t, stdout, "ID")
	assert.Contains(t, stdout, "BACKUP")
	assert.Contains(t, stdout, "CLUSTER")
	assert.Contains(t, stdout, "STATUS")
	assert.Contains(t, stdout, "restore-1")
	assert.Contains(t, stdout, "backup-abc")
	assert.Contains(t, stdout, "SUCCEEDED")
}

func TestRestoreList_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.ListBackupRestoresCalls.Returns(
		&backupv1.ListBackupRestoresResponse{
			Items: []*backupv1.BackupRestore{
				{Id: "restore-json", BackupId: "backup-123"},
			},
		},
		nil,
	)

	stdout, _, err := testutil.Exec(t, env, "backup", "restore", "list", "--json")
	require.NoError(t, err)

	var result struct {
		Items []struct {
			ID       string `json:"id"`
			BackupID string `json:"backupId"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Items, 1)
	assert.Equal(t, "restore-json", result.Items[0].ID)
	assert.Equal(t, "backup-123", result.Items[0].BackupID)
}

func TestRestoreList_EmptyResponse(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.ListBackupRestoresCalls.Returns(&backupv1.ListBackupRestoresResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "backup", "restore", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ID")
	assert.Contains(t, stdout, "STATUS")
}

func TestRestoreList_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.ListBackupRestoresCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "backup", "restore", "list")
	require.Error(t, err)
}

func TestRestoreList_ClusterIDFilter(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.ListBackupRestoresCalls.Returns(&backupv1.ListBackupRestoresResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "backup", "restore", "list", "--cluster-id=my-cluster")
	require.NoError(t, err)
	req, _ := env.BackupServer.ListBackupRestoresCalls.Last()
	assert.Equal(t, "my-cluster", req.GetClusterId())
}
