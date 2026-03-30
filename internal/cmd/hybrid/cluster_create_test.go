package hybrid_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"
	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestHybridClusterCreate_Basic(t *testing.T) {
	env := testutil.NewTestEnv(t)

	setupDefaultBooking(env)
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

	setupDefaultBooking(env)
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

	setupDefaultBooking(env)
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

	setupDefaultBooking(env)
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

	setupDefaultBooking(env)
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

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func defaultListPackagesResponse() *bookingv1.ListPackagesResponse {
	return &bookingv1.ListPackagesResponse{
		Items: []*bookingv1.Package{
			{
				Id:     "pkg-default",
				Name:   "default-pkg",
				Status: bookingv1.PackageStatus_PACKAGE_STATUS_ACTIVE,
				ResourceConfiguration: &bookingv1.ResourceConfiguration{
					Cpu:  "1000m",
					Ram:  "4GiB",
					Disk: "100GiB",
				},
			},
		},
	}
}

func setupDefaultBooking(env *testutil.TestEnv) {
	env.BookingServer.ListPackagesCalls.Returns(defaultListPackagesResponse(), nil)
}

// ---------------------------------------------------------------------------
// Package resolution
// ---------------------------------------------------------------------------

func TestHybridClusterCreate_PackageByName(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{
		Items: []*bookingv1.Package{
			{
				Id:     "pkg-alpha",
				Name:   "alpha",
				Status: bookingv1.PackageStatus_PACKAGE_STATUS_ACTIVE,
				ResourceConfiguration: &bookingv1.ResourceConfiguration{
					Cpu:  "2000m",
					Ram:  "8GiB",
					Disk: "200GiB",
				},
			},
			{
				Id:     "pkg-beta",
				Name:   "beta",
				Status: bookingv1.PackageStatus_PACKAGE_STATUS_ACTIVE,
				ResourceConfiguration: &bookingv1.ResourceConfiguration{
					Cpu:  "4000m",
					Ram:  "16GiB",
					Disk: "400GiB",
				},
			},
		},
	}, nil)
	env.Server.CreateClusterCalls.Always(func(_ context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error) {
		return &clusterv1.CreateClusterResponse{
			Cluster: &clusterv1.Cluster{Id: "cluster-pkg-name", Name: req.GetCluster().GetName()},
		}, nil
	})

	_, _, err := testutil.Exec(t, env, "hybrid", "cluster", "create", "env-123",
		"--name", "my-cluster",
		"--package", "beta",
	)
	require.NoError(t, err)

	req, ok := env.Server.CreateClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "pkg-beta", req.GetCluster().GetConfiguration().GetPackageId())
}

func TestHybridClusterCreate_PackageByCPUAndRAM(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{
		Items: []*bookingv1.Package{
			{
				Id:     "pkg-small",
				Name:   "small",
				Status: bookingv1.PackageStatus_PACKAGE_STATUS_ACTIVE,
				ResourceConfiguration: &bookingv1.ResourceConfiguration{
					Cpu:  "1000m",
					Ram:  "4GiB",
					Disk: "100GiB",
				},
			},
			{
				Id:     "pkg-medium",
				Name:   "medium",
				Status: bookingv1.PackageStatus_PACKAGE_STATUS_ACTIVE,
				ResourceConfiguration: &bookingv1.ResourceConfiguration{
					Cpu:  "2000m",
					Ram:  "8GiB",
					Disk: "200GiB",
				},
			},
		},
	}, nil)
	env.Server.CreateClusterCalls.Always(func(_ context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error) {
		return &clusterv1.CreateClusterResponse{
			Cluster: &clusterv1.Cluster{Id: "cluster-cpu-ram", Name: req.GetCluster().GetName()},
		}, nil
	})

	_, _, err := testutil.Exec(t, env, "hybrid", "cluster", "create", "env-123",
		"--name", "resource-cluster",
		"--cpu", "2",
		"--ram", "8Gi",
	)
	require.NoError(t, err)

	req, ok := env.Server.CreateClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "pkg-medium", req.GetCluster().GetConfiguration().GetPackageId())
}

