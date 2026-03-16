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

func TestBackupList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.ListBackupsCalls.Returns(&backupv1.ListBackupsResponse{
		Items: []*backupv1.Backup{
			{
				Id:        "backup-1",
				Name:      "my-backup",
				ClusterId: "cluster-abc",
				Status:    backupv1.BackupStatus_BACKUP_STATUS_SUCCEEDED,
				CreatedAt: timestamppb.Now(),
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "backup", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ID")
	assert.Contains(t, stdout, "NAME")
	assert.Contains(t, stdout, "CLUSTER")
	assert.Contains(t, stdout, "STATUS")
	assert.Contains(t, stdout, "CREATED")
	assert.Contains(t, stdout, "backup-1")
	assert.Contains(t, stdout, "my-backup")
	assert.Contains(t, stdout, "cluster-abc")
	assert.Contains(t, stdout, "SUCCEEDED")

	req, ok := env.BackupServer.ListBackupsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
}

func TestBackupList_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.ListBackupsCalls.Returns(&backupv1.ListBackupsResponse{
		Items: []*backupv1.Backup{
			{Id: "backup-json", ClusterId: "cluster-123"},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "backup", "list", "--json")
	require.NoError(t, err)

	var result struct {
		Items []struct {
			ID        string `json:"id"`
			ClusterID string `json:"clusterId"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Items, 1)
	assert.Equal(t, "backup-json", result.Items[0].ID)
	assert.Equal(t, "cluster-123", result.Items[0].ClusterID)
}

func TestBackupList_EmptyResponse(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.ListBackupsCalls.Returns(&backupv1.ListBackupsResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "backup", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ID")
	assert.Contains(t, stdout, "STATUS")
}

func TestBackupList_ClusterIDFilter(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.ListBackupsCalls.Returns(&backupv1.ListBackupsResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "backup", "list", "--cluster-id=my-cluster")
	require.NoError(t, err)

	req, ok := env.BackupServer.ListBackupsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "my-cluster", req.GetClusterId())
}

func TestBackupList_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.ListBackupsCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "backup", "list")
	require.Error(t, err)
}
