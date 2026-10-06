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

func TestSpaceBackupScheduleList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.ListBackupSchedulesCalls.Returns(&spacebackupv1.ListBackupSchedulesResponse{
		Items: []*spacebackupv1.BackupSchedule{
			{
				Id:              "sched-1",
				Name:            "nightly",
				SpaceId:         "space-abc",
				Schedule:        "0 2 * * *",
				Status:          spacebackupv1.BackupScheduleStatus_BACKUP_SCHEDULE_STATUS_ACTIVE,
				RetentionPeriod: durationpb.New(7 * 24 * time.Hour),
				LastFiredAt:     timestamppb.Now(),
			},
			{
				Id:             "sched-2",
				Name:           "weekly-products",
				SpaceId:        "space-abc",
				Schedule:       "0 3 * * 0",
				CollectionName: new("products"),
				Status:         spacebackupv1.BackupScheduleStatus_BACKUP_SCHEDULE_STATUS_ACTIVE,
				PausedAt:       timestamppb.New(time.Now().Add(-time.Hour)),
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "list")
	require.NoError(t, err)
	for _, h := range []string{"ID", "NAME", "SPACE", "SCHEDULE", "COLLECTION", "STATUS", "PAUSED", "RETENTION", "LAST RUN"} {
		assert.Contains(t, stdout, h)
	}

	assert.Contains(t, stdout, "sched-1")
	assert.Contains(t, stdout, "nightly")
	assert.Contains(t, stdout, "0 2 * * *")
	assert.Contains(t, stdout, "(all)")
	assert.Contains(t, stdout, "ACTIVE")
	assert.Contains(t, stdout, "7 days")
	assert.Contains(t, stdout, "weekly-products")
	assert.Contains(t, stdout, "products")
	assert.Contains(t, stdout, "indefinite")
	assert.Contains(t, stdout, "yes")

	req, ok := env.ServerlessBackupServer.ListBackupSchedulesCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Nil(t, req.SpaceId)
}

func TestSpaceBackupScheduleList_SpaceFilter(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.ListBackupSchedulesCalls.Returns(&spacebackupv1.ListBackupSchedulesResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "list", "--space-id", "space-abc")
	require.NoError(t, err)

	req, ok := env.ServerlessBackupServer.ListBackupSchedulesCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "space-abc", req.GetSpaceId())
}

func TestSpaceBackupScheduleList_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.ListBackupSchedulesCalls.Returns(&spacebackupv1.ListBackupSchedulesResponse{
		Items: []*spacebackupv1.BackupSchedule{{Id: "sched-1", Schedule: "0 2 * * *"}},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "list", "--json")
	require.NoError(t, err)

	var result struct {
		Items []struct {
			ID       string `json:"id"`
			Schedule string `json:"schedule"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Items, 1)
	assert.Equal(t, "sched-1", result.Items[0].ID)
	assert.Equal(t, "0 2 * * *", result.Items[0].Schedule)
}

func TestSpaceBackupScheduleList_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.ListBackupSchedulesCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "schedule", "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list backup schedules")
}
