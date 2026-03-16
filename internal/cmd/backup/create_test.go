package backup_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestBackupCreate_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.BackupServer.CreateBackupFunc = func(_ context.Context, req *backupv1.CreateBackupRequest) (*backupv1.CreateBackupResponse, error) {
		assert.Equal(t, "test-account-id", req.GetBackup().GetAccountId())
		assert.Equal(t, "cluster-abc", req.GetBackup().GetClusterId())
		return &backupv1.CreateBackupResponse{
			Backup: &backupv1.Backup{Id: "backup-new", ClusterId: "cluster-abc"},
		}, nil
	}

	stdout, _, err := testutil.Exec(t, env, "backup", "create", "--cluster-id=cluster-abc", "--retention-days=7")
	require.NoError(t, err)
	assert.Contains(t, stdout, "backup-new")
	assert.Contains(t, stdout, "cluster-abc")
}

func TestBackupCreate_WithRetention(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	var capturedRetention int64
	env.BackupServer.CreateBackupFunc = func(_ context.Context, req *backupv1.CreateBackupRequest) (*backupv1.CreateBackupResponse, error) {
		if req.GetBackup().GetRetentionPeriod() != nil {
			capturedRetention = int64(req.GetBackup().GetRetentionPeriod().AsDuration().Hours()) / 24
		}
		return &backupv1.CreateBackupResponse{
			Backup: &backupv1.Backup{Id: "backup-ret", ClusterId: "cluster-abc"},
		}, nil
	}

	_, _, err := testutil.Exec(t, env, "backup", "create", "--cluster-id=cluster-abc", "--retention-days=7")
	require.NoError(t, err)
	assert.Equal(t, int64(7), capturedRetention)
}

func TestBackupCreate_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.BackupServer.CreateBackupFunc = func(_ context.Context, _ *backupv1.CreateBackupRequest) (*backupv1.CreateBackupResponse, error) {
		return &backupv1.CreateBackupResponse{
			Backup: &backupv1.Backup{Id: "backup-json", ClusterId: "cluster-123"},
		}, nil
	}

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
	t.Cleanup(env.Cleanup)

	_, _, err := testutil.Exec(t, env, "backup", "create", "--cluster-id=cluster-abc", "--retention-days=0")
	require.Error(t, err)
}

func TestBackupCreate_MissingClusterID(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	_, _, err := testutil.Exec(t, env, "backup", "create")
	require.Error(t, err)
}

func TestBackupCreate_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.BackupServer.CreateBackupFunc = func(_ context.Context, _ *backupv1.CreateBackupRequest) (*backupv1.CreateBackupResponse, error) {
		return nil, assert.AnError
	}

	_, _, err := testutil.Exec(t, env, "backup", "create", "--cluster-id=cluster-abc", "--retention-days=7")
	require.Error(t, err)
}
