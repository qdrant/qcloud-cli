package backup_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestBackupCreate_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.EXPECT().CreateBackup(mock.Anything, mock.MatchedBy(func(req *backupv1.CreateBackupRequest) bool {
		assert.Equal(t, "test-account-id", req.GetBackup().GetAccountId())
		assert.Equal(t, "cluster-abc", req.GetBackup().GetClusterId())
		return true
	})).Return(&backupv1.CreateBackupResponse{
		Backup: &backupv1.Backup{Id: "backup-new", ClusterId: "cluster-abc"},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "backup", "create", "--cluster-id=cluster-abc", "--retention-days=7")
	require.NoError(t, err)
	assert.Contains(t, stdout, "backup-new")
	assert.Contains(t, stdout, "cluster-abc")
}

func TestBackupCreate_WithRetention(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.EXPECT().CreateBackup(mock.Anything, mock.MatchedBy(func(req *backupv1.CreateBackupRequest) bool {
		var days int64
		if req.GetBackup().GetRetentionPeriod() != nil {
			days = int64(req.GetBackup().GetRetentionPeriod().AsDuration().Hours()) / 24
		}
		assert.Equal(t, int64(7), days)
		return true
	})).Return(&backupv1.CreateBackupResponse{
		Backup: &backupv1.Backup{Id: "backup-ret", ClusterId: "cluster-abc"},
	}, nil)

	_, _, err := testutil.Exec(t, env, "backup", "create", "--cluster-id=cluster-abc", "--retention-days=7")
	require.NoError(t, err)
}

func TestBackupCreate_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.EXPECT().CreateBackup(mock.Anything, mock.Anything).Return(&backupv1.CreateBackupResponse{
		Backup: &backupv1.Backup{Id: "backup-json", ClusterId: "cluster-123"},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "backup", "create", "--cluster-id=cluster-123", "--retention-days=7", "--json")
	require.NoError(t, err)

	var result struct {
		ID        string `json:"id"`
		ClusterID string `json:"clusterId"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "backup-json", result.ID)
}

func TestBackupCreate_InvalidRetention(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "backup", "create", "--cluster-id=cluster-abc", "--retention-days=0")
	require.Error(t, err)
}

func TestBackupCreate_MissingClusterID(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "backup", "create")
	require.Error(t, err)
}

func TestBackupCreate_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.EXPECT().CreateBackup(mock.Anything, mock.Anything).Return(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "backup", "create", "--cluster-id=cluster-abc", "--retention-days=7")
	require.Error(t, err)
}
