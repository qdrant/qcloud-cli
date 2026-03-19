package backup_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestScheduleDescribe_TextOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.EXPECT().GetBackupSchedule(mock.Anything, mock.MatchedBy(func(req *backupv1.GetBackupScheduleRequest) bool {
		assert.Equal(t, "test-account-id", req.GetAccountId())
		assert.Equal(t, "cluster-abc", req.GetClusterId())
		assert.Equal(t, "schedule-1", req.GetBackupScheduleId())
		return true
	})).Return(
		&backupv1.GetBackupScheduleResponse{
			BackupSchedule: &backupv1.BackupSchedule{
				Id:        "schedule-1",
				ClusterId: "cluster-abc",
				Schedule:  "0 2 * * *",
				Status:    backupv1.BackupScheduleStatus_BACKUP_SCHEDULE_STATUS_ACTIVE,
				CreatedAt: timestamppb.Now(),
			},
		},
		nil,
	)

	stdout, _, err := testutil.Exec(t, env, "backup", "schedule", "describe", "schedule-1", "--cluster-id=cluster-abc")
	require.NoError(t, err)
	assert.Contains(t, stdout, "schedule-1")
	assert.Contains(t, stdout, "cluster-abc")
	assert.Contains(t, stdout, "0 2 * * *")
	assert.Contains(t, stdout, "Next Run:")
	assert.Contains(t, stdout, "ACTIVE")
}

func TestScheduleDescribe_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.EXPECT().GetBackupSchedule(mock.Anything, mock.Anything).Return(
		&backupv1.GetBackupScheduleResponse{
			BackupSchedule: &backupv1.BackupSchedule{Id: "schedule-json", Schedule: "0 4 * * *"},
		},
		nil,
	)

	stdout, _, err := testutil.Exec(t, env, "backup", "schedule", "describe", "schedule-json", "--cluster-id=cluster-abc", "--json")
	require.NoError(t, err)

	var result struct {
		ID       string `json:"id"`
		Schedule string `json:"schedule"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "schedule-json", result.ID)
	assert.Equal(t, "0 4 * * *", result.Schedule)
}

func TestScheduleDescribe_MissingClusterID(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "backup", "schedule", "describe", "schedule-1")
	require.Error(t, err)
}

func TestScheduleDescribe_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.EXPECT().GetBackupSchedule(mock.Anything, mock.Anything).Return(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "backup", "schedule", "describe", "schedule-1", "--cluster-id=cluster-abc")
	require.Error(t, err)
}
