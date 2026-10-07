package backup_test

import (
	"context"
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

func TestRestoreList_AutoPaginates(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.ListBackupRestoresCalls.
		OnCall(0, func(_ context.Context, _ *backupv1.ListBackupRestoresRequest) (*backupv1.ListBackupRestoresResponse, error) {
			return &backupv1.ListBackupRestoresResponse{
				Items:         []*backupv1.BackupRestore{{Id: "restore-page-1"}},
				NextPageToken: new("token-2"),
			}, nil
		}).
		OnCall(1, func(_ context.Context, req *backupv1.ListBackupRestoresRequest) (*backupv1.ListBackupRestoresResponse, error) {
			assert.Equal(t, "token-2", req.GetPageToken())
			return &backupv1.ListBackupRestoresResponse{
				Items: []*backupv1.BackupRestore{{Id: "restore-page-2"}},
			}, nil
		})

	stdout, _, err := testutil.Exec(t, env, "backup", "restore", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "restore-page-1")
	assert.Contains(t, stdout, "restore-page-2")
	assert.Equal(t, 2, env.BackupServer.ListBackupRestoresCalls.Count())
}

func TestRestoreList_ManualPagination(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.ListBackupRestoresCalls.Returns(&backupv1.ListBackupRestoresResponse{
		Items:         []*backupv1.BackupRestore{{Id: "restore-1"}},
		NextPageToken: new("next-token"),
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "backup", "restore", "list",
		"--page-size", "1", "--page-token", "start", "--json")
	require.NoError(t, err)
	assert.Equal(t, 1, env.BackupServer.ListBackupRestoresCalls.Count())

	req, ok := env.BackupServer.ListBackupRestoresCalls.Last()
	require.True(t, ok)
	assert.Equal(t, int32(1), req.GetPageSize())
	assert.Equal(t, "start", req.GetPageToken())

	var result struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
		NextPageToken string `json:"nextPageToken"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Items, 1)
	assert.Equal(t, "restore-1", result.Items[0].ID)
	assert.Equal(t, "next-token", result.NextPageToken)
}
