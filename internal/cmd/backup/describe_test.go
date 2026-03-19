package backup_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestBackupDescribe_TextOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.EXPECT().GetBackup(mock.Anything, mock.MatchedBy(func(req *backupv1.GetBackupRequest) bool {
		assert.Equal(t, "test-account-id", req.GetAccountId())
		assert.Equal(t, "backup-abc", req.GetBackupId())
		return true
	})).Return(&backupv1.GetBackupResponse{
		Backup: &backupv1.Backup{
			Id:        "backup-abc",
			Name:      "my-backup",
			ClusterId: "cluster-123",
			Status:    backupv1.BackupStatus_BACKUP_STATUS_SUCCEEDED,
			CreatedAt: timestamppb.Now(),
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "backup", "describe", "backup-abc")
	require.NoError(t, err)
	assert.Contains(t, stdout, "backup-abc")
	assert.Contains(t, stdout, "my-backup")
	assert.Contains(t, stdout, "cluster-123")
	assert.Contains(t, stdout, "SUCCEEDED")
}

func TestBackupDescribe_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.EXPECT().GetBackup(mock.Anything, mock.Anything).Return(&backupv1.GetBackupResponse{
		Backup: &backupv1.Backup{Id: "backup-json", ClusterId: "cluster-xyz"},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "backup", "describe", "backup-json", "--json")
	require.NoError(t, err)

	var result struct {
		ID        string `json:"id"`
		ClusterID string `json:"clusterId"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "backup-json", result.ID)
	assert.Equal(t, "cluster-xyz", result.ClusterID)
}

func TestBackupDescribe_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.EXPECT().GetBackup(mock.Anything, mock.Anything).Return(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "backup", "describe", "backup-abc")
	require.Error(t, err)
}
