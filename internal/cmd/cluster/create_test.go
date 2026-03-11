package cluster_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"
	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestCreateCluster_WithLabels(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	var capturedLabels map[string]string
	env.Server.CreateClusterFunc = func(_ context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error) {
		capturedLabels = make(map[string]string)
		for _, kv := range req.GetCluster().GetLabels() {
			capturedLabels[kv.GetKey()] = kv.GetValue()
		}
		return &clusterv1.CreateClusterResponse{
			Cluster: &clusterv1.Cluster{Id: "cluster-labeled"},
		}, nil
	}

	_, _, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--label", "env=prod",
		"--label", "team=platform",
	)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"env": "prod", "team": "platform"}, capturedLabels)
}

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

func TestCreateCluster_PackageByUUID(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	var listCallCount int32
	env.BookingServer.ListPackagesFunc = func(_ context.Context, _ *bookingv1.ListPackagesRequest) (*bookingv1.ListPackagesResponse, error) {
		atomic.AddInt32(&listCallCount, 1)
		return &bookingv1.ListPackagesResponse{}, nil
	}

	var capturedPackageID string
	env.Server.CreateClusterFunc = func(_ context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error) {
		capturedPackageID = req.GetCluster().GetConfiguration().GetPackageId()
		return &clusterv1.CreateClusterResponse{
			Cluster: &clusterv1.Cluster{Id: "cluster-pkg-uuid"},
		}, nil
	}

	_, _, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--package", "550e8400-e29b-41d4-a716-446655440000",
	)
	require.NoError(t, err)
	assert.EqualValues(t, 0, atomic.LoadInt32(&listCallCount), "ListPackages should not be called for UUID input")
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", capturedPackageID)
}

func TestCreateCluster_PackageByName(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.BookingServer.ListPackagesFunc = func(_ context.Context, _ *bookingv1.ListPackagesRequest) (*bookingv1.ListPackagesResponse, error) {
		return &bookingv1.ListPackagesResponse{
			Items: []*bookingv1.Package{
				{Id: "pkg-uuid-123", Name: "starter"},
			},
		}, nil
	}

	var capturedPackageID string
	env.Server.CreateClusterFunc = func(_ context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error) {
		capturedPackageID = req.GetCluster().GetConfiguration().GetPackageId()
		return &clusterv1.CreateClusterResponse{
			Cluster: &clusterv1.Cluster{Id: "cluster-named-pkg"},
		}, nil
	}

	_, _, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--package", "starter",
	)
	require.NoError(t, err)
	assert.Equal(t, "pkg-uuid-123", capturedPackageID)
}

func TestCreateCluster_PackageNameNotFound(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.BookingServer.ListPackagesFunc = func(_ context.Context, _ *bookingv1.ListPackagesRequest) (*bookingv1.ListPackagesResponse, error) {
		return &bookingv1.ListPackagesResponse{}, nil
	}

	_, _, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--package", "starter",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "starter")
}
