package iam_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestRoleRemovePermission_RemovesExisting(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))

	env.IAMServer.GetRoleCalls.Returns(&iamv1.GetRoleResponse{
		Role: &iamv1.Role{
			Id:        "role-abc",
			Name:      "Test",
			AccountId: "test-account-id",
			RoleType:  iamv1.RoleType_ROLE_TYPE_CUSTOM,
			Permissions: []*iamv1.Permission{
				{Value: "read:clusters"},
				{Value: "write:clusters"},
			},
		},
	}, nil)

	env.IAMServer.UpdateRoleCalls.Returns(&iamv1.UpdateRoleResponse{
		Role: &iamv1.Role{Id: "role-abc"},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "role", "remove-permission", "role-abc",
		"--permission", "write:clusters",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "Removed 1 permission(s)")

	req, ok := env.IAMServer.UpdateRoleCalls.Last()
	require.True(t, ok)
	perms := req.GetRole().GetPermissions()
	require.Len(t, perms, 1)
	assert.Equal(t, "read:clusters", perms[0].GetValue())
}

func TestRoleRemovePermission_NoMatch(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.GetRoleCalls.Returns(&iamv1.GetRoleResponse{
		Role: &iamv1.Role{
			Id:       "role-abc",
			RoleType: iamv1.RoleType_ROLE_TYPE_CUSTOM,
			Permissions: []*iamv1.Permission{
				{Value: "read:clusters"},
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "role", "remove-permission", "role-abc",
		"--permission", "write:backups",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "No matching permissions to remove")
	assert.Equal(t, 0, env.IAMServer.UpdateRoleCalls.Count())
}

func TestRoleRemovePermission_CannotRemoveAll(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.GetRoleCalls.Returns(&iamv1.GetRoleResponse{
		Role: &iamv1.Role{
			Id:       "role-abc",
			RoleType: iamv1.RoleType_ROLE_TYPE_CUSTOM,
			Permissions: []*iamv1.Permission{
				{Value: "read:clusters"},
			},
		},
	}, nil)

	_, _, err := testutil.Exec(t, env, "iam", "role", "remove-permission", "role-abc",
		"--permission", "read:clusters",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one permission")
	assert.Equal(t, 0, env.IAMServer.UpdateRoleCalls.Count())
}

func TestRoleRemovePermission_BackendError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.GetRoleCalls.Returns(nil, fmt.Errorf("not found"))

	_, _, err := testutil.Exec(t, env, "iam", "role", "remove-permission", "role-abc",
		"--permission", "read:clusters",
	)
	require.Error(t, err)
}

func TestRoleRemovePermission_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "iam", "role", "remove-permission")
	require.Error(t, err)
}
