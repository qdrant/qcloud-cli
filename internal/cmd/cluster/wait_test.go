package cluster_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestWaitCluster_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.
		OnCall(0, func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
			return &clusterv1.GetClusterResponse{
				Cluster: &clusterv1.Cluster{
					Id:    "cluster-abc",
					State: &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_CREATING},
				},
			}, nil
		}).
		OnCall(1, func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
			return &clusterv1.GetClusterResponse{
				Cluster: &clusterv1.Cluster{
					Id:    "cluster-abc",
					State: &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_CREATING},
				},
			}, nil
		}).
		Always(func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
			return &clusterv1.GetClusterResponse{
				Cluster: &clusterv1.Cluster{
					Id:   "cluster-abc",
					Name: "my-cluster",
					State: &clusterv1.ClusterState{
						Phase:    clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY,
						Endpoint: &clusterv1.ClusterEndpoint{Url: "https://abc.aws.cloud.qdrant.io"},
					},
				},
			}, nil
		})

	stdout, stderr, err := testutil.Exec(t, env,
		"cluster", "wait", "cluster-abc",
		"--timeout", "30s",
		"--poll-interval", "10ms",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "cluster-abc")
	assert.Contains(t, stdout, "https://abc.aws.cloud.qdrant.io")
	assert.Contains(t, stderr, "phase=CREATING")
}

func TestWaitCluster_Failure(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{
			Id: "cluster-fail",
			State: &clusterv1.ClusterState{
				Phase:  clusterv1.ClusterPhase_CLUSTER_PHASE_FAILED_TO_CREATE,
				Reason: "quota exceeded",
			},
		},
	}, nil)

	_, _, err := testutil.Exec(t, env,
		"cluster", "wait", "cluster-fail",
		"--timeout", "30s",
		"--poll-interval", "10ms",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "FAILED_TO_CREATE")
	assert.Contains(t, err.Error(), "quota exceeded")
}

func TestWaitCluster_Timeout(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{
			Id:    "cluster-slow",
			State: &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_CREATING},
		},
	}, nil)

	_, _, err := testutil.Exec(t, env,
		"cluster", "wait", "cluster-slow",
		"--timeout", "200ms",
		"--poll-interval", "10ms",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
}
