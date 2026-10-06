package space_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spaceauthv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/auth/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestSpaceKeyDelete_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceApiKeyServer.DeleteSpaceApiKeyCalls.Returns(&spaceauthv1.DeleteSpaceApiKeyResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "key", "delete", "space-abc", "key-123", "--force")
	require.NoError(t, err)
	assert.Contains(t, stdout, "API key key-123 deleted.")

	req, ok := env.ServerlessSpaceApiKeyServer.DeleteSpaceApiKeyCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "space-abc", req.GetSpaceId())
	assert.Equal(t, "key-123", req.GetSpaceApiKeyId())
}

func TestSpaceKeyDelete_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceApiKeyServer.DeleteSpaceApiKeyCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "key", "delete", "space-abc", "key-123", "--force")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete API key")
}

func TestSpaceKeyDelete_MissingArgs(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "key", "delete", "space-abc")
	require.Error(t, err)
	assert.Equal(t, 0, env.ServerlessSpaceApiKeyServer.DeleteSpaceApiKeyCalls.Count())
}
