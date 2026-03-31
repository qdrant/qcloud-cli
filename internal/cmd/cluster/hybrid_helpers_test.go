package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"
)

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

func TestParseServiceType(t *testing.T) {
	tests := []struct {
		input   string
		want    clusterv1.ClusterServiceType
		wantErr bool
	}{
		{"cluster-ip", clusterv1.ClusterServiceType_CLUSTER_SERVICE_TYPE_CLUSTER_IP, false},
		{"node-port", clusterv1.ClusterServiceType_CLUSTER_SERVICE_TYPE_NODE_PORT, false},
		{"load-balancer", clusterv1.ClusterServiceType_CLUSTER_SERVICE_TYPE_LOAD_BALANCER, false},
		{"invalid", clusterv1.ClusterServiceType_CLUSTER_SERVICE_TYPE_UNSPECIFIED, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseServiceType(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
				assert.Equal(t, tt.input, serviceTypeString(got))
			}
		})
	}
}

func TestParseToleration(t *testing.T) {
	opEqual := clusterv1.TolerationOperator_TOLERATION_OPERATOR_EQUAL
	opExists := clusterv1.TolerationOperator_TOLERATION_OPERATOR_EXISTS
	effectNoSchedule := clusterv1.TolerationEffect_TOLERATION_EFFECT_NO_SCHEDULE
	effectPreferNoSchedule := clusterv1.TolerationEffect_TOLERATION_EFFECT_PREFER_NO_SCHEDULE
	effectNoExecute := clusterv1.TolerationEffect_TOLERATION_EFFECT_NO_EXECUTE

	t.Run("key=value:Effect", func(t *testing.T) {
		tol, err := parseToleration("env=prod:NoSchedule")
		require.NoError(t, err)
		assert.Equal(t, "env", tol.GetKey())
		assert.Equal(t, "prod", tol.GetValue())
		assert.Equal(t, opEqual, tol.GetOperator())
		assert.Equal(t, effectNoSchedule, tol.GetEffect())
	})

	t.Run("key=value no effect", func(t *testing.T) {
		tol, err := parseToleration("env=prod")
		require.NoError(t, err)
		assert.Equal(t, "env", tol.GetKey())
		assert.Equal(t, "prod", tol.GetValue())
		assert.Equal(t, opEqual, tol.GetOperator())
		assert.Nil(t, tol.Effect)
	})

	t.Run("key:Exists:Effect", func(t *testing.T) {
		tol, err := parseToleration("dedicated:Exists:NoExecute")
		require.NoError(t, err)
		assert.Equal(t, "dedicated", tol.GetKey())
		assert.Equal(t, opExists, tol.GetOperator())
		assert.Equal(t, effectNoExecute, tol.GetEffect())
	})

	t.Run("key:Exists no effect", func(t *testing.T) {
		tol, err := parseToleration("dedicated:Exists")
		require.NoError(t, err)
		assert.Equal(t, "dedicated", tol.GetKey())
		assert.Equal(t, opExists, tol.GetOperator())
		assert.Nil(t, tol.Effect)
	})

	t.Run("prefer-no-schedule effect", func(t *testing.T) {
		tol, err := parseToleration("tier=spot:PreferNoSchedule")
		require.NoError(t, err)
		assert.Equal(t, effectPreferNoSchedule, tol.GetEffect())
	})

	t.Run("no-execute alias", func(t *testing.T) {
		tol, err := parseToleration("node.kubernetes.io/not-ready:Exists:no-execute")
		require.NoError(t, err)
		assert.Equal(t, effectNoExecute, tol.GetEffect())
	})

	t.Run("empty key with equals", func(t *testing.T) {
		_, err := parseToleration("=value:NoSchedule")
		require.Error(t, err)
	})

	t.Run("empty key without equals", func(t *testing.T) {
		_, err := parseToleration(":Exists")
		require.Error(t, err)
	})

	t.Run("invalid effect", func(t *testing.T) {
		_, err := parseToleration("env=prod:Bogus")
		require.Error(t, err)
	})
}
