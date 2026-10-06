package space_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	spacebackupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestSpaceBackupList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.ListBackupsCalls.Returns(&spacebackupv1.ListBackupsResponse{
		Items: []*spacebackupv1.Backup{
			{
				Id:        "backup-1",
				Name:      "my-space-20260101",
				SpaceId:   "space-abc",
				Status:    spacebackupv1.BackupStatus_BACKUP_STATUS_SUCCEEDED,
				CreatedAt: timestamppb.Now(),
				Stats:     &spacebackupv1.BackupStats{SizeBytes: new(int64(2 * 1024 * 1024 * 1024))},
			},
			{
				Id:             "backup-2",
				Name:           "products-20260101",
				SpaceId:        "space-abc",
				CollectionName: new("products"),
				Status:         spacebackupv1.BackupStatus_BACKUP_STATUS_RUNNING,
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "list")
	require.NoError(t, err)
	for _, h := range []string{"ID", "NAME", "SPACE", "COLLECTION", "STATUS", "SIZE", "CREATED"} {
		assert.Contains(t, stdout, h)
	}

	assert.Contains(t, stdout, "backup-1")
	assert.Contains(t, stdout, "my-space-20260101")
	assert.Contains(t, stdout, "space-abc")
	assert.Contains(t, stdout, "(all)")
	assert.Contains(t, stdout, "SUCCEEDED")
	assert.Contains(t, stdout, "2GiB")
	assert.Contains(t, stdout, "products")
	assert.Contains(t, stdout, "RUNNING")

	req, ok := env.ServerlessBackupServer.ListBackupsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Nil(t, req.SpaceId)
	assert.Nil(t, req.BackupScheduleId)
	assert.Nil(t, req.CollectionName)
}

func TestSpaceBackupList_Filters(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.ListBackupsCalls.Returns(&spacebackupv1.ListBackupsResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "list",
		"--space-id", "space-abc", "--schedule-id", "sched-1", "--collection", "products")
	require.NoError(t, err)

	req, ok := env.ServerlessBackupServer.ListBackupsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "space-abc", req.GetSpaceId())
	assert.Equal(t, "sched-1", req.GetBackupScheduleId())
	assert.Equal(t, "products", req.GetCollectionName())
}

func TestSpaceBackupList_AutoPaginates(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.ListBackupsCalls.
		OnCall(0, func(_ context.Context, _ *spacebackupv1.ListBackupsRequest) (*spacebackupv1.ListBackupsResponse, error) {
			return &spacebackupv1.ListBackupsResponse{
				Items:         []*spacebackupv1.Backup{{Id: "backup-page-1"}},
				NextPageToken: new("token-2"),
			}, nil
		}).
		OnCall(1, func(_ context.Context, req *spacebackupv1.ListBackupsRequest) (*spacebackupv1.ListBackupsResponse, error) {
			assert.Equal(t, "token-2", req.GetPageToken())
			return &spacebackupv1.ListBackupsResponse{
				Items: []*spacebackupv1.Backup{{Id: "backup-page-2"}},
			}, nil
		})

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "backup-page-1")
	assert.Contains(t, stdout, "backup-page-2")
	assert.Equal(t, 2, env.ServerlessBackupServer.ListBackupsCalls.Count())
}

func TestSpaceBackupList_ManualPagination(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.ListBackupsCalls.Returns(&spacebackupv1.ListBackupsResponse{
		Items:         []*spacebackupv1.Backup{{Id: "backup-1"}},
		NextPageToken: new("next-token"),
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "list",
		"--page-size", "1", "--page-token", "start", "--json")
	require.NoError(t, err)
	assert.Equal(t, 1, env.ServerlessBackupServer.ListBackupsCalls.Count())

	req, ok := env.ServerlessBackupServer.ListBackupsCalls.Last()
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
	assert.Equal(t, "backup-1", result.Items[0].ID)
	assert.Equal(t, "next-token", result.NextPageToken)
}

func TestSpaceBackupList_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.ListBackupsCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list backups")
}
