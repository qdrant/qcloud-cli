package cluster_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"
	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

const (
	pkgID1 = "00000000-0000-0000-0000-000000000001"
	pkgID2 = "00000000-0000-0000-0000-000000000002"
	pkgID3 = "00000000-0000-0000-0000-000000000003"
	pkgID4 = "00000000-0000-0000-0000-000000000004"
	pkgIDA = "00000000-0000-0000-0000-0000000000AA"
	pkgIDB = "00000000-0000-0000-0000-0000000000BB"
)

func baseCluster() *clusterv1.Cluster {
	return &clusterv1.Cluster{
		Id:                    "cluster-123",
		Name:                  "test-cluster",
		CloudProviderId:       "aws",
		CloudProviderRegionId: "us-east-1",
		Configuration: &clusterv1.ClusterConfiguration{
			PackageId:           pkgID1,
			NumberOfNodes:       1,
			AdditionalResources: &clusterv1.AdditionalResources{Disk: 0},
		},
		State: &clusterv1.ClusterState{
			Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY,
		},
	}
}

func newPkg(id, cpu, ram, disk string) *bookingv1.Package {
	return &bookingv1.Package{
		Id:   id,
		Name: id,
		ResourceConfiguration: &bookingv1.ResourceConfiguration{
			Cpu:  cpu,
			Ram:  ram,
			Disk: disk,
		},
		MultiAz: false,
	}
}

type scaleEnv struct {
	cluster    *clusterv1.Cluster
	currentPkg *bookingv1.Package
	newPkg     *bookingv1.Package
}

func setupScale(env *testutil.TestEnv, s scaleEnv) {
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{Cluster: s.cluster}, nil)
	env.BookingServer.GetPackageCalls.Returns(&bookingv1.GetPackageResponse{Package: s.currentPkg}, nil)
	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{Items: []*bookingv1.Package{s.newPkg}}, nil)
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		return &clusterv1.UpdateClusterResponse{Cluster: req.GetCluster()}, nil
	})
}

func TestScale_NoChanges(t *testing.T) {
	env := testutil.NewTestEnv(t)
	pkg := newPkg(pkgID1, "1000m", "4GiB", "50GiB")
	setupScale(env, scaleEnv{
		cluster:    baseCluster(),
		currentPkg: pkg,
		newPkg:     pkg,
	})

	_, _, err := testutil.Exec(t, env, "cluster", "scale", "cluster-123", "--force")
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, pkgID1, req.GetCluster().GetConfiguration().GetPackageId())
	assert.Equal(t, uint32(0), req.GetCluster().GetConfiguration().GetAdditionalResources().GetDisk())
	assert.Equal(t, 0, env.BookingServer.ListPackagesCalls.Count())
}

func TestScale_Nodes(t *testing.T) {
	env := testutil.NewTestEnv(t)
	pkg := newPkg(pkgID1, "1000m", "4GiB", "50GiB")
	setupScale(env, scaleEnv{
		cluster:    baseCluster(),
		currentPkg: pkg,
		newPkg:     pkg,
	})

	_, _, err := testutil.Exec(t, env, "cluster", "scale", "cluster-123", "--nodes", "3", "--force")
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, uint32(3), req.GetCluster().GetConfiguration().GetNumberOfNodes())
	assert.Equal(t, 0, env.BookingServer.ListPackagesCalls.Count())
}

func TestScale_CPU(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupScale(env, scaleEnv{
		cluster:    baseCluster(),
		currentPkg: newPkg(pkgID1, "1000m", "4GiB", "50GiB"),
		newPkg:     newPkg(pkgID2, "2000m", "4GiB", "50GiB"),
	})

	_, _, err := testutil.Exec(t, env, "cluster", "scale", "cluster-123", "--cpu", "2", "--force")
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, pkgID2, req.GetCluster().GetConfiguration().GetPackageId())
	assert.Equal(t, 1, env.BookingServer.ListPackagesCalls.Count())
}

func TestScale_RAM(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupScale(env, scaleEnv{
		cluster:    baseCluster(),
		currentPkg: newPkg(pkgID1, "1000m", "4GiB", "50GiB"),
		newPkg:     newPkg(pkgID3, "1000m", "8GiB", "50GiB"),
	})

	_, _, err := testutil.Exec(t, env, "cluster", "scale", "cluster-123", "--ram", "8GiB", "--force")
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, pkgID3, req.GetCluster().GetConfiguration().GetPackageId())
}

func TestScale_DiskAbovePackageMinimum(t *testing.T) {
	env := testutil.NewTestEnv(t)
	pkg := newPkg(pkgID1, "1000m", "4GiB", "50GiB")
	setupScale(env, scaleEnv{
		cluster:    baseCluster(),
		currentPkg: pkg,
		newPkg:     pkg,
	})

	_, _, err := testutil.Exec(t, env, "cluster", "scale", "cluster-123", "--disk", "150GiB", "--force")
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	// 150GiB total - 50GiB pkg = 100GiB additional
	assert.Equal(t, uint32(100), req.GetCluster().GetConfiguration().GetAdditionalResources().GetDisk())
}

