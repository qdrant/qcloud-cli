package cluster_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestSuspend_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))

	env.ClusterServer.SuspendClusterCalls.Returns(&clusterv1.SuspendClusterResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "suspend", "cluster-123", "--force")
	require.NoError(t, err)

	req, ok := env.ClusterServer.SuspendClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "cluster-123", req.GetClusterId())
	assert.Contains(t, stdout, "cluster-123")
	assert.Contains(t, stdout, "suspended")
}

func TestSuspend_MissingArgs(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "cluster", "suspend")
	require.Error(t, err)
}
