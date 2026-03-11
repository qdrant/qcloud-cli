package cluster_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestCreateCluster_NoWait(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	var getCallCount int32
	env.Server.CreateClusterFunc = func(_ context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error) {
		return &clusterv1.CreateClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id:   "cluster-abc",
				Name: req.GetCluster().GetName(),
			},
		}, nil
	}
	env.Server.GetClusterFunc = func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		atomic.AddInt32(&getCallCount, 1)
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{
				State: &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_CREATING},
			},
		}, nil
	}

	stdout, _, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "cluster-abc")
	assert.EqualValues(t, 0, atomic.LoadInt32(&getCallCount), "GetCluster should not be called without --wait")
}

func TestCreateCluster_WaitSuccess(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.Server.CreateClusterFunc = func(_ context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error) {
		return &clusterv1.CreateClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id:   "cluster-xyz",
				Name: req.GetCluster().GetName(),
			},
		}, nil
	}

	var callCount int32
	env.Server.GetClusterFunc = func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		n := atomic.AddInt32(&callCount, 1)
		if n < 3 {
			return &clusterv1.GetClusterResponse{
				Cluster: &clusterv1.Cluster{
					Id:    "cluster-xyz",
					State: &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_CREATING},
				},
			}, nil
		}
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id: "cluster-xyz",
				State: &clusterv1.ClusterState{
					Phase:    clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY,
					Endpoint: &clusterv1.ClusterEndpoint{Url: "https://xyz.aws.cloud.qdrant.io"},
				},
			},
		}, nil
	}

	stdout, stderr, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--wait",
		"--wait-timeout", "30s",
		"--wait-poll-interval", "10ms",
	)
	require.NoError(t, err)
	assert.Contains(t, stderr, "phase=CREATING")
	assert.Contains(t, stdout, "cluster-xyz")
	assert.Contains(t, stdout, "ready")
	assert.Contains(t, stdout, "https://xyz.aws.cloud.qdrant.io")
}

func TestCreateCluster_WaitFailure(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.Server.CreateClusterFunc = func(_ context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error) {
		return &clusterv1.CreateClusterResponse{
			Cluster: &clusterv1.Cluster{Id: "cluster-fail"},
		}, nil
	}
	env.Server.GetClusterFunc = func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id: "cluster-fail",
				State: &clusterv1.ClusterState{
					Phase:  clusterv1.ClusterPhase_CLUSTER_PHASE_FAILED_TO_CREATE,
					Reason: "quota exceeded",
				},
			},
		}, nil
	}

	_, _, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--wait",
		"--wait-timeout", "30s",
		"--wait-poll-interval", "10ms",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "FAILED_TO_CREATE")
	assert.Contains(t, err.Error(), "quota exceeded")
}

func TestCreateCluster_WaitTimeout(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.Server.CreateClusterFunc = func(_ context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error) {
		return &clusterv1.CreateClusterResponse{
			Cluster: &clusterv1.Cluster{Id: "cluster-slow"},
		}, nil
	}
	env.Server.GetClusterFunc = func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id:    "cluster-slow",
				State: &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_CREATING},
			},
		}, nil
	}

	_, _, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--wait",
		"--wait-timeout", "50ms",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
}
