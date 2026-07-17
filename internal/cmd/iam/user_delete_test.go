package iam_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestUserDelete_WithForce_ByEmail(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{{Id: testUserID, Email: "alice@example.com"}},
	}, nil)
	env.IAMServer.DeleteUserCalls.Returns(&iamv1.DeleteUserResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "user", "delete", "alice@example.com", "--force")
	require.NoError(t, err)
	assert.Contains(t, stdout, "alice@example.com")
	assert.Contains(t, stdout, "deleted")

	req, ok := env.IAMServer.DeleteUserCalls.Last()
	require.True(t, ok)
	assert.Equal(t, testUserID, req.GetUserId())
}

func TestUserDelete_ByID(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{{Id: testUserID, Email: "alice@example.com"}},
	}, nil)
	env.IAMServer.DeleteUserCalls.Returns(&iamv1.DeleteUserResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "iam", "user", "delete", testUserID, "--force")
	require.NoError(t, err)

	req, ok := env.IAMServer.DeleteUserCalls.Last()
	require.True(t, ok)
	assert.Equal(t, testUserID, req.GetUserId())
}

func TestUserDelete_Aborted(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{{Id: testUserID, Email: "alice@example.com"}},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "user", "delete", "alice@example.com")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Aborted.")
	assert.Equal(t, 0, env.IAMServer.DeleteUserCalls.Count())
}

func TestUserDelete_ResolveUserError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListUsersCalls.Returns(nil, fmt.Errorf("connection refused"))

	_, _, err := testutil.Exec(t, env, "iam", "user", "delete", "alice@example.com", "--force")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestUserDelete_BackendError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{{Id: testUserID, Email: "alice@example.com"}},
	}, nil)
	env.IAMServer.DeleteUserCalls.Returns(nil, fmt.Errorf("internal server error"))

	_, _, err := testutil.Exec(t, env, "iam", "user", "delete", "alice@example.com", "--force")
	require.Error(t, err)
}

func TestUserDelete_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "iam", "user", "delete")
	require.Error(t, err)
}
