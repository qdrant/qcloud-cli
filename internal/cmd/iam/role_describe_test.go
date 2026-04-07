package iam_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestRoleDescribe_TextOutput(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))

	now := time.Now()
	env.IAMServer.GetRoleCalls.Returns(&iamv1.GetRoleResponse{
		Role: &iamv1.Role{
			Id:             "role-abc",
			Name:           "Cluster Viewer",
			Description:    "Read-only access to clusters",
			RoleType:       iamv1.RoleType_ROLE_TYPE_CUSTOM,
			CreatedAt:      timestamppb.New(now),
			LastModifiedAt: timestamppb.New(now),
			Permissions: []*iamv1.Permission{
				{Value: "read:clusters", Category: new("Cluster")},
				{Value: "read:backups", Category: new("Backup")},
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "role", "describe", "role-abc")
	require.NoError(t, err)
	assert.Contains(t, stdout, "role-abc")
	assert.Contains(t, stdout, "Cluster Viewer")
	assert.Contains(t, stdout, "Read-only access to clusters")
	assert.Contains(t, stdout, "CUSTOM")
	assert.Contains(t, stdout, "read:clusters")
	assert.Contains(t, stdout, "read:backups")
	assert.Contains(t, stdout, "Cluster")
	assert.Contains(t, stdout, "Backup")

	req, ok := env.IAMServer.GetRoleCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "role-abc", req.GetRoleId())
}

func TestRoleDescribe_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.GetRoleCalls.Returns(&iamv1.GetRoleResponse{
		Role: &iamv1.Role{
			Id:       "role-json",
			Name:     "Test",
			RoleType: iamv1.RoleType_ROLE_TYPE_CUSTOM,
			Permissions: []*iamv1.Permission{
				{Value: "read:clusters"},
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "role", "describe", "role-json", "--json")
	require.NoError(t, err)

	var result struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "role-json", result.ID)
	assert.Equal(t, "Test", result.Name)
}

func TestRoleDescribe_BackendError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.GetRoleCalls.Returns(nil, fmt.Errorf("not found"))

	_, _, err := testutil.Exec(t, env, "iam", "role", "describe", "role-abc")
	require.Error(t, err)
}

func TestRoleDescribe_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "iam", "role", "describe")
	require.Error(t, err)
}
