package iam_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

const testRoleCategory = "Cluster"

func TestUserDescribe_ByID(t *testing.T) {
	env := testutil.NewTestEnv(t)

	userID := testUserID
	cat := testRoleCategory
	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{
			{Id: userID, Email: "alice@example.com", Status: iamv1.UserStatus_USER_STATUS_ACTIVE},
		},
	}, nil)
	env.IAMServer.ListUserRolesCalls.Returns(&iamv1.ListUserRolesResponse{
		Roles: []*iamv1.Role{
			{
				Id:   "role-id-1",
				Name: "admin",
				Permissions: []*iamv1.Permission{
					{Value: "read:clusters", Category: &cat},
					{Value: "write:clusters", Category: &cat},
				},
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "user", "describe", userID)
	require.NoError(t, err)

	assert.Contains(t, stdout, userID)
	assert.Contains(t, stdout, "alice@example.com")
	assert.Contains(t, stdout, "ACTIVE")
	assert.Contains(t, stdout, "role-id-1")
	assert.Contains(t, stdout, "admin")
	assert.Contains(t, stdout, "read:clusters")
	assert.Contains(t, stdout, "write:clusters")
	assert.Contains(t, stdout, "Cluster")

	req, ok := env.IAMServer.ListUserRolesCalls.Last()
	require.True(t, ok)
	assert.Equal(t, userID, req.GetUserId())
}

func TestUserDescribe_PermissionsDeduplicatedWithRoles(t *testing.T) {
	env := testutil.NewTestEnv(t)

	userID := testUserID
	cat := testRoleCategory
	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{{Id: userID, Email: "alice@example.com"}},
	}, nil)
	env.IAMServer.ListUserRolesCalls.Returns(&iamv1.ListUserRolesResponse{
		Roles: []*iamv1.Role{
			{
				Id:   "role-id-1",
				Name: "admin",
				Permissions: []*iamv1.Permission{
					{Value: "read:clusters", Category: &cat},
					{Value: "write:clusters", Category: &cat},
				},
			},
			{
				Id:   "role-id-2",
				Name: "viewer",
				Permissions: []*iamv1.Permission{
					{Value: "read:clusters", Category: &cat},
				},
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "user", "describe", userID)
	require.NoError(t, err)

	// read:clusters appears in both roles — should be listed once with both role names
	assert.Contains(t, stdout, "admin, viewer")
	// write:clusters only in admin
	assert.Contains(t, stdout, "write:clusters")
}

func TestUserDescribe_NoPermissions(t *testing.T) {
	env := testutil.NewTestEnv(t)

	userID := testUserID
	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{
			{Id: userID, Email: "alice@example.com"},
		},
	}, nil)
	env.IAMServer.ListUserRolesCalls.Returns(&iamv1.ListUserRolesResponse{
		Roles: []*iamv1.Role{
			{Id: "role-id-1", Name: "viewer"},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "user", "describe", userID)
	require.NoError(t, err)

	assert.NotContains(t, stdout, "Effective Permissions")
}

func TestUserDescribe_ByEmail(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{
			{Id: "user-id-abc", Email: "alice@example.com"},
		},
	}, nil)
	env.IAMServer.ListUserRolesCalls.Returns(&iamv1.ListUserRolesResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "user", "describe", "alice@example.com")
	require.NoError(t, err)

	assert.Contains(t, stdout, "alice@example.com")
	req, ok := env.IAMServer.ListUserRolesCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "user-id-abc", req.GetUserId())
}

func TestUserDescribe_JSON(t *testing.T) {
	env := testutil.NewTestEnv(t)

	userID := testUserID
	cat := testRoleCategory
	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{
			{Id: userID, Email: "alice@example.com", Status: iamv1.UserStatus_USER_STATUS_ACTIVE},
		},
	}, nil)
	env.IAMServer.ListUserRolesCalls.Returns(&iamv1.ListUserRolesResponse{
		Roles: []*iamv1.Role{
			{Id: "role-id-1", Name: "admin", Permissions: []*iamv1.Permission{
				{Value: "read:clusters", Category: &cat},
			}},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "user", "describe", userID, "--json")
	require.NoError(t, err)

	var got struct {
		User struct {
			Id    string `json:"id"`
			Email string `json:"email"`
		} `json:"user"`
		Roles []struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"roles"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &got))
	assert.Equal(t, userID, got.User.Id)
	assert.Equal(t, "alice@example.com", got.User.Email)
	require.Len(t, got.Roles, 1)
	assert.Equal(t, "role-id-1", got.Roles[0].Id)
	assert.Equal(t, "admin", got.Roles[0].Name)
}

func TestUserDescribe_NotFound(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{Items: nil}, nil)

	_, _, err := testutil.Exec(t, env, "iam", "user", "describe", "nobody@example.com")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