func TestScale_DiskBelowPackageMinimumIsOverridden(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupScale(env, scaleEnv{
		cluster:    baseCluster(),
		currentPkg: newPkg(pkgID1, "1000m", "4GiB", "50GiB"),
		newPkg:     newPkg(pkgID4, "2000m", "4GiB", "100GiB"),
	})

	_, _, err := testutil.Exec(t, env, "cluster", "scale", "cluster-123", "--cpu", "2", "--disk", "80GiB", "--force")
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	// effective = max(80GiB requested, 100GiB pkg) = 100GiB; additional = 100 - 100 = 0
	assert.Equal(t, uint32(0), req.GetCluster().GetConfiguration().GetAdditionalResources().GetDisk())
}

func TestScale_NewPackageWithLargerDiskUpgradesDiskSilently(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupScale(env, scaleEnv{
		cluster:    baseCluster(),
		currentPkg: newPkg(pkgID1, "1000m", "4GiB", "50GiB"),
		newPkg:     newPkg(pkgID4, "2000m", "4GiB", "100GiB"),
	})

	_, _, err := testutil.Exec(t, env, "cluster", "scale", "cluster-123", "--cpu", "2", "--force")
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	// effective = max(50GiB current, 100GiB new pkg) = 100GiB; additional = 100 - 100 = 0
	assert.Equal(t, uint32(0), req.GetCluster().GetConfiguration().GetAdditionalResources().GetDisk())
}

func TestScale_DiskDownscaleIsRejected(t *testing.T) {
	env := testutil.NewTestEnv(t)
	pkg := newPkg(pkgID1, "1000m", "4GiB", "50GiB")
	cluster := baseCluster()
	cluster.Configuration.AdditionalResources = &clusterv1.AdditionalResources{Disk: 50} // 50GiB pkg + 50GiB extra = 100GiB total
	setupScale(env, scaleEnv{
		cluster:    cluster,
		currentPkg: pkg,
		newPkg:     pkg,
	})

	_, _, err := testutil.Exec(t, env, "cluster", "scale", "cluster-123", "--disk", "80GiB", "--force")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be downscaled")
	assert.Equal(t, 0, env.Server.UpdateClusterCalls.Count())
}

func TestScale_AbortWithoutForce(t *testing.T) {
	env := testutil.NewTestEnv(t)
	pkg := newPkg(pkgID1, "1000m", "4GiB", "50GiB")
	setupScale(env, scaleEnv{
		cluster:    baseCluster(),
		currentPkg: pkg,
		newPkg:     pkg,
	})

	stdout, _, err := testutil.Exec(t, env, "cluster", "scale", "cluster-123")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Aborted.")
	assert.Equal(t, 0, env.Server.UpdateClusterCalls.Count())
}

func TestScale_ConfirmPromptShowsDiskCorrectly(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupScale(env, scaleEnv{
		cluster:    baseCluster(),
		currentPkg: newPkg(pkgID1, "1000m", "4GiB", "50GiB"),
		newPkg:     newPkg(pkgID2, "2000m", "4GiB", "50GiB"),
	})

	_, stderr, err := testutil.Exec(t, env, "cluster", "scale", "cluster-123", "--cpu", "2")
	require.NoError(t, err)

	assert.Contains(t, stderr, "Disk:    50GiB")
}

func TestScale_MissingClusterID(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "cluster", "scale")
	require.Error(t, err)
}

func TestScale_GetClusterError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.Server.GetClusterCalls.Returns(nil, status.Error(codes.NotFound, "not found"))

	_, _, err := testutil.Exec(t, env, "cluster", "scale", "cluster-123", "--force")
	require.Error(t, err)
	assert.Equal(t, 0, env.BookingServer.GetPackageCalls.Count())
}

func TestScale_GetPackageError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{Cluster: baseCluster()}, nil)
	env.BookingServer.GetPackageCalls.Returns(nil, status.Error(codes.Internal, "package error"))

	_, _, err := testutil.Exec(t, env, "cluster", "scale", "cluster-123", "--force")
	require.Error(t, err)
	assert.Equal(t, 0, env.BookingServer.ListPackagesCalls.Count())
}

func TestScale_PackageNotFound(t *testing.T) {
	env := testutil.NewTestEnv(t)
	currentPkg := newPkg(pkgID1, "1000m", "4GiB", "50GiB")
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{Cluster: baseCluster()}, nil)
	env.BookingServer.GetPackageCalls.Returns(&bookingv1.GetPackageResponse{Package: currentPkg}, nil)
	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{Items: nil}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "scale", "cluster-123", "--cpu", "1", "--force")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no package found")
}

