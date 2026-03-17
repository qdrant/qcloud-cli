package cluster_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"
	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestCreateCluster_WithLabels(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.CreateClusterCalls.Returns(&clusterv1.CreateClusterResponse{
		Cluster: &clusterv1.Cluster{Id: "cluster-labeled"},
	}, nil)

	_, _, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--package", "00000000-0000-0000-0000-000000000001",
		"--label", "env=prod",
		"--label", "team=platform",
	)
	require.NoError(t, err)

	req, ok := env.Server.CreateClusterCalls.Last()
	require.True(t, ok)
	capturedLabels := make(map[string]string)
	for _, kv := range req.GetCluster().GetLabels() {
		capturedLabels[kv.GetKey()] = kv.GetValue()
	}
	assert.Equal(t, map[string]string{"env": "prod", "team": "platform"}, capturedLabels)
}

func TestCreateCluster_NoWait(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.CreateClusterCalls.Always(func(_ context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error) {
		return &clusterv1.CreateClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id:   "cluster-abc",
				Name: req.GetCluster().GetName(),
			},
		}, nil
	})
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{
			State: &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_CREATING},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--package", "00000000-0000-0000-0000-000000000001",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "cluster-abc")
	assert.Equal(t, 0, env.Server.GetClusterCalls.Count(), "GetCluster should not be called without --wait")
}

func TestCreateCluster_WaitSuccess(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.CreateClusterCalls.Always(func(_ context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error) {
		return &clusterv1.CreateClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id:   "cluster-xyz",
				Name: req.GetCluster().GetName(),
			},
		}, nil
	})
	env.Server.GetClusterCalls.
		OnCall(0, func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
			return &clusterv1.GetClusterResponse{
				Cluster: &clusterv1.Cluster{
					Id:    "cluster-xyz",
					State: &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_CREATING},
				},
			}, nil
		}).
		OnCall(1, func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
			return &clusterv1.GetClusterResponse{
				Cluster: &clusterv1.Cluster{
					Id:    "cluster-xyz",
					State: &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_CREATING},
				},
			}, nil
		}).
		Always(func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
			return &clusterv1.GetClusterResponse{
				Cluster: &clusterv1.Cluster{
					Id: "cluster-xyz",
					State: &clusterv1.ClusterState{
						Phase:    clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY,
						Endpoint: &clusterv1.ClusterEndpoint{Url: "https://xyz.aws.cloud.qdrant.io"},
					},
				},
			}, nil
		})

	stdout, stderr, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--package", "00000000-0000-0000-0000-000000000001",
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

	env.Server.CreateClusterCalls.Returns(&clusterv1.CreateClusterResponse{
		Cluster: &clusterv1.Cluster{Id: "cluster-fail"},
	}, nil)
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
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--package", "00000000-0000-0000-0000-000000000001",
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

	env.Server.CreateClusterCalls.Returns(&clusterv1.CreateClusterResponse{
		Cluster: &clusterv1.Cluster{Id: "cluster-slow"},
	}, nil)
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{
			Id:    "cluster-slow",
			State: &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_CREATING},
		},
	}, nil)

	_, _, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--package", "00000000-0000-0000-0000-000000000001",
		"--wait",
		"--wait-timeout", "50ms",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
}

func TestCreateCluster_PackageByUUID(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{}, nil)
	env.Server.CreateClusterCalls.Returns(&clusterv1.CreateClusterResponse{
		Cluster: &clusterv1.Cluster{Id: "cluster-pkg-uuid"},
	}, nil)

	_, _, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--package", "550e8400-e29b-41d4-a716-446655440000",
	)
	require.NoError(t, err)
	assert.Equal(t, 0, env.BookingServer.ListPackagesCalls.Count(), "ListPackages should not be called for UUID input")

	req, ok := env.Server.CreateClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", req.GetCluster().GetConfiguration().GetPackageId())
}

