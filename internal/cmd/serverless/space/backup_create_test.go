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

func echoCreateBackup(env *testutil.TestEnv) {
	env.ServerlessBackupServer.CreateBackupCalls.Always(
		func(_ context.Context, req *spacebackupv1.CreateBackupRequest) (*spacebackupv1.CreateBackupResponse, error) {
			b := req.GetBackup()
			b.Id = "backup-new"
			return &spacebackupv1.CreateBackupResponse{Backup: b}, nil
		})
}

func TestSpaceBackupCreate_Basic(t *testing.T) {
	env := testutil.NewTestEnv(t)
	echoCreateBackup(env)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "create", "--space-id", "space-abc")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Backup backup-new created for space space-abc.")

	req, ok := env.ServerlessBackupServer.CreateBackupCalls.Last()
	require.True(t, ok)
	b := req.GetBackup()
	assert.Equal(t, "test-account-id", b.GetAccountId())
	assert.Equal(t, "space-abc", b.GetSpaceId())
	assert.Nil(t, b.CollectionName)
	assert.Nil(t, b.RetentionPeriod)
}

func TestSpaceBackupCreate_CollectionAndRetention(t *testing.T) {
	env := testutil.NewTestEnv(t)
	echoCreateBackup(env)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "create",
		"--space-id", "space-abc", "--collection", "products", "--retention-days", "7")
	require.NoError(t, err)

	req, ok := env.ServerlessBackupServer.CreateBackupCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "products", req.GetBackup().GetCollectionName())
	assert.Equal(t, 7*24*time.Hour, req.GetBackup().GetRetentionPeriod().AsDuration())
}

func TestSpaceBackupCreate_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	echoCreateBackup(env)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "create", "--space-id", "space-abc", "--json")
	require.NoError(t, err)

	var result struct {
		ID      string `json:"id"`
		SpaceID string `json:"spaceId"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "backup-new", result.ID)
	assert.Equal(t, "space-abc", result.SpaceID)
}

func TestSpaceBackupCreate_InputErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"missing space id", nil, "space-id"},
		{"retention too low", []string{"--space-id", "space-abc", "--retention-days", "0"}, "between 1 and 365"},
		{"retention too high", []string{"--space-id", "space-abc", "--retention-days", "366"}, "between 1 and 365"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := testutil.NewTestEnv(t)

			args := append([]string{"serverless", "space", "backup", "create"}, tt.args...)
			_, _, err := testutil.Exec(t, env, args...)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.Equal(t, 0, env.ServerlessBackupServer.CreateBackupCalls.Count())
		})
	}
}

func TestSpaceBackupCreate_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessBackupServer.CreateBackupCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "backup", "create", "--space-id", "space-abc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create backup")
}
