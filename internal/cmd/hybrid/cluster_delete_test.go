package hybrid_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestHybridClusterDelete_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.DeleteClusterCalls.Returns(&clusterv1.DeleteClusterResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "cluster", "delete", "cluster-abc", "--force")
	require.NoError(t, err)

	req, ok := env.Server.DeleteClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "cluster-abc", req.GetClusterId())
	assert.Contains(t, stdout, "cluster-abc")
	assert.Contains(t, stdout, "deleted")
}

func TestHybridClusterDelete_WithoutForce(t *testing.T) {
	env := testutil.NewTestEnv(t)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "cluster", "delete", "cluster-abc")
	require.NoError(t, err)

	assert.Contains(t, stdout, "Aborted.")
	assert.Equal(t, 0, env.Server.DeleteClusterCalls.Count())
}

func TestHybridClusterDelete_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "hybrid", "cluster", "delete")
	require.Error(t, err)
}

func TestHybridClusterDelete_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.DeleteClusterCalls.Returns(nil, fmt.Errorf("delete failed"))

	_, _, err := testutil.Exec(t, env, "hybrid", "cluster", "delete", "cluster-abc", "--force")
	require.Error(t, err)
}
