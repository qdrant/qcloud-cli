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
	spacev1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestSpaceBackupDescribe_TextOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.GetBackupCalls.Returns(&spacebackupv1.GetBackupResponse{Backup: &spacebackupv1.Backup{
		Id:               "backup-1",
		Name:             "my-space-20260101",
		SpaceId:          "space-abc",
		CollectionName:   new("products"),
		Status:           spacebackupv1.BackupStatus_BACKUP_STATUS_SUCCEEDED,
		CreatedAt:        timestamppb.Now(),
		BackupScheduleId: new("sched-1"),
		RetentionPeriod:  durationpb.New(14 * 24 * time.Hour),
		Stats: &spacebackupv1.BackupStats{
			CollectionCount: new(uint32(1)),
			SizeBytes:       new(int64(512 * 1024 * 1024)),
			TotalPoints:     new(uint64(123456)),
			Duration:        durationpb.New(90 * time.Second),
			Progress:        new("0.5/0.5 GiB (100%)"),
		},
		SpaceInfo: &spacebackupv1.SpaceInfo{
			Name:          "my-space",
			CloudRegionId: "aws-eu-central-1",
			Configuration: &spacev1.SpaceConfiguration{
				AllowedIpSourceRanges: []string{"10.0.0.0/8"},
			},
		},
	}}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "describe", "backup-1")
	require.NoError(t, err)
	assert.Contains(t, stdout, "backup-1")
	assert.Contains(t, stdout, "my-space-20260101")
	assert.Contains(t, stdout, "space-abc")
	assert.Contains(t, stdout, "Collection:  products")
	assert.Contains(t, stdout, "SUCCEEDED")
	assert.Contains(t, stdout, "Schedule:    sched-1")
	assert.Contains(t, stdout, "Retention:   14 days")
	assert.Contains(t, stdout, "Collections: 1")
	assert.Contains(t, stdout, "512MiB")
	assert.Contains(t, stdout, "Points:      123456")
	assert.Contains(t, stdout, "Duration:    1m30s")
	assert.Contains(t, stdout, "(100%)")
	assert.Contains(t, stdout, "Space Snapshot:")
	assert.Contains(t, stdout, "my-space")
	assert.Contains(t, stdout, "aws-eu-central-1")
	assert.Contains(t, stdout, "10.0.0.0/8")

	req, ok := env.ServerlessBackupServer.GetBackupCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "backup-1", req.GetBackupId())
}

func TestSpaceBackupDescribe_MinimalBackup(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.GetBackupCalls.Returns(&spacebackupv1.GetBackupResponse{Backup: &spacebackupv1.Backup{
		Id:     "backup-1",
		Status: spacebackupv1.BackupStatus_BACKUP_STATUS_RUNNING,
	}}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "describe", "backup-1")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Collection:  (all)")
	assert.Contains(t, stdout, "Retention:   indefinite")
	assert.NotContains(t, stdout, "Schedule:")
	assert.NotContains(t, stdout, "Statistics:")
	assert.NotContains(t, stdout, "Space Snapshot")
}

func TestSpaceBackupDescribe_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.GetBackupCalls.Returns(&spacebackupv1.GetBackupResponse{Backup: &spacebackupv1.Backup{
		Id:      "backup-1",
		SpaceId: "space-abc",
	}}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "describe", "backup-1", "--json")
	require.NoError(t, err)

	var result struct {
		ID      string `json:"id"`
		SpaceID string `json:"spaceId"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "backup-1", result.ID)
	assert.Equal(t, "space-abc", result.SpaceID)
}

func TestSpaceBackupDescribe_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.GetBackupCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "describe", "backup-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get backup")
}

func TestSpaceBackupDescribe_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "describe")
	require.Error(t, err)
	assert.Equal(t, 0, env.ServerlessBackupServer.GetBackupCalls.Count())
}
