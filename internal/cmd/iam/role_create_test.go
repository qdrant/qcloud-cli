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

func TestRoleCreate_Success(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))

	env.IAMServer.CreateRoleCalls.Returns(&iamv1.CreateRoleResponse{
		Role: &iamv1.Role{
			Id:   "new-role-id",
			Name: "Cluster Viewer",
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "role", "create",
		"--name", "Cluster Viewer",
		"--description", "Read-only cluster access",
		"--permission", "read:clusters",
		"--permission", "read:backups",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "new-role-id")
	assert.Contains(t, stdout, "Cluster Viewer")
	assert.Contains(t, stdout, "created")

	req, ok := env.IAMServer.CreateRoleCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetRole().GetAccountId())
	assert.Equal(t, "Cluster Viewer", req.GetRole().GetName())
	assert.Equal(t, "Read-only cluster access", req.GetRole().GetDescription())
	assert.Equal(t, iamv1.RoleType_ROLE_TYPE_CUSTOM, req.GetRole().GetRoleType())
	require.Len(t, req.GetRole().GetPermissions(), 2)
	assert.Equal(t, "read:clusters", req.GetRole().GetPermissions()[0].GetValue())
	assert.Equal(t, "read:backups", req.GetRole().GetPermissions()[1].GetValue())
}

func TestRoleCreate_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.CreateRoleCalls.Returns(&iamv1.CreateRoleResponse{
		Role: &iamv1.Role{
			Id:   "json-role-id",
			Name: "JSON Role",
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "role", "create",
		"--name", "JSON Role",
		"--permission", "read:clusters",
		"--json",
	)
	require.NoError(t, err)

	var result struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "json-role-id", result.ID)
	assert.Equal(t, "JSON Role", result.Name)
}

func TestRoleCreate_BackendError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.CreateRoleCalls.Returns(nil, fmt.Errorf("permission denied"))

	_, _, err := testutil.Exec(t, env, "iam", "role", "create",
		"--name", "Test",
		"--permission", "read:clusters",
	)
	require.Error(t, err)
}

func TestRoleCreate_MissingName(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "iam", "role", "create",
		"--permission", "read:clusters",
	)
	require.Error(t, err)
}

func TestRoleCreate_MissingPermission(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "iam", "role", "create",
		"--name", "Test",
	)
	require.Error(t, err)
}
