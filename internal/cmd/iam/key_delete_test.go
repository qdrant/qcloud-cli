package iam_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/auth/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestKeyDelete_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))

	env.AuthServer.DeleteManagementKeyCalls.Returns(&authv1.DeleteManagementKeyResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "key", "delete", "key-abc", "--force")
	require.NoError(t, err)
	assert.Contains(t, stdout, "key-abc")
	assert.Contains(t, stdout, "deleted")

	req, ok := env.AuthServer.DeleteManagementKeyCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "key-abc", req.GetManagementKeyId())
}

func TestKeyDelete_Aborted(t *testing.T) {
	env := testutil.NewTestEnv(t)

	stdout, _, err := testutil.Exec(t, env, "iam", "key", "delete", "key-abc")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Aborted.")
	assert.Equal(t, 0, env.AuthServer.DeleteManagementKeyCalls.Count())
}

func TestKeyDelete_BackendError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AuthServer.DeleteManagementKeyCalls.Returns(nil, fmt.Errorf("internal server error"))

	_, _, err := testutil.Exec(t, env, "iam", "key", "delete", "key-abc", "--force")
	require.Error(t, err)
}

func TestKeyDelete_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "iam", "key", "delete")
	require.Error(t, err)
}

func TestKeyDeleteCompletion(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AuthServer.ListManagementKeysCalls.Returns(&authv1.ListManagementKeysResponse{
		Items: []*authv1.ManagementKey{
			{Id: "key-uuid-1", Prefix: "abc123"},
			{Id: "key-uuid-2", Prefix: "def456"},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "__complete", "iam", "key", "delete", "")
	require.NoError(t, err)
	assert.Contains(t, stdout, "key-uuid-1")
	assert.Contains(t, stdout, "abc123")
	assert.Contains(t, stdout, "key-uuid-2")
	assert.Contains(t, stdout, "def456")
}
