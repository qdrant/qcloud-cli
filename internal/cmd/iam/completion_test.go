package iam_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestUserCompletion(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{
			{Id: "user-uuid-1", Email: "alice@example.com"},
			{Id: "user-uuid-2", Email: "bob@example.com"},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "__complete", "iam", "user", "describe", "")
	require.NoError(t, err)
	assert.Contains(t, stdout, "user-uuid-1")
	assert.Contains(t, stdout, "alice@example.com")
	assert.Contains(t, stdout, "user-uuid-2")
	assert.Contains(t, stdout, "bob@example.com")
}

func TestUserCompletion_StopsAfterFirstArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	stdout, _, err := testutil.Exec(t, env, "__complete", "iam", "user", "describe", "user-uuid-1", "")
	require.NoError(t, err)
	assert.NotContains(t, stdout, "user-uuid")
}

func TestUserThenRoleCompletion_FirstArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{
			{Id: "user-uuid-1", Email: "alice@example.com"},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "__complete", "iam", "user", "assign-role", "")
	require.NoError(t, err)
	assert.Contains(t, stdout, "user-uuid-1")
	assert.Contains(t, stdout, "alice@example.com")
}

