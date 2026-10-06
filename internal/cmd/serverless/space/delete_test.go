package space_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spacev1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestSpaceDelete_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.DeleteSpaceCalls.Returns(&spacev1.DeleteSpaceResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "delete", "space-abc", "--force")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Space space-abc deleted.")

	req, ok := env.ServerlessSpaceServer.DeleteSpaceCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "space-abc", req.GetSpaceId())
	assert.Nil(t, req.DeleteBackups)
}

func TestSpaceDelete_DeleteBackups(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.DeleteSpaceCalls.Returns(&spacev1.DeleteSpaceResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "delete", "space-abc", "--delete-backups", "-f")
	require.NoError(t, err)

	req, ok := env.ServerlessSpaceServer.DeleteSpaceCalls.Last()
	require.True(t, ok)
	require.NotNil(t, req.DeleteBackups)
	assert.True(t, req.GetDeleteBackups())
}

func TestSpaceDelete_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.DeleteSpaceCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "delete", "space-abc", "--force")
	require.Error(t, err)
}

func TestSpaceDelete_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "delete", "--force")
	require.Error(t, err)
	assert.Equal(t, 0, env.ServerlessSpaceServer.DeleteSpaceCalls.Count())
}
