package hybrid_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func setupHybridClusterUpdateHandlers(env *testutil.TestEnv) {
	env.Server.GetClusterCalls.Always(func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{Id: req.GetClusterId(), Name: "my-cluster"},
		}, nil
	})
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		return &clusterv1.UpdateClusterResponse{Cluster: req.GetCluster()}, nil
	})
}

func TestHybridClusterUpdate_Label(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupHybridClusterUpdateHandlers(env)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "cluster", "update", "cluster-abc", "--label", "env=prod")
	require.NoError(t, err)
	assert.Contains(t, stdout, "updated")

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	labels := make(map[string]string)
	for _, kv := range req.GetCluster().GetLabels() {
		labels[kv.GetKey()] = kv.GetValue()
	}
	assert.Equal(t, "prod", labels["env"])
}

func TestHybridClusterUpdate_ServiceType(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupHybridClusterUpdateHandlers(env)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "cluster", "update", "cluster-abc", "--service-type", "node-port")
	require.NoError(t, err)
	assert.Contains(t, stdout, "updated")

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, clusterv1.ClusterServiceType_CLUSTER_SERVICE_TYPE_NODE_PORT, req.GetCluster().GetConfiguration().GetServiceType())
}

func TestHybridClusterUpdate_DBConfig_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupHybridClusterUpdateHandlers(env)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "cluster", "update", "cluster-abc",
		"--replication-factor", "2",
		"--force",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "updated")

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	rf := req.GetCluster().GetConfiguration().GetDatabaseConfiguration().GetCollection().GetReplicationFactor()
	assert.Equal(t, uint32(2), rf)
}

func TestHybridClusterUpdate_DBConfig_WithoutForce(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupHybridClusterUpdateHandlers(env)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "cluster", "update", "cluster-abc",
		"--replication-factor", "2",
	)
	require.NoError(t, err)

	assert.Contains(t, stdout, "Aborted.")
	assert.Equal(t, 0, env.Server.UpdateClusterCalls.Count())
}

func TestHybridClusterUpdate_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "hybrid", "cluster", "update")
	require.Error(t, err)
}

func TestHybridClusterUpdate_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{Id: "cluster-abc", Name: "my-cluster"},
	}, nil)
	env.Server.UpdateClusterCalls.Returns(nil, fmt.Errorf("internal server error"))

	_, _, err := testutil.Exec(t, env, "hybrid", "cluster", "update", "cluster-abc", "--label", "env=prod")
	require.Error(t, err)
}
