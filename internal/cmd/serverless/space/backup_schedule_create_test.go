package space_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spacebackupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func echoCreateBackupSchedule(env *testutil.TestEnv) {
	env.ServerlessBackupServer.CreateBackupScheduleCalls.Always(
		func(_ context.Context, req *spacebackupv1.CreateBackupScheduleRequest) (*spacebackupv1.CreateBackupScheduleResponse, error) {
			sched := req.GetBackupSchedule()
			sched.Id = "sched-new"
			return &spacebackupv1.CreateBackupScheduleResponse{BackupSchedule: sched}, nil
		})
}

func TestSpaceBackupScheduleCreate_Basic(t *testing.T) {
	env := testutil.NewTestEnv(t)
	echoCreateBackupSchedule(env)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "create",
		"--space-id", "space-abc", "--name", "nightly", "--schedule", "0 2 * * *")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Backup schedule sched-new (nightly) created for space space-abc.")

	req, ok := env.ServerlessBackupServer.CreateBackupScheduleCalls.Last()
	require.True(t, ok)
	sched := req.GetBackupSchedule()
	assert.Equal(t, "test-account-id", sched.GetAccountId())
	assert.Equal(t, "space-abc", sched.GetSpaceId())
	assert.Equal(t, "nightly", sched.GetName())
	assert.Equal(t, "0 2 * * *", sched.GetSchedule())
	assert.Nil(t, sched.CollectionName)
	assert.Nil(t, sched.RetentionPeriod)
}

func TestSpaceBackupScheduleCreate_CollectionAndRetention(t *testing.T) {
	env := testutil.NewTestEnv(t)
	echoCreateBackupSchedule(env)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "create",
		"--space-id", "space-abc", "--name", "nightly", "--schedule", "0 2 * * *",
		"--collection", "products", "--retention-days", "30")
	require.NoError(t, err)

	req, ok := env.ServerlessBackupServer.CreateBackupScheduleCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "products", req.GetBackupSchedule().GetCollectionName())
	assert.Equal(t, 30*24*time.Hour, req.GetBackupSchedule().GetRetentionPeriod().AsDuration())
}

func TestSpaceBackupScheduleCreate_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	echoCreateBackupSchedule(env)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "create",
		"--space-id", "space-abc", "--name", "nightly", "--schedule", "0 2 * * *", "--json")
	require.NoError(t, err)

	var result struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Schedule string `json:"schedule"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "sched-new", result.ID)
	assert.Equal(t, "nightly", result.Name)
	assert.Equal(t, "0 2 * * *", result.Schedule)
}

func TestSpaceBackupScheduleCreate_InputErrors(t *testing.T) {
	base := []string{"--space-id", "space-abc", "--name", "nightly", "--schedule", "0 2 * * *"}
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"missing space id", []string{"--name", "nightly", "--schedule", "0 2 * * *"}, "space-id"},
		{"missing name", []string{"--space-id", "space-abc", "--schedule", "0 2 * * *"}, "name"},
		{"missing schedule", []string{"--space-id", "space-abc", "--name", "nightly"}, "schedule"},
		{"retention too low", append(base, "--retention-days", "0"), "between 1 and 365"},
		{"retention too high", append(base, "--retention-days", "366"), "between 1 and 365"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := testutil.NewTestEnv(t)

			args := append([]string{"serverless", "space", "backup", "schedule", "create"}, tt.args...)
			_, _, err := testutil.Exec(t, env, args...)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.Equal(t, 0, env.ServerlessBackupServer.CreateBackupScheduleCalls.Count())
		})
	}
}

func TestSpaceBackupScheduleCreate_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.CreateBackupScheduleCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "create",
		"--space-id", "space-abc", "--name", "nightly", "--schedule", "0 2 * * *")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create backup schedule")
}