func TestCreateCluster_PackageByName(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{
		Items: []*bookingv1.Package{
			{Id: "pkg-uuid-123", Name: "starter"},
		},
	}, nil)
	env.Server.CreateClusterCalls.Returns(&clusterv1.CreateClusterResponse{
		Cluster: &clusterv1.Cluster{Id: "cluster-named-pkg"},
	}, nil)

	_, _, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--package", "starter",
	)
	require.NoError(t, err)

	req, ok := env.Server.CreateClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "pkg-uuid-123", req.GetCluster().GetConfiguration().GetPackageId())
}

func TestCreateCluster_PackageNameNotFound(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{}, nil)

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

func TestCreateCluster_AutoGeneratedName(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.SuggestClusterNameCalls.Returns(&clusterv1.SuggestClusterNameResponse{Name: "eager-pelican"}, nil)
	env.Server.CreateClusterCalls.Returns(&clusterv1.CreateClusterResponse{
		Cluster: &clusterv1.Cluster{Id: "cluster-auto", Name: "eager-pelican"},
	}, nil)

	_, _, err := testutil.Exec(t, env,
		"cluster", "create",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--package", "00000000-0000-0000-0000-000000000001",
	)
	require.NoError(t, err)

	req, ok := env.Server.CreateClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "eager-pelican", req.GetCluster().GetName())
}

func TestCreateCluster_PackageByResources(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{
		Items: []*bookingv1.Package{
			{
				Id:   "pkg-res-1",
				Name: "starter",
				ResourceConfiguration: &bookingv1.ResourceConfiguration{
					Cpu:  "1000m",
					Ram:  "1GiB",
					Disk: "100GiB",
				},
			},
		},
	}, nil)
	env.Server.CreateClusterCalls.Returns(&clusterv1.CreateClusterResponse{
		Cluster: &clusterv1.Cluster{Id: "cluster-by-resources"},
	}, nil)

	_, _, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--cpu", "1000m",
		"--ram", "1GiB",
		"--disk", "100GiB",
	)
	require.NoError(t, err)

	req, ok := env.Server.CreateClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "pkg-res-1", req.GetCluster().GetConfiguration().GetPackageId())
	assert.Nil(t, req.GetCluster().GetConfiguration().GetAdditionalResources(), "no additional disk expected when requested == package disk")
}

func TestCreateCluster_PackageByResourcesPartial(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{
		Items: []*bookingv1.Package{
			{
				Id:   "pkg-cpu-only",
				Name: "starter",
				ResourceConfiguration: &bookingv1.ResourceConfiguration{
					Cpu:  "500m",
					Ram:  "512MiB",
					Disk: "50GiB",
				},
			},
		},
	}, nil)
	env.Server.CreateClusterCalls.Returns(&clusterv1.CreateClusterResponse{
		Cluster: &clusterv1.Cluster{Id: "cluster-cpu-only"},
	}, nil)

	_, _, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--cpu", "500m",
	)
	require.NoError(t, err)

	req, ok := env.Server.CreateClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "pkg-cpu-only", req.GetCluster().GetConfiguration().GetPackageId())
}

func TestCreateCluster_PackageByResourcesNoMatch(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{
		Items: []*bookingv1.Package{
			{
				Id:   "pkg-other",
				Name: "starter",
				ResourceConfiguration: &bookingv1.ResourceConfiguration{
					Cpu: "500m",
				},
			},
		},
	}, nil)

	_, _, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--cpu", "9999m",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no package found")
}

func TestCreateCluster_PackageByResourcesAmbiguous(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{
		Items: []*bookingv1.Package{
			{
				Id:   "pkg-a",
				Name: "starter-a",
				ResourceConfiguration: &bookingv1.ResourceConfiguration{
					Cpu: "1000m",
				},
			},
			{
				Id:   "pkg-b",
				Name: "starter-b",
				ResourceConfiguration: &bookingv1.ResourceConfiguration{
					Cpu: "1000m",
				},
			},
		},
	}, nil)

	_, _, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--cpu", "1000m",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "multiple packages match")
}

