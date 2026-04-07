package iam_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestRoleDelete_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))

	env.IAMServer.DeleteRoleCalls.Returns(&iamv1.DeleteRoleResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "role", "delete", "role-abc", "--force")
	require.NoError(t, err)
	assert.Contains(t, stdout, "role-abc")
	assert.Contains(t, stdout, "deleted")

	req, ok := env.IAMServer.DeleteRoleCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "role-abc", req.GetRoleId())
}

func TestRoleDelete_Aborted(t *testing.T) {
	env := testutil.NewTestEnv(t)

	stdout, _, err := testutil.Exec(t, env, "iam", "role", "delete", "role-abc")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Aborted.")
	assert.Equal(t, 0, env.IAMServer.DeleteRoleCalls.Count())
}

func TestRoleDelete_BackendError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.DeleteRoleCalls.Returns(nil, fmt.Errorf("internal server error"))

	_, _, err := testutil.Exec(t, env, "iam", "role", "delete", "role-abc", "--force")
	require.Error(t, err)
}

func TestRoleDelete_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "iam", "role", "delete")
	require.Error(t, err)
}
