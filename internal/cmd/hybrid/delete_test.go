package hybrid_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hybridv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/hybrid/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestHybridDelete_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.HybridServer.DeleteHybridCloudEnvironmentCalls.Returns(&hybridv1.DeleteHybridCloudEnvironmentResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "delete", "env-abc", "--force")
	require.NoError(t, err)

	req, ok := env.HybridServer.DeleteHybridCloudEnvironmentCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "env-abc", req.GetHybridCloudEnvironmentId())
	assert.Contains(t, stdout, "env-abc")
	assert.Contains(t, stdout, "deleted")
}

func TestHybridDelete_WithoutForce(t *testing.T) {
	env := testutil.NewTestEnv(t)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "delete", "env-abc")
	require.NoError(t, err)

	assert.Contains(t, stdout, "Aborted.")
	assert.Equal(t, 0, env.HybridServer.DeleteHybridCloudEnvironmentCalls.Count())
}

func TestHybridDelete_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "hybrid", "delete")
	require.Error(t, err)
}

func TestHybridDelete_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.HybridServer.DeleteHybridCloudEnvironmentCalls.Returns(nil, fmt.Errorf("delete failed"))

	_, _, err := testutil.Exec(t, env, "hybrid", "delete", "env-abc", "--force")
	require.Error(t, err)
}
