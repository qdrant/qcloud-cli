package iam_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestRoleAssignPermission_AddsNew(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))

	env.IAMServer.GetRoleCalls.Returns(&iamv1.GetRoleResponse{
		Role: &iamv1.Role{
			Id:        "role-abc",
			Name:      "Test",
			AccountId: "test-account-id",
			RoleType:  iamv1.RoleType_ROLE_TYPE_CUSTOM,
			Permissions: []*iamv1.Permission{
				{Value: "read:clusters"},
			},
		},
	}, nil)

	env.IAMServer.UpdateRoleCalls.Returns(&iamv1.UpdateRoleResponse{
		Role: &iamv1.Role{Id: "role-abc"},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "role", "assign-permission", "role-abc",
		"--permission", "write:clusters",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "Added 1 permission(s)")

	req, ok := env.IAMServer.UpdateRoleCalls.Last()
	require.True(t, ok)
	perms := req.GetRole().GetPermissions()
	require.Len(t, perms, 2)
	assert.Equal(t, "read:clusters", perms[0].GetValue())
	assert.Equal(t, "write:clusters", perms[1].GetValue())
}

func TestRoleAssignPermission_Deduplicates(t *testing.T) {
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

	stdout, _, err := testutil.Exec(t, env, "iam", "role", "assign-permission", "role-abc",
		"--permission", "read:clusters",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "No new permissions to add")
	assert.Equal(t, 0, env.IAMServer.UpdateRoleCalls.Count())
}

func TestRoleAssignPermission_BackendError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.GetRoleCalls.Returns(nil, fmt.Errorf("not found"))

	_, _, err := testutil.Exec(t, env, "iam", "role", "assign-permission", "role-abc",
		"--permission", "read:clusters",
	)
	require.Error(t, err)
}

func TestRoleAssignPermission_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "iam", "role", "assign-permission")
	require.Error(t, err)
}

func TestRoleAssignPermission_MissingPermission(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "iam", "role", "assign-permission", "role-abc")
	require.Error(t, err)
}
