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

func TestScheduleList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.EXPECT().ListBackupSchedules(mock.Anything, mock.MatchedBy(func(req *backupv1.ListBackupSchedulesRequest) bool {
		assert.Equal(t, "test-account-id", req.GetAccountId())
		return true
	})).Return(
		&backupv1.ListBackupSchedulesResponse{
			Items: []*backupv1.BackupSchedule{
				{
					Id:        "schedule-1",
					ClusterId: "cluster-abc",
					Schedule:  "0 2 * * *",
					Status:    backupv1.BackupScheduleStatus_BACKUP_SCHEDULE_STATUS_ACTIVE,
					CreatedAt: timestamppb.Now(),
				},
			},
		},
		nil,
	)

	stdout, _, err := testutil.Exec(t, env, "backup", "schedule", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ID")
	assert.Contains(t, stdout, "CLUSTER")
	assert.Contains(t, stdout, "SCHEDULE")
	assert.Contains(t, stdout, "STATUS")
	assert.Contains(t, stdout, "NEXT RUN")
	assert.Contains(t, stdout, "schedule-1")
	assert.Contains(t, stdout, "cluster-abc")
	assert.Contains(t, stdout, "0 2 * * *")
	assert.Contains(t, stdout, "ACTIVE")
}

func TestScheduleList_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.EXPECT().ListBackupSchedules(mock.Anything, mock.Anything).Return(
		&backupv1.ListBackupSchedulesResponse{
			Items: []*backupv1.BackupSchedule{
				{Id: "schedule-json", Schedule: "0 3 * * *"},
			},
		},
		nil,
	)

	stdout, _, err := testutil.Exec(t, env, "backup", "schedule", "list", "--json")
	require.NoError(t, err)

	var result struct {
		Items []struct {
			ID       string `json:"id"`
			Schedule string `json:"schedule"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Items, 1)
	assert.Equal(t, "schedule-json", result.Items[0].ID)
	assert.Equal(t, "0 3 * * *", result.Items[0].Schedule)
}

func TestScheduleList_EmptyResponse(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.EXPECT().ListBackupSchedules(mock.Anything, mock.Anything).Return(&backupv1.ListBackupSchedulesResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "backup", "schedule", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ID")
	assert.Contains(t, stdout, "SCHEDULE")
}

func TestScheduleList_ClusterIDFilter(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BackupServer.EXPECT().ListBackupSchedules(mock.Anything, mock.MatchedBy(func(req *backupv1.ListBackupSchedulesRequest) bool {
		assert.Equal(t, "my-cluster", req.GetClusterId())
		return true
	})).Return(&backupv1.ListBackupSchedulesResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "backup", "schedule", "list", "--cluster-id=my-cluster")
	require.NoError(t, err)
}
