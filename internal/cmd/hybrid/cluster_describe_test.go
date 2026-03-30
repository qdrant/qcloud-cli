package hybrid_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestHybridClusterDescribe_FullOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	version := "1.8.0"
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{
			Id:                    "cluster-abc",
			Name:                  "my-cluster",
			CloudProviderId:       "hybrid",
			CloudProviderRegionId: "env-123",
			State:                 &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY},
			Configuration: &clusterv1.ClusterConfiguration{
				Version:       &version,
				NumberOfNodes: 3,
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "cluster", "describe", "cluster-abc")
	require.NoError(t, err)

	assert.Contains(t, stdout, "cluster-abc")
	assert.Contains(t, stdout, "my-cluster")
	assert.Contains(t, stdout, "HEALTHY")
	assert.Contains(t, stdout, "env-123")
	assert.Contains(t, stdout, "1.8.0")
	assert.Contains(t, stdout, "3")
}

func TestHybridClusterDescribe_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "hybrid", "cluster", "describe")
	require.Error(t, err)
}

func TestHybridClusterDescribe_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Returns(nil, fmt.Errorf("not found"))

	_, _, err := testutil.Exec(t, env, "hybrid", "cluster", "describe", "cluster-abc")
	require.Error(t, err)
}

func TestHybridClusterDescribe_NewFields(t *testing.T) {
	env := testutil.NewTestEnv(t)

	version := "1.13.0"
	costLabel := "team-platform"
	dbStorageClass := "fast-ssd"
	snapStorageClass := "cold-hdd"
	volSnapClass := "snap-class"
	volAttrClass := "attr-class"
	enableTLS := true
	vectorsOnDisk := true
	auditEnabled := true
	trustFwd := true
	maxFiles := uint32(14)

	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{
			Id:                    "cluster-full",
			Name:                  "full-cluster",
			CloudProviderId:       "hybrid",
			CloudProviderRegionId: "env-456",
			CostAllocationLabel:   &costLabel,
			State:                 &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY},
			Configuration: &clusterv1.ClusterConfiguration{
				Version:             &version,
				NumberOfNodes:       3,
				RestartPolicy:       clusterv1.ClusterConfigurationRestartPolicy_CLUSTER_CONFIGURATION_RESTART_POLICY_ROLLING.Enum(),
				RebalanceStrategy:   clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_BY_COUNT.Enum(),
				AdditionalResources: &clusterv1.AdditionalResources{Disk: 50},
				TopologySpreadConstraints: []*commonv1.TopologySpreadConstraint{
					{TopologyKey: "topology.kubernetes.io/zone"},
				},
				ClusterStorageConfiguration: &clusterv1.ClusterStorageConfiguration{
					DatabaseStorageClass:  &dbStorageClass,
					SnapshotStorageClass:  &snapStorageClass,
					VolumeSnapshotClass:   &volSnapClass,
					VolumeAttributesClass: &volAttrClass,
				},
				DatabaseConfiguration: &clusterv1.DatabaseConfiguration{
					LogLevel: clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_DEBUG.Enum(),
					Collection: &clusterv1.DatabaseConfigurationCollection{
						Vectors: &clusterv1.DatabaseConfigurationCollectionVectors{
							OnDisk: &vectorsOnDisk,
						},
					},
					Service: &clusterv1.DatabaseConfigurationService{
						EnableTls: &enableTLS,
						ApiKey:    &commonv1.SecretKeyRef{Name: "my-secret", Key: "api-key"},
					},
					Tls: &clusterv1.DatabaseConfigurationTls{
						Cert: &commonv1.SecretKeyRef{Name: "tls-sec", Key: "cert"},
						Key:  &commonv1.SecretKeyRef{Name: "tls-sec", Key: "key"},
					},
					AuditLogging: &clusterv1.DatabaseConfigurationAuditLogging{
						Enabled:               auditEnabled,
						Rotation:              clusterv1.AuditLogRotation_AUDIT_LOG_ROTATION_DAILY.Enum(),
						MaxLogFiles:           &maxFiles,
						TrustForwardedHeaders: &trustFwd,
					},
				},
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "cluster", "describe", "cluster-full")
	require.NoError(t, err)

	assert.Contains(t, stdout, "team-platform")
	assert.Contains(t, stdout, "rolling")
	assert.Contains(t, stdout, "by-count")
	assert.Contains(t, stdout, "50")
	assert.Contains(t, stdout, "topology.kubernetes.io/zone")
	assert.Contains(t, stdout, "fast-ssd")
	assert.Contains(t, stdout, "cold-hdd")
	assert.Contains(t, stdout, "snap-class")
	assert.Contains(t, stdout, "attr-class")
	assert.Contains(t, stdout, "debug")
	assert.Contains(t, stdout, "yes")
	assert.Contains(t, stdout, "my-secret:api-key")
	assert.Contains(t, stdout, "tls-sec:cert")
	assert.Contains(t, stdout, "tls-sec:key")
	assert.Contains(t, stdout, "daily")
	assert.Contains(t, stdout, "14")
}