func TestScale_AmbiguousPackageMatch(t *testing.T) {
	env := testutil.NewTestEnv(t)
	currentPkg := newPkg(pkgID1, "1000m", "4GiB", "50GiB")
	pkgA := newPkg(pkgIDA, "1000m", "4GiB", "50GiB")
	pkgB := newPkg(pkgIDB, "1000m", "4GiB", "80GiB")
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{Cluster: baseCluster()}, nil)
	env.BookingServer.GetPackageCalls.Returns(&bookingv1.GetPackageResponse{Package: currentPkg}, nil)
	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{Items: []*bookingv1.Package{pkgA, pkgB}}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "scale", "cluster-123", "--cpu", "1", "--force")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "multiple packages match")
}

func TestScale_UpdateClusterError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	pkg := newPkg(pkgID1, "1000m", "4GiB", "50GiB")
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{Cluster: baseCluster()}, nil)
	env.BookingServer.GetPackageCalls.Returns(&bookingv1.GetPackageResponse{Package: pkg}, nil)
	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{Items: []*bookingv1.Package{pkg}}, nil)
	env.Server.UpdateClusterCalls.Returns(nil, status.Error(codes.Internal, "update failed"))

	_, _, err := testutil.Exec(t, env, "cluster", "scale", "cluster-123", "--force")
	require.Error(t, err)
}

func TestScale_WaitSuccess(t *testing.T) {
	env := testutil.NewTestEnv(t)
	pkg := newPkg(pkgID1, "1000m", "4GiB", "50GiB")
	cluster := baseCluster()

	env.Server.GetClusterCalls.
		OnCall(0, func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
			return &clusterv1.GetClusterResponse{Cluster: cluster}, nil
		}).
		OnCall(1, func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
			return &clusterv1.GetClusterResponse{Cluster: &clusterv1.Cluster{
				Id:   "cluster-123",
				Name: "test-cluster",
				State: &clusterv1.ClusterState{
					Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_SCALING,
				},
			}}, nil
		}).
		Always(func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
			return &clusterv1.GetClusterResponse{Cluster: &clusterv1.Cluster{
				Id:   "cluster-123",
				Name: "test-cluster",
				State: &clusterv1.ClusterState{
					Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY,
				},
			}}, nil
		})
	env.BookingServer.GetPackageCalls.Returns(&bookingv1.GetPackageResponse{Package: pkg}, nil)
	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{Items: []*bookingv1.Package{pkg}}, nil)
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		return &clusterv1.UpdateClusterResponse{Cluster: req.GetCluster()}, nil
	})

	stdout, stderr, err := testutil.Exec(t, env,
		"cluster", "scale", "cluster-123", "--force",
		"--wait",
		"--wait-timeout", "30s",
		"--wait-poll-interval", "10ms",
	)
	require.NoError(t, err)
	assert.Contains(t, stderr, "Scaling Cluster")
	assert.Contains(t, stdout, "scaled successfully")
}

func TestScale_WaitTimeout(t *testing.T) {
	env := testutil.NewTestEnv(t)
	pkg := newPkg(pkgID1, "1000m", "4GiB", "50GiB")
	cluster := baseCluster()

	env.Server.GetClusterCalls.
		OnCall(0, func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
			return &clusterv1.GetClusterResponse{Cluster: cluster}, nil
		}).
		Always(func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
			return &clusterv1.GetClusterResponse{Cluster: &clusterv1.Cluster{
				Id: "cluster-123",
				State: &clusterv1.ClusterState{
					Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_SCALING,
				},
			}}, nil
		})
	env.BookingServer.GetPackageCalls.Returns(&bookingv1.GetPackageResponse{Package: pkg}, nil)
	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{Items: []*bookingv1.Package{pkg}}, nil)
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		return &clusterv1.UpdateClusterResponse{Cluster: req.GetCluster()}, nil
	})

	_, _, err := testutil.Exec(t, env,
		"cluster", "scale", "cluster-123", "--force",
		"--wait",
		"--wait-timeout", "200ms",
		"--wait-poll-interval", "10ms",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
}

func TestScale_NoWait(t *testing.T) {
	env := testutil.NewTestEnv(t)
	pkg := newPkg(pkgID1, "1000m", "4GiB", "50GiB")
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{Cluster: baseCluster()}, nil)
	env.BookingServer.GetPackageCalls.Returns(&bookingv1.GetPackageResponse{Package: pkg}, nil)
	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{Items: []*bookingv1.Package{pkg}}, nil)
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		c := req.GetCluster()
		c.State = &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_SCALING}
		return &clusterv1.UpdateClusterResponse{Cluster: c}, nil
	})

	stdout, _, err := testutil.Exec(t, env, "cluster", "scale", "cluster-123", "--force")
	require.NoError(t, err)
	assert.Contains(t, stdout, "it will take some time")
	assert.Equal(t, 1, env.Server.GetClusterCalls.Count(), "GetCluster should be called only once without --wait")
}

