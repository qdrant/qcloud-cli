package iam_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestRoleUpdate_Name(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))

	env.IAMServer.GetRoleCalls.Returns(&iamv1.GetRoleResponse{
		Role: &iamv1.Role{
			Id:       "role-abc",
			Name:     "Old Name",
			RoleType: iamv1.RoleType_ROLE_TYPE_CUSTOM,
			Permissions: []*iamv1.Permission{
				{Value: "read:clusters"},
			},
		},
	}, nil)

	env.IAMServer.UpdateRoleCalls.Returns(&iamv1.UpdateRoleResponse{
		Role: &iamv1.Role{
			Id:   "role-abc",
			Name: "New Name",
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "role", "update", "role-abc", "--name", "New Name")
	require.NoError(t, err)
	assert.Contains(t, stdout, "role-abc")
	assert.Contains(t, stdout, "New Name")
	assert.Contains(t, stdout, "updated")

	req, ok := env.IAMServer.UpdateRoleCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "New Name", req.GetRole().GetName())
	// Permissions should be preserved from fetch.
	require.Len(t, req.GetRole().GetPermissions(), 1)
	assert.Equal(t, "read:clusters", req.GetRole().GetPermissions()[0].GetValue())
}

func TestRoleUpdate_Description(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.GetRoleCalls.Returns(&iamv1.GetRoleResponse{
		Role: &iamv1.Role{
			Id:       "role-abc",
			Name:     "Test",
			RoleType: iamv1.RoleType_ROLE_TYPE_CUSTOM,
			Permissions: []*iamv1.Permission{
				{Value: "read:clusters"},
			},
		},
	}, nil)

	env.IAMServer.UpdateRoleCalls.Returns(&iamv1.UpdateRoleResponse{
		Role: &iamv1.Role{
			Id:   "role-abc",
			Name: "Test",
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "role", "update", "role-abc", "--description", "Updated desc")
	require.NoError(t, err)
	assert.Contains(t, stdout, "updated")

	req, ok := env.IAMServer.UpdateRoleCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "Updated desc", req.GetRole().GetDescription())
}

func TestRoleUpdate_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.GetRoleCalls.Returns(&iamv1.GetRoleResponse{
		Role: &iamv1.Role{
			Id:       "role-abc",
			Name:     "Test",
			RoleType: iamv1.RoleType_ROLE_TYPE_CUSTOM,
			Permissions: []*iamv1.Permission{
				{Value: "read:clusters"},
			},
		},
	}, nil)

	env.IAMServer.UpdateRoleCalls.Returns(&iamv1.UpdateRoleResponse{
		Role: &iamv1.Role{
			Id:   "role-abc",
			Name: "Renamed",
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "role", "update", "role-abc", "--name", "Renamed", "--json")
	require.NoError(t, err)

	var result struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "role-abc", result.ID)
	assert.Equal(t, "Renamed", result.Name)
}

func TestRoleUpdate_BackendError(t *testing.T) {
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

	env.IAMServer.UpdateRoleCalls.Returns(nil, fmt.Errorf("internal server error"))

	_, _, err := testutil.Exec(t, env, "iam", "role", "update", "role-abc", "--name", "X")
	require.Error(t, err)
}

func TestRoleUpdate_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "iam", "role", "update")
	require.Error(t, err)
}
