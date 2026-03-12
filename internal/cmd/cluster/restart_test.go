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

func TestRestart_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))
	t.Cleanup(env.Cleanup)

	var capturedReq *clusterv1.RestartClusterRequest
	env.Server.RestartClusterFunc = func(_ context.Context, req *clusterv1.RestartClusterRequest) (*clusterv1.RestartClusterResponse, error) {
		capturedReq = req
		return &clusterv1.RestartClusterResponse{}, nil
	}

	stdout, _, err := testutil.Exec(t, env, "cluster", "restart", "cluster-123", "--force")
	require.NoError(t, err)
	assert.Equal(t, "test-account-id", capturedReq.GetAccountId())
	assert.Equal(t, "cluster-123", capturedReq.GetClusterId())
	assert.Contains(t, stdout, "cluster-123")
	assert.Contains(t, stdout, "restarting")
}

func TestRestart_MissingArgs(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	_, _, err := testutil.Exec(t, env, "cluster", "restart")
	require.Error(t, err)
}

func TestRestart_NoWait(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	var getCallCount int32
	env.Server.RestartClusterFunc = func(_ context.Context, _ *clusterv1.RestartClusterRequest) (*clusterv1.RestartClusterResponse, error) {
		return &clusterv1.RestartClusterResponse{}, nil
	}
	env.Server.GetClusterFunc = func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		atomic.AddInt32(&getCallCount, 1)
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{
				State: &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_UPDATING},
			},
		}, nil
	}

	stdout, _, err := testutil.Exec(t, env, "cluster", "restart", "cluster-123", "--force")
	require.NoError(t, err)
	assert.Contains(t, stdout, "restarting")
	assert.EqualValues(t, 0, atomic.LoadInt32(&getCallCount), "GetCluster should not be called without --wait")
}

func TestRestart_WaitSuccess(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.Server.RestartClusterFunc = func(_ context.Context, _ *clusterv1.RestartClusterRequest) (*clusterv1.RestartClusterResponse, error) {
		return &clusterv1.RestartClusterResponse{}, nil
	}

	var callCount int32
	env.Server.GetClusterFunc = func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		n := atomic.AddInt32(&callCount, 1)
		if n < 3 {
			return &clusterv1.GetClusterResponse{
				Cluster: &clusterv1.Cluster{
					Id:    "cluster-123",
					State: &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_UPDATING},
				},
			}, nil
		}
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id:   "cluster-123",
				Name: "my-cluster",
				State: &clusterv1.ClusterState{
					Phase:    clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY,
					Endpoint: &clusterv1.ClusterEndpoint{Url: "https://cluster-123.aws.cloud.qdrant.io"},
				},
			},
		}, nil
	}

	stdout, stderr, err := testutil.Exec(t, env,
		"cluster", "restart", "cluster-123", "--force",
		"--wait",
		"--wait-timeout", "30s",
		"--wait-poll-interval", "10ms",
	)
	require.NoError(t, err)
	assert.Contains(t, stderr, "phase=UPDATING")
	assert.Contains(t, stdout, "cluster-123")
	assert.Contains(t, stdout, "ready")
	assert.Contains(t, stdout, "https://cluster-123.aws.cloud.qdrant.io")
}

func TestRestart_WaitFailure(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.Server.RestartClusterFunc = func(_ context.Context, _ *clusterv1.RestartClusterRequest) (*clusterv1.RestartClusterResponse, error) {
		return &clusterv1.RestartClusterResponse{}, nil
	}
	env.Server.GetClusterFunc = func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id: "cluster-123",
				State: &clusterv1.ClusterState{
					Phase:  clusterv1.ClusterPhase_CLUSTER_PHASE_FAILED_TO_SYNC,
					Reason: "sync failed",
				},
			},
		}, nil
	}

	_, _, err := testutil.Exec(t, env,
		"cluster", "restart", "cluster-123", "--force",
		"--wait",
		"--wait-timeout", "30s",
		"--wait-poll-interval", "10ms",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "FAILED_TO_SYNC")
	assert.Contains(t, err.Error(), "sync failed")
}

func TestRestart_WaitTimeout(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.Server.RestartClusterFunc = func(_ context.Context, _ *clusterv1.RestartClusterRequest) (*clusterv1.RestartClusterResponse, error) {
		return &clusterv1.RestartClusterResponse{}, nil
	}
	env.Server.GetClusterFunc = func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id:    "cluster-123",
				State: &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_UPDATING},
			},
		}, nil
	}

	_, _, err := testutil.Exec(t, env,
		"cluster", "restart", "cluster-123", "--force",
		"--wait",
		"--wait-timeout", "50ms",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
}