func TestHybridClusterCreate_PackageMutualExclusion(t *testing.T) {
	env := testutil.NewTestEnv(t)

	setupDefaultBooking(env)
	env.Server.CreateClusterCalls.Returns(&clusterv1.CreateClusterResponse{
		Cluster: &clusterv1.Cluster{Id: "should-not-reach"},
	}, nil)

	_, _, err := testutil.Exec(t, env, "hybrid", "cluster", "create", "env-123",
		"--name", "bad-combo",
		"--package", "default-pkg",
		"--cpu", "2",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
	assert.Equal(t, 0, env.Server.CreateClusterCalls.Count(), "CreateCluster should not be called")
}

func TestHybridClusterCreate_PackageWithDisk(t *testing.T) {
	env := testutil.NewTestEnv(t)

	setupDefaultBooking(env)
	env.Server.CreateClusterCalls.Always(func(_ context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error) {
		return &clusterv1.CreateClusterResponse{
			Cluster: &clusterv1.Cluster{Id: "cluster-disk", Name: req.GetCluster().GetName()},
		}, nil
	})

	_, _, err := testutil.Exec(t, env, "hybrid", "cluster", "create", "env-123",
		"--name", "disk-cluster",
		"--disk", "150GiB",
	)
	require.NoError(t, err)

	req, ok := env.Server.CreateClusterCalls.Last()
	require.True(t, ok)
	cfg := req.GetCluster().GetConfiguration()
	assert.Equal(t, "pkg-default", cfg.GetPackageId())
	assert.Equal(t, uint32(50), cfg.GetAdditionalResources().GetDisk(), "150GiB requested - 100GiB included = 50GiB additional")
}

// ---------------------------------------------------------------------------
// Cluster configuration flags
// ---------------------------------------------------------------------------

func TestHybridClusterCreate_ClusterConfigFlags(t *testing.T) {
	env := testutil.NewTestEnv(t)

	setupDefaultBooking(env)
	env.Server.CreateClusterCalls.Always(func(_ context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error) {
		return &clusterv1.CreateClusterResponse{
			Cluster: &clusterv1.Cluster{Id: "cluster-cfgflags", Name: req.GetCluster().GetName()},
		}, nil
	})

	_, _, err := testutil.Exec(t, env, "hybrid", "cluster", "create", "env-123",
		"--name", "cfg-cluster",
		"--restart-policy", "rolling",
		"--rebalance-strategy", "by-count-and-size",
		"--topology-spread-constraint", "topology.kubernetes.io/zone:1:do-not-schedule",
		"--cost-allocation-label", "team-backend",
	)
	require.NoError(t, err)

	req, ok := env.Server.CreateClusterCalls.Last()
	require.True(t, ok)
	cluster := req.GetCluster()
	cfg := cluster.GetConfiguration()

	assert.Equal(t,
		clusterv1.ClusterConfigurationRestartPolicy_CLUSTER_CONFIGURATION_RESTART_POLICY_ROLLING,
		cfg.GetRestartPolicy())
	assert.Equal(t,
		clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_BY_COUNT_AND_SIZE,
		cfg.GetRebalanceStrategy())

	require.Len(t, cfg.GetTopologySpreadConstraints(), 1)
	tsc := cfg.GetTopologySpreadConstraints()[0]
	assert.Equal(t, "topology.kubernetes.io/zone", tsc.GetTopologyKey())
	assert.Equal(t, int32(1), tsc.GetMaxSkew())
	assert.Equal(t,
		commonv1.TopologySpreadConstraintWhenUnsatisfiable_TOPOLOGY_SPREAD_CONSTRAINT_WHEN_UNSATISFIABLE_DO_NOT_SCHEDULE,
		tsc.GetWhenUnsatisfiable())

	assert.Equal(t, "team-backend", cluster.GetCostAllocationLabel())
}

// ---------------------------------------------------------------------------
// Storage configuration flags
// ---------------------------------------------------------------------------

func TestHybridClusterCreate_StorageConfigFlags(t *testing.T) {
	env := testutil.NewTestEnv(t)

	setupDefaultBooking(env)
	env.Server.CreateClusterCalls.Always(func(_ context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error) {
		return &clusterv1.CreateClusterResponse{
			Cluster: &clusterv1.Cluster{Id: "cluster-storage", Name: req.GetCluster().GetName()},
		}, nil
	})

	_, _, err := testutil.Exec(t, env, "hybrid", "cluster", "create", "env-123",
		"--name", "storage-cluster",
		"--database-storage-class", "fast-ssd",
		"--snapshot-storage-class", "cold-hdd",
		"--volume-snapshot-class", "snap-class",
		"--volume-attributes-class", "attr-class",
	)
	require.NoError(t, err)

	req, ok := env.Server.CreateClusterCalls.Last()
	require.True(t, ok)
	sc := req.GetCluster().GetConfiguration().GetClusterStorageConfiguration()
	require.NotNil(t, sc)
	assert.Equal(t, "fast-ssd", sc.GetDatabaseStorageClass())
	assert.Equal(t, "cold-hdd", sc.GetSnapshotStorageClass())
	assert.Equal(t, "snap-class", sc.GetVolumeSnapshotClass())
	assert.Equal(t, "attr-class", sc.GetVolumeAttributesClass())
}

// ---------------------------------------------------------------------------
// Database configuration flags
// ---------------------------------------------------------------------------

func TestHybridClusterCreate_DatabaseConfigFlags(t *testing.T) {
	env := testutil.NewTestEnv(t)

	setupDefaultBooking(env)
	env.Server.CreateClusterCalls.Always(func(_ context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error) {
		return &clusterv1.CreateClusterResponse{
			Cluster: &clusterv1.Cluster{Id: "cluster-dbcfg", Name: req.GetCluster().GetName()},
		}, nil
	})

	_, _, err := testutil.Exec(t, env, "hybrid", "cluster", "create", "env-123",
		"--name", "db-cluster",
		"--db-log-level", "debug",
		"--vectors-on-disk",
		"--enable-tls",
		"--api-key-secret", "my-secret:api-key",
		"--read-only-api-key-secret", "my-secret:ro-key",
		"--tls-cert-secret", "tls-secrets:cert.pem",
		"--tls-key-secret", "tls-secrets:key.pem",
		"--audit-logging",
		"--audit-log-rotation", "daily",
		"--audit-log-max-files", "14",
		"--audit-log-trust-forwarded-headers",
	)
	require.NoError(t, err)

	req, ok := env.Server.CreateClusterCalls.Last()
	require.True(t, ok)
	dbCfg := req.GetCluster().GetConfiguration().GetDatabaseConfiguration()
	require.NotNil(t, dbCfg)

	// Log level
	assert.Equal(t,
		clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_DEBUG,
		dbCfg.GetLogLevel())

	// Vectors on disk
	require.NotNil(t, dbCfg.GetCollection())
	require.NotNil(t, dbCfg.GetCollection().GetVectors())
	assert.True(t, dbCfg.GetCollection().GetVectors().GetOnDisk())

	// Service / TLS flags
	svc := dbCfg.GetService()
	require.NotNil(t, svc)
	assert.True(t, svc.GetEnableTls())
	assert.Equal(t, &commonv1.SecretKeyRef{Name: "my-secret", Key: "api-key"}, svc.GetApiKey())
	assert.Equal(t, &commonv1.SecretKeyRef{Name: "my-secret", Key: "ro-key"}, svc.GetReadOnlyApiKey())

	// TLS cert/key
	tlsCfg := dbCfg.GetTls()
	require.NotNil(t, tlsCfg)
	assert.Equal(t, &commonv1.SecretKeyRef{Name: "tls-secrets", Key: "cert.pem"}, tlsCfg.GetCert())
	assert.Equal(t, &commonv1.SecretKeyRef{Name: "tls-secrets", Key: "key.pem"}, tlsCfg.GetKey())

	// Audit logging
	audit := dbCfg.GetAuditLogging()
	require.NotNil(t, audit)
	assert.True(t, audit.GetEnabled())
	assert.Equal(t, clusterv1.AuditLogRotation_AUDIT_LOG_ROTATION_DAILY, audit.GetRotation())
	assert.Equal(t, uint32(14), audit.GetMaxLogFiles())
	assert.True(t, audit.GetTrustForwardedHeaders())
}
