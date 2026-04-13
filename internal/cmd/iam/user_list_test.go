package iam_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestUserList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{
			{
				Id:     "user-1",
				Email:  "alice@example.com",
				Status: iamv1.UserStatus_USER_STATUS_ACTIVE,
			},
			{
				Id:     "user-2",
				Email:  "bob@example.com",
				Status: iamv1.UserStatus_USER_STATUS_BLOCKED,
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "user", "list")
	require.NoError(t, err)

	assert.Contains(t, stdout, "user-1")
	assert.Contains(t, stdout, "alice@example.com")
	assert.Contains(t, stdout, "ACTIVE")
	assert.Contains(t, stdout, "user-2")
	assert.Contains(t, stdout, "bob@example.com")
	assert.Contains(t, stdout, "BLOCKED")

	req, ok := env.IAMServer.ListUsersCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
}

func TestUserList_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{{Id: "user-1", Email: "alice@example.com"}},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "user", "list", "--json")
	require.NoError(t, err)

	assert.Contains(t, stdout, `"id"`)
	assert.Contains(t, stdout, "user-1")
}

func TestUserList_Error(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListUsersCalls.Returns(nil, fmt.Errorf("permission denied"))

	_, _, err := testutil.Exec(t, env, "iam", "user", "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestUserList_NoHeaders(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{{Id: "user-1", Email: "alice@example.com"}},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "user", "list", "--no-headers")
	require.NoError(t, err)
	assert.NotContains(t, stdout, "ID")
	assert.NotContains(t, stdout, "EMAIL")
	assert.Contains(t, stdout, "user-1")
}
