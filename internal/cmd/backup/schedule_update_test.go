package backup_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestScheduleUpdate_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.GetBackupScheduleCalls.Returns(
		&backupv1.GetBackupScheduleResponse{
			BackupSchedule: &backupv1.BackupSchedule{
				Id:        "schedule-1",
				ClusterId: "cluster-abc",
				Schedule:  "0 2 * * *",
			},
		},
		nil,
	)
	env.BackupServer.UpdateBackupScheduleCalls.Returns(
		&backupv1.UpdateBackupScheduleResponse{
			BackupSchedule: &backupv1.BackupSchedule{Id: "schedule-1", Schedule: "0 3 * * *"},
		},
		nil,
	)

	stdout, _, err := testutil.Exec(t, env, "backup", "schedule", "update", "schedule-1",
		"--cluster-id=cluster-abc", "--schedule=0 3 * * *")
	require.NoError(t, err)
	getReq, _ := env.BackupServer.GetBackupScheduleCalls.Last()
	assert.Equal(t, "cluster-abc", getReq.GetClusterId())
	assert.Equal(t, "schedule-1", getReq.GetBackupScheduleId())
	updateReq, _ := env.BackupServer.UpdateBackupScheduleCalls.Last()
	assert.Equal(t, "schedule-1", updateReq.GetBackupSchedule().GetId())
	assert.Equal(t, "0 3 * * *", updateReq.GetBackupSchedule().GetSchedule())
	assert.Contains(t, stdout, "schedule-1")
	assert.Contains(t, stdout, "updated")
}

func TestScheduleUpdate_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.GetBackupScheduleCalls.Returns(
		&backupv1.GetBackupScheduleResponse{
			BackupSchedule: &backupv1.BackupSchedule{Id: "schedule-1", Schedule: "0 2 * * *"},
		},
		nil,
	)
	env.BackupServer.UpdateBackupScheduleCalls.Returns(
		&backupv1.UpdateBackupScheduleResponse{
			BackupSchedule: &backupv1.BackupSchedule{Id: "schedule-1", Schedule: "0 4 * * *"},
		},
		nil,
	)

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

	env.BackupServer.GetBackupScheduleCalls.Returns(
		&backupv1.GetBackupScheduleResponse{
			BackupSchedule: &backupv1.BackupSchedule{Id: "schedule-1", Schedule: "0 2 * * *"},
		},
		nil,
	)

	_, _, err := testutil.Exec(t, env, "backup", "schedule", "update", "schedule-1",
		"--cluster-id=cluster-abc", "--retention-days=0")
	require.Error(t, err)
}

func TestScheduleUpdate_WithRetention(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.GetBackupScheduleCalls.Returns(
		&backupv1.GetBackupScheduleResponse{
			BackupSchedule: &backupv1.BackupSchedule{Id: "schedule-1", Schedule: "0 2 * * *"},
		},
		nil,
	)
	env.BackupServer.UpdateBackupScheduleCalls.Returns(
		&backupv1.UpdateBackupScheduleResponse{
			BackupSchedule: &backupv1.BackupSchedule{Id: "schedule-1"},
		},
		nil,
	)

	_, _, err := testutil.Exec(t, env, "backup", "schedule", "update", "schedule-1",
		"--cluster-id=cluster-abc", "--retention-days=14")
	require.NoError(t, err)
	req, _ := env.BackupServer.UpdateBackupScheduleCalls.Last()
	var retentionDays int64
	if req.GetBackupSchedule().GetRetentionPeriod() != nil {
		retentionDays = int64(req.GetBackupSchedule().GetRetentionPeriod().AsDuration().Hours()) / 24
	}
	assert.Equal(t, int64(14), retentionDays)
}

func TestScheduleUpdate_MissingClusterID(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "backup", "schedule", "update", "schedule-1")
	require.Error(t, err)
}

func TestScheduleUpdate_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.GetBackupScheduleCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "backup", "schedule", "update", "schedule-1",
		"--cluster-id=cluster-abc", "--schedule=0 3 * * *")
	require.Error(t, err)
}
