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

func TestRoleList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))

	env.IAMServer.ListRolesCalls.Returns(&iamv1.ListRolesResponse{
		Items: []*iamv1.Role{
			{
				Id:        "role-abc",
				Name:      "Admin",
				RoleType:  iamv1.RoleType_ROLE_TYPE_SYSTEM,
				CreatedAt: timestamppb.New(time.Now().Add(-24 * time.Hour)),
				Permissions: []*iamv1.Permission{
					{Value: "read:clusters"},
					{Value: "write:clusters"},
				},
			},
			{
				Id:        "role-def",
				Name:      "Viewer",
				RoleType:  iamv1.RoleType_ROLE_TYPE_CUSTOM,
				CreatedAt: timestamppb.New(time.Now().Add(-1 * time.Hour)),
				Permissions: []*iamv1.Permission{
					{Value: "read:clusters"},
				},
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "role", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ID")
	assert.Contains(t, stdout, "NAME")
	assert.Contains(t, stdout, "TYPE")
	assert.Contains(t, stdout, "PERMISSIONS")
	assert.Contains(t, stdout, "CREATED")
	assert.Contains(t, stdout, "role-abc")
	assert.Contains(t, stdout, "Admin")
	assert.Contains(t, stdout, "SYSTEM")
	assert.Contains(t, stdout, "2")
	assert.Contains(t, stdout, "role-def")
	assert.Contains(t, stdout, "Viewer")
	assert.Contains(t, stdout, "CUSTOM")

	req, ok := env.IAMServer.ListRolesCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
}

func TestRoleList_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListRolesCalls.Returns(&iamv1.ListRolesResponse{
		Items: []*iamv1.Role{
			{Id: "role-json", Name: "Test Role", RoleType: iamv1.RoleType_ROLE_TYPE_CUSTOM},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "role", "list", "--json")
	require.NoError(t, err)

	var result struct {
		Items []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Items, 1)
	assert.Equal(t, "role-json", result.Items[0].ID)
	assert.Equal(t, "Test Role", result.Items[0].Name)
}

func TestRoleList_Empty(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListRolesCalls.Returns(&iamv1.ListRolesResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "role", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ID")
	assert.Contains(t, stdout, "NAME")
}

func TestRoleList_BackendError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListRolesCalls.Returns(nil, fmt.Errorf("internal server error"))

	_, _, err := testutil.Exec(t, env, "iam", "role", "list")
	require.Error(t, err)
}
