package hybrid_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func setupHybridClusterUpdateHandlers(env *testutil.TestEnv) {
	env.Server.GetClusterCalls.Always(func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{Id: req.GetClusterId(), Name: "my-cluster", CloudProviderId: "hybrid"},
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
		Cluster: &clusterv1.Cluster{Id: "cluster-abc", Name: "my-cluster", CloudProviderId: "hybrid"},
	}, nil)
	env.Server.UpdateClusterCalls.Returns(nil, fmt.Errorf("internal server error"))

	_, _, err := testutil.Exec(t, env, "hybrid", "cluster", "update", "cluster-abc", "--label", "env=prod")
	require.Error(t, err)
}

func TestHybridClusterUpdate_ClusterConfigFlags(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupHybridClusterUpdateHandlers(env)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "cluster", "update", "cluster-abc",
		"--restart-policy", "parallel",
		"--rebalance-strategy", "by-size",
		"--topology-spread-constraint", "topology.kubernetes.io/zone:2:schedule-anyway",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "updated")

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	cfg := req.GetCluster().GetConfiguration()
	require.NotNil(t, cfg)

	assert.Equal(t,
		clusterv1.ClusterConfigurationRestartPolicy_CLUSTER_CONFIGURATION_RESTART_POLICY_PARALLEL,
		cfg.GetRestartPolicy())
	assert.Equal(t,
		clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_BY_SIZE,
		cfg.GetRebalanceStrategy())

	require.Len(t, cfg.GetTopologySpreadConstraints(), 1)
	tsc := cfg.GetTopologySpreadConstraints()[0]
	assert.Equal(t, "topology.kubernetes.io/zone", tsc.GetTopologyKey())
	assert.Equal(t, int32(2), tsc.GetMaxSkew())
	assert.Equal(t,
		commonv1.TopologySpreadConstraintWhenUnsatisfiable_TOPOLOGY_SPREAD_CONSTRAINT_WHEN_UNSATISFIABLE_SCHEDULE_ANYWAY,
		tsc.GetWhenUnsatisfiable())
}

func TestHybridClusterUpdate_StorageAndResourceFlags(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupHybridClusterUpdateHandlers(env)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "cluster", "update", "cluster-abc",
		"--database-storage-class", "fast-ssd",
		"--snapshot-storage-class", "cold-hdd",
		"--volume-snapshot-class", "snap-class",
		"--volume-attributes-class", "attr-class",
		"--additional-disk", "50",
		"--cost-allocation-label", "billing-team",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "updated")

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	cluster := req.GetCluster()
	cfg := cluster.GetConfiguration()
	require.NotNil(t, cfg)

	sc := cfg.GetClusterStorageConfiguration()
	require.NotNil(t, sc)
	assert.Equal(t, "fast-ssd", sc.GetDatabaseStorageClass())
	assert.Equal(t, "cold-hdd", sc.GetSnapshotStorageClass())
	assert.Equal(t, "snap-class", sc.GetVolumeSnapshotClass())
	assert.Equal(t, "attr-class", sc.GetVolumeAttributesClass())

	addRes := cfg.GetAdditionalResources()
	require.NotNil(t, addRes)
	assert.Equal(t, uint32(50), addRes.GetDisk())

	assert.Equal(t, "billing-team", cluster.GetCostAllocationLabel())
}

func TestHybridClusterUpdate_ExtendedDBConfig_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupHybridClusterUpdateHandlers(env)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "cluster", "update", "cluster-abc",
		"--vectors-on-disk",
		"--db-log-level", "warn",
		"--enable-tls",
		"--api-key-secret", "my-secret:key",
		"--read-only-api-key-secret", "my-secret:ro",
		"--tls-cert-secret", "tls:cert",
		"--tls-key-secret", "tls:key",
		"--audit-logging",
		"--audit-log-rotation", "hourly",
		"--audit-log-max-files", "7",
		"--audit-log-trust-forwarded-headers",
		"--force",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "updated")

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	dbCfg := req.GetCluster().GetConfiguration().GetDatabaseConfiguration()
	require.NotNil(t, dbCfg)

	assert.Equal(t,
		clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_WARN,
		dbCfg.GetLogLevel())

	require.NotNil(t, dbCfg.GetCollection())
	require.NotNil(t, dbCfg.GetCollection().GetVectors())
	assert.True(t, dbCfg.GetCollection().GetVectors().GetOnDisk())

	svc := dbCfg.GetService()
	require.NotNil(t, svc)
	assert.True(t, svc.GetEnableTls())
	require.NotNil(t, svc.GetApiKey())
	assert.True(t, proto.Equal(&commonv1.SecretKeyRef{Name: "my-secret", Key: "key"}, svc.GetApiKey()))
	require.NotNil(t, svc.GetReadOnlyApiKey())
	assert.True(t, proto.Equal(&commonv1.SecretKeyRef{Name: "my-secret", Key: "ro"}, svc.GetReadOnlyApiKey()))

	tlsCfg := dbCfg.GetTls()
	require.NotNil(t, tlsCfg)
	require.NotNil(t, tlsCfg.GetCert())
	assert.True(t, proto.Equal(&commonv1.SecretKeyRef{Name: "tls", Key: "cert"}, tlsCfg.GetCert()))
	require.NotNil(t, tlsCfg.GetKey())
	assert.True(t, proto.Equal(&commonv1.SecretKeyRef{Name: "tls", Key: "key"}, tlsCfg.GetKey()))

	audit := dbCfg.GetAuditLogging()
	require.NotNil(t, audit)
	assert.True(t, audit.GetEnabled())
	assert.Equal(t, clusterv1.AuditLogRotation_AUDIT_LOG_ROTATION_HOURLY, audit.GetRotation())
	assert.Equal(t, uint32(7), audit.GetMaxLogFiles())
	assert.True(t, audit.GetTrustForwardedHeaders())
}
