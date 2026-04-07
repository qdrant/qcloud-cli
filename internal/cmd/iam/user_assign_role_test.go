package iam_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestUserAssignRole_ByRoleID(t *testing.T) {
	env := testutil.NewTestEnv(t)

	userID := testUserID
	roleID := testRoleID

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{{Id: userID, Email: "alice@example.com"}},
	}, nil)
	env.IAMServer.AssignUserRolesCalls.Returns(&iamv1.AssignUserRolesResponse{}, nil)
	env.IAMServer.ListUserRolesCalls.Returns(&iamv1.ListUserRolesResponse{
		Roles: []*iamv1.Role{{Id: roleID, Name: "admin"}},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "user", "assign-role",
		"alice@example.com", "--role", roleID)
	require.NoError(t, err)
	assert.Contains(t, stdout, "alice@example.com")
	assert.Contains(t, stdout, roleID)
	assert.Contains(t, stdout, "admin")

	req, ok := env.IAMServer.AssignUserRolesCalls.Last()
	require.True(t, ok)
	assert.Equal(t, userID, req.GetUserId())
	assert.Equal(t, []string{roleID}, req.GetRoleIdsToAdd())
}

func TestUserAssignRole_ByRoleName(t *testing.T) {
	env := testutil.NewTestEnv(t)

	userID := testUserID
	roleID := testRoleID

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{{Id: userID, Email: "alice@example.com"}},
	}, nil)
	env.IAMServer.ListRolesCalls.Returns(&iamv1.ListRolesResponse{
		Items: []*iamv1.Role{{Id: roleID, Name: "admin"}},
	}, nil)
	env.IAMServer.AssignUserRolesCalls.Returns(&iamv1.AssignUserRolesResponse{}, nil)
	env.IAMServer.ListUserRolesCalls.Returns(&iamv1.ListUserRolesResponse{
		Roles: []*iamv1.Role{{Id: roleID, Name: "admin"}},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "user", "assign-role",
		"alice@example.com", "--role", "admin")
	require.NoError(t, err)
	assert.Contains(t, stdout, "admin")

	req, ok := env.IAMServer.AssignUserRolesCalls.Last()
	require.True(t, ok)
	assert.Equal(t, []string{roleID}, req.GetRoleIdsToAdd())
}

func TestUserAssignRole_MissingRole(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "iam", "user", "assign-role", "alice@example.com")
	require.Error(t, err)
}

func TestUserAssignRole_ResolveUserError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListUsersCalls.Returns(nil, fmt.Errorf("connection refused"))

	_, _, err := testutil.Exec(t, env, "iam", "user", "assign-role",
		"alice@example.com", "--role", testRoleID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestUserAssignRole_ResolveRoleError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{{Id: testUserID, Email: "alice@example.com"}},
	}, nil)
	env.IAMServer.ListRolesCalls.Returns(nil, fmt.Errorf("service unavailable"))

	_, _, err := testutil.Exec(t, env, "iam", "user", "assign-role",
		"alice@example.com", "--role", "admin")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service unavailable")
}

func TestUserAssignRole_RoleNameNotFound(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{{Id: testUserID, Email: "alice@example.com"}},
	}, nil)
	env.IAMServer.ListRolesCalls.Returns(&iamv1.ListRolesResponse{
		Items: []*iamv1.Role{{Id: testRoleID, Name: "viewer"}},
	}, nil)

	_, _, err := testutil.Exec(t, env, "iam", "user", "assign-role",
		"alice@example.com", "--role", "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `role "nonexistent" not found`)
}

func TestUserAssignRole_AssignError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{{Id: testUserID, Email: "alice@example.com"}},
	}, nil)
	env.IAMServer.AssignUserRolesCalls.Returns(nil, fmt.Errorf("forbidden"))

	_, _, err := testutil.Exec(t, env, "iam", "user", "assign-role",
		"alice@example.com", "--role", testRoleID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestUserAssignRole_ListUserRolesError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{{Id: testUserID, Email: "alice@example.com"}},
	}, nil)
	env.IAMServer.AssignUserRolesCalls.Returns(&iamv1.AssignUserRolesResponse{}, nil)
	env.IAMServer.ListUserRolesCalls.Returns(nil, fmt.Errorf("timeout"))

	_, _, err := testutil.Exec(t, env, "iam", "user", "assign-role",
		"alice@example.com", "--role", testRoleID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}
