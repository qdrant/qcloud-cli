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

func TestSpaceBackupScheduleDescribe_TextOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.GetBackupScheduleCalls.Returns(&spacebackupv1.GetBackupScheduleResponse{
		BackupSchedule: &spacebackupv1.BackupSchedule{
			Id:              "sched-1",
			Name:            "nightly",
			SpaceId:         "space-abc",
			CollectionName:  new("products"),
			Schedule:        "0 2 * * *",
			Status:          spacebackupv1.BackupScheduleStatus_BACKUP_SCHEDULE_STATUS_ACTIVE,
			RetentionPeriod: durationpb.New(30 * 24 * time.Hour),
			LastFiredAt:     timestamppb.Now(),
			CreatedAt:       timestamppb.Now(),
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "describe", "sched-1",
		"--space-id", "space-abc")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ID:          sched-1")
	assert.Contains(t, stdout, "Name:        nightly")
	assert.Contains(t, stdout, "Space:       space-abc")
	assert.Contains(t, stdout, "Collection:  products")
	assert.Contains(t, stdout, "Schedule:    0 2 * * *")
	assert.Contains(t, stdout, "Status:      ACTIVE")
	assert.Contains(t, stdout, "Paused:      no")
	assert.Contains(t, stdout, "Next Run:")
	assert.Contains(t, stdout, "Last Run:")
	assert.Contains(t, stdout, "Retention:   30 days")
	assert.Contains(t, stdout, "Created:")

	req, ok := env.ServerlessBackupServer.GetBackupScheduleCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "space-abc", req.GetSpaceId())
	assert.Equal(t, "sched-1", req.GetBackupScheduleId())
}

func TestSpaceBackupScheduleDescribe_PausedHidesNextRun(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.GetBackupScheduleCalls.Returns(&spacebackupv1.GetBackupScheduleResponse{
		BackupSchedule: &spacebackupv1.BackupSchedule{
			Id:       "sched-1",
			Schedule: "0 2 * * *",
			PausedAt: timestamppb.New(time.Now().Add(-time.Hour)),
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "describe", "sched-1",
		"--space-id", "space-abc")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Paused:      yes")
	assert.NotContains(t, stdout, "Next Run:")
}

func TestSpaceBackupScheduleDescribe_ScheduledPause(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.GetBackupScheduleCalls.Returns(&spacebackupv1.GetBackupScheduleResponse{
		BackupSchedule: &spacebackupv1.BackupSchedule{
			Id:       "sched-1",
			Schedule: "0 2 * * *",
			PausedAt: timestamppb.New(time.Now().Add(48 * time.Hour)),
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "describe", "sched-1",
		"--space-id", "space-abc")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Paused:      scheduled")
	assert.Contains(t, stdout, "Next Run:")
}

func TestSpaceBackupScheduleDescribe_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.GetBackupScheduleCalls.Returns(&spacebackupv1.GetBackupScheduleResponse{
		BackupSchedule: &spacebackupv1.BackupSchedule{Id: "sched-1", Name: "nightly"},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "describe", "sched-1",
		"--space-id", "space-abc", "--json")
	require.NoError(t, err)

	var result struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "sched-1", result.ID)
	assert.Equal(t, "nightly", result.Name)
}

func TestSpaceBackupScheduleDescribe_MissingSpaceID(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "describe", "sched-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "space-id")
	assert.Equal(t, 0, env.ServerlessBackupServer.GetBackupScheduleCalls.Count())
}

func TestSpaceBackupScheduleDescribe_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.GetBackupScheduleCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "describe", "sched-1",
		"--space-id", "space-abc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get backup schedule")
}
