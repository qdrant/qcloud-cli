package hybrid_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestHybridClusterDescribe_FullOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	version := "1.8.0"
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{
			Id:                    "cluster-abc",
			Name:                  "my-cluster",
			CloudProviderRegionId: "env-123",
			State:                 &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY},
			Configuration: &clusterv1.ClusterConfiguration{
				Version:       &version,
				NumberOfNodes: 3,
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "cluster", "describe", "cluster-abc")
	require.NoError(t, err)

	assert.Contains(t, stdout, "cluster-abc")
	assert.Contains(t, stdout, "my-cluster")
	assert.Contains(t, stdout, "HEALTHY")
	assert.Contains(t, stdout, "env-123")
	assert.Contains(t, stdout, "1.8.0")
	assert.Contains(t, stdout, "3")
}

func TestHybridClusterDescribe_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "hybrid", "cluster", "describe")
	require.Error(t, err)
}

func TestHybridClusterDescribe_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Returns(nil, fmt.Errorf("not found"))

	_, _, err := testutil.Exec(t, env, "hybrid", "cluster", "describe", "cluster-abc")
	require.Error(t, err)
}