func TestCreateCluster_NoPackageOrResources(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--package")
}

func TestCreateCluster_AdditionalDisk(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BookingServer.GetPackageCalls.Returns(&bookingv1.GetPackageResponse{
		Package: &bookingv1.Package{
			Id:   "pkg-100gib",
			Name: "starter",
			ResourceConfiguration: &bookingv1.ResourceConfiguration{
				Cpu:  "1000m",
				Ram:  "1GiB",
				Disk: "100GiB",
			},
		},
	}, nil)
	env.Server.CreateClusterCalls.Returns(&clusterv1.CreateClusterResponse{
		Cluster: &clusterv1.Cluster{Id: "cluster-extra-disk"},
	}, nil)

	_, _, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--package", "550e8400-e29b-41d4-a716-446655440000",
		"--disk", "200GiB",
	)
	require.NoError(t, err)

	req, ok := env.Server.CreateClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, uint32(100), req.GetCluster().GetConfiguration().GetAdditionalResources().GetDisk())
}

func TestCreateCluster_DiskEqualToPackage(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BookingServer.GetPackageCalls.Returns(&bookingv1.GetPackageResponse{
		Package: &bookingv1.Package{
			Id:   "pkg-100gib",
			Name: "starter",
			ResourceConfiguration: &bookingv1.ResourceConfiguration{
				Cpu:  "1000m",
				Ram:  "1GiB",
				Disk: "100GiB",
			},
		},
	}, nil)
	env.Server.CreateClusterCalls.Returns(&clusterv1.CreateClusterResponse{
		Cluster: &clusterv1.Cluster{Id: "cluster-same-disk"},
	}, nil)

	_, _, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--package", "550e8400-e29b-41d4-a716-446655440000",
		"--disk", "100GiB",
	)
	require.NoError(t, err)

	req, ok := env.Server.CreateClusterCalls.Last()
	require.True(t, ok)
	assert.Nil(t, req.GetCluster().GetConfiguration().GetAdditionalResources(), "no additional disk when requested == package disk")
}

func TestCreateCluster_PackageByMultiAZ(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{
		Items: []*bookingv1.Package{
			{
				Id:   "pkg-multiaz",
				Name: "multiaz-starter",
				ResourceConfiguration: &bookingv1.ResourceConfiguration{
					Cpu: "1000m",
				},
				MultiAz: true,
			},
		},
	}, nil)
	env.Server.CreateClusterCalls.Returns(&clusterv1.CreateClusterResponse{
		Cluster: &clusterv1.Cluster{Id: "cluster-multiaz"},
	}, nil)

	_, _, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--cpu", "1000m",
		"--multi-az",
	)
	require.NoError(t, err)

	req, ok := env.Server.CreateClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "pkg-multiaz", req.GetCluster().GetConfiguration().GetPackageId())
}

func TestCreateCluster_MultiAZAloneRequiresResourceFlag(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--multi-az",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--package")
}

func TestCreateCluster_ExplicitNameSkipsSuggest(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.SuggestClusterNameCalls.Returns(&clusterv1.SuggestClusterNameResponse{Name: "should-not-use"}, nil)
	env.Server.CreateClusterCalls.Always(func(_ context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error) {
		return &clusterv1.CreateClusterResponse{
			Cluster: &clusterv1.Cluster{Id: "cluster-named", Name: req.GetCluster().GetName()},
		}, nil
	})

	_, _, err := testutil.Exec(t, env,
		"cluster", "create",
		"--name", "my-cluster",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
		"--package", "00000000-0000-0000-0000-000000000001",
	)
	require.NoError(t, err)
	assert.Equal(t, 0, env.Server.SuggestClusterNameCalls.Count(), "SuggestClusterName should not be called when --name is provided")
}
