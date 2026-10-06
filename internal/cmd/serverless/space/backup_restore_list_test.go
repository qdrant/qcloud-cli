package space_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	spacebackupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestSpaceBackupRestoreList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.ListBackupRestoresCalls.Returns(&spacebackupv1.ListBackupRestoresResponse{
		Items: []*spacebackupv1.BackupRestore{{
			Id:        "restore-1",
			BackupId:  "backup-1",
			SpaceId:   "space-abc",
			Status:    spacebackupv1.BackupRestoreStatus_BACKUP_RESTORE_STATUS_RUNNING,
			CreatedAt: timestamppb.Now(),
			Stats: &spacebackupv1.BackupRestoreStats{
				Progress: new("35.2/75.4 GiB (46%)"),
				Duration: durationpb.New(2 * time.Minute),
			},
		}},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "restore", "list")
	require.NoError(t, err)
	for _, h := range []string{"ID", "BACKUP", "SPACE", "STATUS", "PROGRESS", "DURATION", "CREATED"} {
		assert.Contains(t, stdout, h)
	}

	assert.Contains(t, stdout, "restore-1")
	assert.Contains(t, stdout, "backup-1")
	assert.Contains(t, stdout, "space-abc")
	assert.Contains(t, stdout, "RUNNING")
	assert.Contains(t, stdout, "(46%)")
	assert.Contains(t, stdout, "2m0s")

	req, ok := env.ServerlessBackupServer.ListBackupRestoresCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Nil(t, req.SpaceId)
}

func TestSpaceBackupRestoreList_SpaceFilter(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.ListBackupRestoresCalls.Returns(&spacebackupv1.ListBackupRestoresResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "restore", "list", "--space-id", "space-abc")
	require.NoError(t, err)

	req, ok := env.ServerlessBackupServer.ListBackupRestoresCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "space-abc", req.GetSpaceId())
}

func TestSpaceBackupRestoreList_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.ListBackupRestoresCalls.Returns(&spacebackupv1.ListBackupRestoresResponse{
		Items: []*spacebackupv1.BackupRestore{{Id: "restore-1", BackupId: "backup-1"}},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "restore", "list", "--json")
	require.NoError(t, err)

	var result struct {
		Items []struct {
			ID       string `json:"id"`
			BackupID string `json:"backupId"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Items, 1)
	assert.Equal(t, "restore-1", result.Items[0].ID)
	assert.Equal(t, "backup-1", result.Items[0].BackupID)
}

func TestSpaceBackupRestoreList_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.ListBackupRestoresCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "restore", "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list backup restores")
}
