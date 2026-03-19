package cluster_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestRestart_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))

	env.ClusterServer.RestartClusterCalls.Returns(&clusterv1.RestartClusterResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "restart", "cluster-123", "--force")
	require.NoError(t, err)
	assert.Contains(t, stdout, "cluster-123")
	assert.Contains(t, stdout, "restarting")

	req, ok := env.ClusterServer.RestartClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "cluster-123", req.GetClusterId())
}

func TestRestart_MissingArgs(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "cluster", "restart")
	require.Error(t, err)
}

func TestRestart_NoWait(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ClusterServer.RestartClusterCalls.Returns(&clusterv1.RestartClusterResponse{}, nil)
	env.ClusterServer.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{
			State: &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_UPDATING},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "restart", "cluster-123", "--force")
	require.NoError(t, err)
	assert.Contains(t, stdout, "restarting")
	assert.Equal(t, 0, env.ClusterServer.GetClusterCalls.Count(), "GetCluster should not be called without --wait")
}

func TestRestart_WaitSuccess(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ClusterServer.RestartClusterCalls.Returns(&clusterv1.RestartClusterResponse{}, nil)
	env.ClusterServer.GetClusterCalls.
		OnCall(0, func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
			return &clusterv1.GetClusterResponse{
				Cluster: &clusterv1.Cluster{
					Id:    "cluster-123",
					State: &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_UPDATING},
				},
			}, nil
		}).
		OnCall(1, func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
			return &clusterv1.GetClusterResponse{
				Cluster: &clusterv1.Cluster{
					Id:    "cluster-123",
					State: &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_UPDATING},
				},
			}, nil
		}).
		Always(func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
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
		})

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

	env.ClusterServer.RestartClusterCalls.Returns(&clusterv1.RestartClusterResponse{}, nil)
	env.ClusterServer.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{
			Id: "cluster-123",
			State: &clusterv1.ClusterState{
				Phase:  clusterv1.ClusterPhase_CLUSTER_PHASE_FAILED_TO_SYNC,
				Reason: "sync failed",
			},
		},
	}, nil)

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

	env.ClusterServer.RestartClusterCalls.Returns(&clusterv1.RestartClusterResponse{}, nil)
	env.ClusterServer.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{
			Id:    "cluster-123",
			State: &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_UPDATING},
		},
	}, nil)

	_, _, err := testutil.Exec(t, env,
		"cluster", "restart", "cluster-123", "--force",
		"--wait",
		"--wait-timeout", "50ms",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
}
