package space_test

import (
	"context"
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

func setupScheduleUpdate(env *testutil.TestEnv, existing *spacebackupv1.BackupSchedule) {
	env.ServerlessBackupServer.GetBackupScheduleCalls.Returns(&spacebackupv1.GetBackupScheduleResponse{
		BackupSchedule: existing,
	}, nil)
	env.ServerlessBackupServer.UpdateBackupScheduleCalls.Always(
		func(_ context.Context, req *spacebackupv1.UpdateBackupScheduleRequest) (*spacebackupv1.UpdateBackupScheduleResponse, error) {
			return &spacebackupv1.UpdateBackupScheduleResponse{BackupSchedule: req.GetBackupSchedule()}, nil
		})
}

func existingSchedule() *spacebackupv1.BackupSchedule {
	return &spacebackupv1.BackupSchedule{
		Id:              "sched-1",
		AccountId:       "test-account-id",
		SpaceId:         "space-abc",
		Name:            "nightly",
		Schedule:        "0 2 * * *",
		RetentionPeriod: durationpb.New(7 * 24 * time.Hour),
	}
}

func TestSpaceBackupScheduleUpdate_ChangedFieldsOnly(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupScheduleUpdate(env, existingSchedule())

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "update", "sched-1",
		"--space-id", "space-abc", "--schedule", "0 4 * * *")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Backup schedule sched-1 (nightly) updated.")

	getReq, ok := env.ServerlessBackupServer.GetBackupScheduleCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "space-abc", getReq.GetSpaceId())
	assert.Equal(t, "sched-1", getReq.GetBackupScheduleId())

	req, ok := env.ServerlessBackupServer.UpdateBackupScheduleCalls.Last()
	require.True(t, ok)
	sched := req.GetBackupSchedule()
	assert.Equal(t, "0 4 * * *", sched.GetSchedule())
	assert.Equal(t, "nightly", sched.GetName())
	assert.Equal(t, 7*24*time.Hour, sched.GetRetentionPeriod().AsDuration())
	assert.Nil(t, sched.PausedAt)
}

func TestSpaceBackupScheduleUpdate_NameAndRetention(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupScheduleUpdate(env, existingSchedule())

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "update", "sched-1",
		"--space-id", "space-abc", "--name", "renamed", "--retention-days", "90")
	require.NoError(t, err)

	req, ok := env.ServerlessBackupServer.UpdateBackupScheduleCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "renamed", req.GetBackupSchedule().GetName())
	assert.Equal(t, 90*24*time.Hour, req.GetBackupSchedule().GetRetentionPeriod().AsDuration())
}

func TestSpaceBackupScheduleUpdate_Pause(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupScheduleUpdate(env, existingSchedule())

	before := time.Now()
	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "update", "sched-1",
		"--space-id", "space-abc", "--pause")
	require.NoError(t, err)

	req, ok := env.ServerlessBackupServer.UpdateBackupScheduleCalls.Last()
	require.True(t, ok)
	require.NotNil(t, req.GetBackupSchedule().PausedAt)
	assert.WithinRange(t, req.GetBackupSchedule().GetPausedAt().AsTime(), before.Add(-time.Second), time.Now().Add(time.Second))
}

func TestSpaceBackupScheduleUpdate_Resume(t *testing.T) {
	env := testutil.NewTestEnv(t)
	paused := existingSchedule()
	paused.PausedAt = timestamppb.New(time.Now().Add(-time.Hour))
	setupScheduleUpdate(env, paused)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "update", "sched-1",
		"--space-id", "space-abc", "--resume")
	require.NoError(t, err)

	req, ok := env.ServerlessBackupServer.UpdateBackupScheduleCalls.Last()
	require.True(t, ok)
	assert.Nil(t, req.GetBackupSchedule().PausedAt)
}

func TestSpaceBackupScheduleUpdate_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupScheduleUpdate(env, existingSchedule())

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "update", "sched-1",
		"--space-id", "space-abc", "--name", "renamed", "--json")
	require.NoError(t, err)

	var result struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "sched-1", result.ID)
	assert.Equal(t, "renamed", result.Name)
}

func TestSpaceBackupScheduleUpdate_InputErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"missing space id", []string{"sched-1", "--name", "x"}, "space-id"},
		{"pause and resume", []string{"sched-1", "--space-id", "space-abc", "--pause", "--resume"}, "none of the others can be"},
		{"retention too low", []string{"sched-1", "--space-id", "space-abc", "--retention-days", "0"}, "between 1 and 365"},
		{"retention too high", []string{"sched-1", "--space-id", "space-abc", "--retention-days", "366"}, "between 1 and 365"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := testutil.NewTestEnv(t)
			setupScheduleUpdate(env, existingSchedule())

			args := append([]string{"serverless", "space", "backup", "schedule", "update"}, tt.args...)
			_, _, err := testutil.Exec(t, env, args...)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.Equal(t, 0, env.ServerlessBackupServer.UpdateBackupScheduleCalls.Count())
		})
	}
}

func TestSpaceBackupScheduleUpdate_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessBackupServer.GetBackupScheduleCalls.Returns(&spacebackupv1.GetBackupScheduleResponse{
		BackupSchedule: existingSchedule(),
	}, nil)
	env.ServerlessBackupServer.UpdateBackupScheduleCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "update", "sched-1",
		"--space-id", "space-abc", "--name", "x")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update backup schedule")
}
