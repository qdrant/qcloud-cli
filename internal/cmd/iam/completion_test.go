package iam_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestRoleIDCompletion_Describe(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.IAMServer.ListRolesCalls.Returns(&iamv1.ListRolesResponse{
		Items: []*iamv1.Role{
			{Id: "role-uuid-1", Name: "Admin"},
			{Id: "role-uuid-2", Name: "Viewer"},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "__complete", "iam", "role", "describe", "")
	require.NoError(t, err)
	assert.Contains(t, stdout, "role-uuid-1")
	assert.Contains(t, stdout, "Admin")
	assert.Contains(t, stdout, "role-uuid-2")
	assert.Contains(t, stdout, "Viewer")
}

func TestRoleIDCompletion_Delete(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.IAMServer.ListRolesCalls.Returns(&iamv1.ListRolesResponse{
		Items: []*iamv1.Role{
			{Id: "role-uuid-1", Name: "Admin"},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "__complete", "iam", "role", "delete", "")
	require.NoError(t, err)
	assert.Contains(t, stdout, "role-uuid-1")
	assert.Contains(t, stdout, "Admin")
}

func TestRoleIDCompletion_AssignPermission(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.IAMServer.ListRolesCalls.Returns(&iamv1.ListRolesResponse{
		Items: []*iamv1.Role{
			{Id: "role-uuid-1", Name: "Custom Role"},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "__complete", "iam", "role", "assign-permission", "")
	require.NoError(t, err)
	assert.Contains(t, stdout, "role-uuid-1")
	assert.Contains(t, stdout, "Custom Role")
}

func TestPermissionCompletion_Create(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.IAMServer.ListPermissionsCalls.Returns(&iamv1.ListPermissionsResponse{
		Permissions: []*iamv1.Permission{
			{Value: "read:clusters", Category: new("Cluster")},
			{Value: "write:backups", Category: new("Backup")},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "__complete", "iam", "role", "create", "--name", "test", "--permission", "")
	require.NoError(t, err)
	assert.Contains(t, stdout, "read:clusters")
	assert.Contains(t, stdout, "Cluster")
	assert.Contains(t, stdout, "write:backups")
	assert.Contains(t, stdout, "Backup")
}

func TestPermissionCompletion_AssignPermission(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.IAMServer.ListPermissionsCalls.Returns(&iamv1.ListPermissionsResponse{
		Permissions: []*iamv1.Permission{
			{Value: "read:clusters", Category: new("Cluster")},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "__complete", "iam", "role", "assign-permission", "some-role-id", "--permission", "")
	require.NoError(t, err)
	assert.Contains(t, stdout, "read:clusters")
	assert.Contains(t, stdout, "Cluster")
}

func TestPermissionCompletion_RemovePermission(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.IAMServer.ListPermissionsCalls.Returns(&iamv1.ListPermissionsResponse{
		Permissions: []*iamv1.Permission{
			{Value: "write:backups", Category: new("Backup")},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "__complete", "iam", "role", "remove-permission", "some-role-id", "--permission", "")
	require.NoError(t, err)
	assert.Contains(t, stdout, "write:backups")
	assert.Contains(t, stdout, "Backup")
}