func TestScale_PrintsSuccessWhenClusterIsHealthy(t *testing.T) {
	env := testutil.NewTestEnv(t)
	pkg := newPkg(pkgID1, "1000m", "4GiB", "50GiB")
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{Cluster: baseCluster()}, nil)
	env.BookingServer.GetPackageCalls.Returns(&bookingv1.GetPackageResponse{Package: pkg}, nil)
	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{Items: []*bookingv1.Package{pkg}}, nil)
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		c := req.GetCluster()
		c.State = &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY}
		return &clusterv1.UpdateClusterResponse{Cluster: c}, nil
	})

	stdout, _, err := testutil.Exec(t, env, "cluster", "scale", "cluster-123", "--force")
	require.NoError(t, err)
	assert.Contains(t, stdout, "scaled successfully")
}

func TestScale_PrintsScalingWhenClusterIsNotYetHealthy(t *testing.T) {
	env := testutil.NewTestEnv(t)
	pkg := newPkg(pkgID1, "1000m", "4GiB", "50GiB")
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{Cluster: baseCluster()}, nil)
	env.BookingServer.GetPackageCalls.Returns(&bookingv1.GetPackageResponse{Package: pkg}, nil)
	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{Items: []*bookingv1.Package{pkg}}, nil)
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		c := req.GetCluster()
		c.State = &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_SCALING}
		return &clusterv1.UpdateClusterResponse{Cluster: c}, nil
	})

	stdout, _, err := testutil.Exec(t, env, "cluster", "scale", "cluster-123", "--force")
	require.NoError(t, err)
	assert.Contains(t, stdout, "is scaling")
	assert.Contains(t, stdout, "cluster-123")
}

func TestScale_NoResourceFlagsSkipsPackageResolution(t *testing.T) {
	env := testutil.NewTestEnv(t)
	pkg := newPkg(pkgID1, "1000m", "4GiB", "50GiB")
	setupScale(env, scaleEnv{
		cluster:    baseCluster(),
		currentPkg: pkg,
		newPkg:     pkg,
	})

	_, _, err := testutil.Exec(t, env, "cluster", "scale", "cluster-123", "--force")
	require.NoError(t, err)

	assert.Equal(t, 0, env.BookingServer.ListPackagesCalls.Count())
	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, pkgID1, req.GetCluster().GetConfiguration().GetPackageId())
}

func TestScale_DiskPerformance(t *testing.T) {
	env := testutil.NewTestEnv(t)
	pkg := newPkg(pkgID1, "1000m", "4GiB", "50GiB")
	setupScale(env, scaleEnv{
		cluster:    baseCluster(),
		currentPkg: pkg,
		newPkg:     pkg,
	})

	_, _, err := testutil.Exec(t, env, "cluster", "scale", "cluster-123", "--disk-performance", "performance", "--force")
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, commonv1.StorageTierType_STORAGE_TIER_TYPE_PERFORMANCE, req.GetCluster().GetConfiguration().GetClusterStorageConfiguration().GetStorageTierType())
}

func TestScale_InvalidDiskPerformance(t *testing.T) {
	env := testutil.NewTestEnv(t)
	pkg := newPkg(pkgID1, "1000m", "4GiB", "50GiB")
	setupScale(env, scaleEnv{
		cluster:    baseCluster(),
		currentPkg: pkg,
		newPkg:     pkg,
	})

	_, _, err := testutil.Exec(t, env, "cluster", "scale", "cluster-123", "--disk-performance", "ultra", "--force")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ultra")
}

func TestScale_DeprecatedCurrentPackageSucceedsWithNoResourceFlags(t *testing.T) {
	env := testutil.NewTestEnv(t)
	currentPkg := newPkg(pkgID1, "1000m", "4GiB", "50GiB")
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{Cluster: baseCluster()}, nil)
	env.BookingServer.GetPackageCalls.Returns(&bookingv1.GetPackageResponse{Package: currentPkg}, nil)
	// ListPackages returns empty — simulates a deprecated package absent from the active list.
	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{Items: nil}, nil)
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		return &clusterv1.UpdateClusterResponse{Cluster: req.GetCluster()}, nil
	})

	_, _, err := testutil.Exec(t, env, "cluster", "scale", "cluster-123", "--force")
	require.NoError(t, err)

	assert.Equal(t, 0, env.BookingServer.ListPackagesCalls.Count())
	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, pkgID1, req.GetCluster().GetConfiguration().GetPackageId())
}
