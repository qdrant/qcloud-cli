package backup_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestScheduleCreate_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.CreateBackupScheduleCalls.Returns(
		&backupv1.CreateBackupScheduleResponse{
			BackupSchedule: &backupv1.BackupSchedule{Id: "schedule-new", ClusterId: "cluster-abc"},
		},
		nil,
	)

	stdout, _, err := testutil.Exec(t, env, "backup", "schedule", "create",
		"--cluster-id=cluster-abc", "--schedule=0 2 * * *", "--retention-days=30")
	require.NoError(t, err)
	req, _ := env.BackupServer.CreateBackupScheduleCalls.Last()
	assert.Equal(t, "test-account-id", req.GetBackupSchedule().GetAccountId())
	assert.Equal(t, "cluster-abc", req.GetBackupSchedule().GetClusterId())
	assert.Equal(t, "0 2 * * *", req.GetBackupSchedule().GetSchedule())
	assert.Contains(t, stdout, "schedule-new")
	assert.Contains(t, stdout, "cluster-abc")
}

func TestScheduleCreate_WithRetention(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.CreateBackupScheduleCalls.Returns(
		&backupv1.CreateBackupScheduleResponse{
			BackupSchedule: &backupv1.BackupSchedule{Id: "schedule-ret", ClusterId: "cluster-abc"},
		},
		nil,
	)

	_, _, err := testutil.Exec(t, env, "backup", "schedule", "create",
		"--cluster-id=cluster-abc", "--schedule=0 2 * * *", "--retention-days=30")
	require.NoError(t, err)
	req, _ := env.BackupServer.CreateBackupScheduleCalls.Last()
	var retentionDays int64
	if req.GetBackupSchedule().GetRetentionPeriod() != nil {
		retentionDays = int64(req.GetBackupSchedule().GetRetentionPeriod().AsDuration().Hours()) / 24
	}
	assert.Equal(t, int64(30), retentionDays)
}

func TestScheduleCreate_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.CreateBackupScheduleCalls.Returns(
		&backupv1.CreateBackupScheduleResponse{
			BackupSchedule: &backupv1.BackupSchedule{Id: "schedule-json", Schedule: "0 5 * * *"},
		},
		nil,
	)

	stdout, _, err := testutil.Exec(t, env, "backup", "schedule", "create",
		"--cluster-id=cluster-abc", "--schedule=0 5 * * *", "--retention-days=30", "--json")
	require.NoError(t, err)

	var result struct {
		ID       string `json:"id"`
		Schedule string `json:"schedule"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "schedule-json", result.ID)
}

func TestScheduleCreate_InvalidRetention(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "backup", "schedule", "create",
		"--cluster-id=cluster-abc", "--schedule=0 2 * * *", "--retention-days=0")
	require.Error(t, err)
}

func TestScheduleCreate_MissingFlags(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "backup", "schedule", "create", "--cluster-id=cluster-abc")
	require.Error(t, err)
}

func TestScheduleCreate_MissingClusterID(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "backup", "schedule", "create",
		"--schedule=0 2 * * *", "--retention-days=30")
	require.Error(t, err)
}

func TestScheduleCreate_MissingRetentionDays(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "backup", "schedule", "create",
		"--cluster-id=cluster-abc", "--schedule=0 2 * * *")
	require.Error(t, err)
}
