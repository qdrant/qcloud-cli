package hybrid

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"
)

func TestParseRestartPolicy(t *testing.T) {
	tests := []struct {
		input   string
		want    clusterv1.ClusterConfigurationRestartPolicy
		wantErr bool
	}{
		{"rolling", clusterv1.ClusterConfigurationRestartPolicy_CLUSTER_CONFIGURATION_RESTART_POLICY_ROLLING, false},
		{"parallel", clusterv1.ClusterConfigurationRestartPolicy_CLUSTER_CONFIGURATION_RESTART_POLICY_PARALLEL, false},
		{"automatic", clusterv1.ClusterConfigurationRestartPolicy_CLUSTER_CONFIGURATION_RESTART_POLICY_AUTOMATIC, false},
		{"invalid", clusterv1.ClusterConfigurationRestartPolicy_CLUSTER_CONFIGURATION_RESTART_POLICY_UNSPECIFIED, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseRestartPolicy(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
				assert.Equal(t, tt.input, restartPolicyString(got))
			}
		})
	}
}

func TestParseRebalanceStrategy(t *testing.T) {
	tests := []struct {
		input   string
		want    clusterv1.ClusterConfigurationRebalanceStrategy
		wantErr bool
	}{
		{"by-count", clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_BY_COUNT, false},
		{"by-size", clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_BY_SIZE, false},
		{"by-count-and-size", clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_BY_COUNT_AND_SIZE, false},
		{"invalid", clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_UNSPECIFIED, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseRebalanceStrategy(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
				assert.Equal(t, tt.input, rebalanceStrategyString(got))
			}
		})
	}
}

func TestParseGpuType(t *testing.T) {
	tests := []struct {
		input   string
		want    clusterv1.ClusterConfigurationGpuType
		wantErr bool
	}{
		{"nvidia", clusterv1.ClusterConfigurationGpuType_CLUSTER_CONFIGURATION_GPU_TYPE_NVIDIA, false},
		{"amd", clusterv1.ClusterConfigurationGpuType_CLUSTER_CONFIGURATION_GPU_TYPE_AMD, false},
		{"intel", clusterv1.ClusterConfigurationGpuType_CLUSTER_CONFIGURATION_GPU_TYPE_UNSPECIFIED, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseGpuType(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
				assert.Equal(t, tt.input, gpuTypeString(got))
			}
		})
	}
}

func TestParseDBLogLevel(t *testing.T) {
	tests := []struct {
		input   string
		want    clusterv1.DatabaseConfigurationLogLevel
		wantErr bool
	}{
		{"trace", clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_TRACE, false},
		{"debug", clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_DEBUG, false},
		{"info", clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_INFO, false},
		{"warn", clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_WARN, false},
		{"error", clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_ERROR, false},
		{"off", clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_OFF, false},
		{"verbose", clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_UNSPECIFIED, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseDBLogLevel(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
				assert.Equal(t, tt.input, dbLogLevelString(got))
			}
		})
	}
}

func TestParseAuditLogRotation(t *testing.T) {
	tests := []struct {
		input   string
		want    clusterv1.AuditLogRotation
		wantErr bool
	}{
		{"daily", clusterv1.AuditLogRotation_AUDIT_LOG_ROTATION_DAILY, false},
		{"hourly", clusterv1.AuditLogRotation_AUDIT_LOG_ROTATION_HOURLY, false},
		{"weekly", clusterv1.AuditLogRotation_AUDIT_LOG_ROTATION_UNSPECIFIED, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseAuditLogRotation(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
				assert.Equal(t, tt.input, auditLogRotationString(got))
			}
		})
	}
}

func TestParseTopologySpreadConstraint(t *testing.T) {
	t.Run("key only", func(t *testing.T) {
		tsc, err := parseTopologySpreadConstraint("topology.kubernetes.io/zone")
		require.NoError(t, err)
		assert.Equal(t, "topology.kubernetes.io/zone", tsc.GetTopologyKey())
		assert.Nil(t, tsc.MaxSkew)
		assert.Nil(t, tsc.WhenUnsatisfiable)
	})

	t.Run("key and maxSkew", func(t *testing.T) {
		tsc, err := parseTopologySpreadConstraint("topology.kubernetes.io/zone:3")
		require.NoError(t, err)
		assert.Equal(t, "topology.kubernetes.io/zone", tsc.GetTopologyKey())
		assert.Equal(t, int32(3), tsc.GetMaxSkew())
		assert.Nil(t, tsc.WhenUnsatisfiable)
	})

	t.Run("full format", func(t *testing.T) {
		tsc, err := parseTopologySpreadConstraint("topology.kubernetes.io/zone:1:do-not-schedule")
		require.NoError(t, err)
		assert.Equal(t, "topology.kubernetes.io/zone", tsc.GetTopologyKey())
		assert.Equal(t, int32(1), tsc.GetMaxSkew())
		assert.Equal(t,
			commonv1.TopologySpreadConstraintWhenUnsatisfiable_TOPOLOGY_SPREAD_CONSTRAINT_WHEN_UNSATISFIABLE_DO_NOT_SCHEDULE,
			tsc.GetWhenUnsatisfiable())
	})

	t.Run("schedule-anyway", func(t *testing.T) {
		tsc, err := parseTopologySpreadConstraint("zone:2:schedule-anyway")
		require.NoError(t, err)
		assert.Equal(t,
			commonv1.TopologySpreadConstraintWhenUnsatisfiable_TOPOLOGY_SPREAD_CONSTRAINT_WHEN_UNSATISFIABLE_SCHEDULE_ANYWAY,
			tsc.GetWhenUnsatisfiable())
	})

	t.Run("empty key", func(t *testing.T) {
		_, err := parseTopologySpreadConstraint("")
		require.Error(t, err)
	})

	t.Run("invalid maxSkew", func(t *testing.T) {
		_, err := parseTopologySpreadConstraint("zone:abc")
		require.Error(t, err)
	})

	t.Run("invalid whenUnsatisfiable", func(t *testing.T) {
		_, err := parseTopologySpreadConstraint("zone:1:invalid")
		require.Error(t, err)
	})
}

func TestParseSecretKeyRef(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		ref, err := parseSecretKeyRef("my-secret:api-key")
		require.NoError(t, err)
		assert.Equal(t, "my-secret", ref.GetName())
		assert.Equal(t, "api-key", ref.GetKey())
	})

	t.Run("missing key", func(t *testing.T) {
		_, err := parseSecretKeyRef("my-secret")
		require.Error(t, err)
	})

	t.Run("empty name", func(t *testing.T) {
		_, err := parseSecretKeyRef(":key")
		require.Error(t, err)
	})

	t.Run("empty key", func(t *testing.T) {
		_, err := parseSecretKeyRef("name:")
		require.Error(t, err)
	})
}

func TestSecretKeyRefString(t *testing.T) {
	assert.Equal(t, "name:key", secretKeyRefString(&commonv1.SecretKeyRef{Name: "name", Key: "key"}))
	assert.Empty(t, secretKeyRefString(nil))
}
