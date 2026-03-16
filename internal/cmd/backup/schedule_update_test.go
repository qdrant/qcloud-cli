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

func TestScheduleUpdate_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.BackupServer.GetBackupScheduleFunc = func(_ context.Context, req *backupv1.GetBackupScheduleRequest) (*backupv1.GetBackupScheduleResponse, error) {
		assert.Equal(t, "cluster-abc", req.GetClusterId())
		assert.Equal(t, "schedule-1", req.GetBackupScheduleId())
		return &backupv1.GetBackupScheduleResponse{
			BackupSchedule: &backupv1.BackupSchedule{
				Id:        "schedule-1",
				ClusterId: "cluster-abc",
				Schedule:  "0 2 * * *",
			},
		}, nil
	}

	env.BackupServer.UpdateBackupScheduleFunc = func(_ context.Context, req *backupv1.UpdateBackupScheduleRequest) (*backupv1.UpdateBackupScheduleResponse, error) {
		assert.Equal(t, "schedule-1", req.GetBackupSchedule().GetId())
		assert.Equal(t, "0 3 * * *", req.GetBackupSchedule().GetSchedule())
		return &backupv1.UpdateBackupScheduleResponse{
			BackupSchedule: &backupv1.BackupSchedule{Id: "schedule-1", Schedule: "0 3 * * *"},
		}, nil
	}

	stdout, _, err := testutil.Exec(t, env, "backup", "schedule", "update", "schedule-1",
		"--cluster-id=cluster-abc", "--schedule=0 3 * * *")
	require.NoError(t, err)
	assert.Contains(t, stdout, "schedule-1")
	assert.Contains(t, stdout, "updated")
}

func TestScheduleUpdate_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.BackupServer.GetBackupScheduleFunc = func(_ context.Context, _ *backupv1.GetBackupScheduleRequest) (*backupv1.GetBackupScheduleResponse, error) {
		return &backupv1.GetBackupScheduleResponse{
			BackupSchedule: &backupv1.BackupSchedule{Id: "schedule-1", Schedule: "0 2 * * *"},
		}, nil
	}

	env.BackupServer.UpdateBackupScheduleFunc = func(_ context.Context, _ *backupv1.UpdateBackupScheduleRequest) (*backupv1.UpdateBackupScheduleResponse, error) {
		return &backupv1.UpdateBackupScheduleResponse{
			BackupSchedule: &backupv1.BackupSchedule{Id: "schedule-1", Schedule: "0 4 * * *"},
		}, nil
	}

	stdout, _, err := testutil.Exec(t, env, "backup", "schedule", "update", "schedule-1",
		"--cluster-id=cluster-abc", "--schedule=0 4 * * *", "--json")
	require.NoError(t, err)

	var result struct {
		ID       string `json:"id"`
		Schedule string `json:"schedule"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "schedule-1", result.ID)
}

func TestScheduleUpdate_InvalidRetention(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.BackupServer.GetBackupScheduleFunc = func(_ context.Context, _ *backupv1.GetBackupScheduleRequest) (*backupv1.GetBackupScheduleResponse, error) {
		return &backupv1.GetBackupScheduleResponse{
			BackupSchedule: &backupv1.BackupSchedule{Id: "schedule-1", Schedule: "0 2 * * *"},
		}, nil
	}

	_, _, err := testutil.Exec(t, env, "backup", "schedule", "update", "schedule-1",
		"--cluster-id=cluster-abc", "--retention-days=0")
	require.Error(t, err)
}

func TestScheduleUpdate_MissingClusterID(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	_, _, err := testutil.Exec(t, env, "backup", "schedule", "update", "schedule-1")
	require.Error(t, err)
}

func TestScheduleUpdate_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.BackupServer.GetBackupScheduleFunc = func(_ context.Context, _ *backupv1.GetBackupScheduleRequest) (*backupv1.GetBackupScheduleResponse, error) {
		return nil, assert.AnError
	}

	_, _, err := testutil.Exec(t, env, "backup", "schedule", "update", "schedule-1",
		"--cluster-id=cluster-abc", "--schedule=0 3 * * *")
	require.Error(t, err)
}
