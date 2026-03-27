package hybrid_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestHybridClusterCreate_Basic(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.CreateClusterCalls.Returns(&clusterv1.CreateClusterResponse{
		Cluster: &clusterv1.Cluster{Id: "cluster-new", Name: "my-cluster"},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "cluster", "create", "env-123", "--name", "my-cluster")
	require.NoError(t, err)

	assert.Contains(t, stdout, "cluster-new")
	assert.Contains(t, stdout, "my-cluster")

	req, ok := env.Server.CreateClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "hybrid", req.GetCluster().GetCloudProviderId())
	assert.Equal(t, "env-123", req.GetCluster().GetCloudProviderRegionId())
	assert.Equal(t, "test-account-id", req.GetCluster().GetAccountId())
}

func TestHybridClusterCreate_AutoName(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.SuggestClusterNameCalls.Returns(&clusterv1.SuggestClusterNameResponse{Name: "eager-pelican"}, nil)
	env.Server.CreateClusterCalls.Always(func(_ context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error) {
		return &clusterv1.CreateClusterResponse{
			Cluster: &clusterv1.Cluster{Id: "cluster-auto", Name: req.GetCluster().GetName()},
		}, nil
	})

	_, _, err := testutil.Exec(t, env, "hybrid", "cluster", "create", "env-123")
	require.NoError(t, err)

	assert.Equal(t, 1, env.Server.SuggestClusterNameCalls.Count())

	req, ok := env.Server.CreateClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "eager-pelican", req.GetCluster().GetName())
}

func TestHybridClusterCreate_WithFlags(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.CreateClusterCalls.Returns(&clusterv1.CreateClusterResponse{
		Cluster: &clusterv1.Cluster{Id: "cluster-flags", Name: "flagged"},
	}, nil)

	_, _, err := testutil.Exec(t, env, "hybrid", "cluster", "create", "env-123",
		"--name", "flagged",
		"--nodes", "3",
		"--version", "1.8.0",
		"--service-type", "load-balancer",
		"--label", "env=prod",
		"--node-selector", "zone=us-east",
	)
	require.NoError(t, err)

	req, ok := env.Server.CreateClusterCalls.Last()
	require.True(t, ok)
	cfg := req.GetCluster().GetConfiguration()
	assert.Equal(t, uint32(3), cfg.GetNumberOfNodes())
	assert.Equal(t, "1.8.0", cfg.GetVersion())
	assert.Equal(t, clusterv1.ClusterServiceType_CLUSTER_SERVICE_TYPE_LOAD_BALANCER, cfg.GetServiceType())

	labels := make(map[string]string)
	for _, kv := range req.GetCluster().GetLabels() {
		labels[kv.GetKey()] = kv.GetValue()
	}
	assert.Equal(t, "prod", labels["env"])

	nodeSelectors := make(map[string]string)
	for _, kv := range cfg.GetNodeSelector() {
		nodeSelectors[kv.GetKey()] = kv.GetValue()
	}
	assert.Equal(t, "us-east", nodeSelectors["zone"])
}

func TestHybridClusterCreate_NoWait(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.CreateClusterCalls.Returns(&clusterv1.CreateClusterResponse{
		Cluster: &clusterv1.Cluster{Id: "cluster-nowait", Name: "nowait"},
	}, nil)
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{State: &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_CREATING}},
	}, nil)

	_, _, err := testutil.Exec(t, env, "hybrid", "cluster", "create", "env-123", "--name", "nowait")
	require.NoError(t, err)

	assert.Equal(t, 0, env.Server.GetClusterCalls.Count(), "GetCluster should not be called without --wait")
}

func TestHybridClusterCreate_WithWait(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.SuggestClusterNameCalls.Returns(&clusterv1.SuggestClusterNameResponse{Name: "eager-pelican"}, nil)
	env.Server.CreateClusterCalls.Always(func(_ context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error) {
		return &clusterv1.CreateClusterResponse{
			Cluster: &clusterv1.Cluster{Id: "cluster-wait", Name: req.GetCluster().GetName()},
		}, nil
	})
	env.Server.GetClusterCalls.Always(func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id:    "cluster-wait",
				State: &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY},
			},
		}, nil
	})

	stdout, _, err := testutil.Exec(t, env, "hybrid", "cluster", "create", "env-123",
		"--wait",
		"--wait-timeout", "30s",
		"--wait-poll-interval", "10ms",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "cluster-wait")
}
