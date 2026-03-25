package hybrid

import (
	"fmt"
	"strings"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	hybridv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/hybrid/v1"
)

func phaseString(p hybridv1.HybridCloudEnvironmentStatusPhase) string {
	return strings.TrimPrefix(p.String(), "HYBRID_CLOUD_ENVIRONMENT_STATUS_PHASE_")
}

func clusterCreationStatusString(s hybridv1.QdrantClusterCreationStatus) string {
	return strings.TrimPrefix(s.String(), "QDRANT_CLUSTER_CREATION_STATUS_")
}

func componentPhaseString(p hybridv1.HybridCloudEnvironmentComponentStatusPhase) string {
	return strings.TrimPrefix(p.String(), "HYBRID_CLOUD_ENVIRONMENT_COMPONENT_STATUS_PHASE_")
}

const (
	logLevelDebug = "debug"
	logLevelInfo  = "info"
	logLevelWarn  = "warn"
	logLevelError = "error"
)

func parseLogLevel(s string) (hybridv1.HybridCloudEnvironmentConfigurationLogLevel, error) {
	switch s {
	case logLevelDebug:
		return hybridv1.HybridCloudEnvironmentConfigurationLogLevel_HYBRID_CLOUD_ENVIRONMENT_CONFIGURATION_LOG_LEVEL_DEBUG, nil
	case logLevelInfo:
		return hybridv1.HybridCloudEnvironmentConfigurationLogLevel_HYBRID_CLOUD_ENVIRONMENT_CONFIGURATION_LOG_LEVEL_INFO, nil
	case logLevelWarn:
		return hybridv1.HybridCloudEnvironmentConfigurationLogLevel_HYBRID_CLOUD_ENVIRONMENT_CONFIGURATION_LOG_LEVEL_WARN, nil
	case logLevelError:
		return hybridv1.HybridCloudEnvironmentConfigurationLogLevel_HYBRID_CLOUD_ENVIRONMENT_CONFIGURATION_LOG_LEVEL_ERROR, nil
	default:
		return hybridv1.HybridCloudEnvironmentConfigurationLogLevel_HYBRID_CLOUD_ENVIRONMENT_CONFIGURATION_LOG_LEVEL_UNSPECIFIED,
			fmt.Errorf("invalid log level %q: must be one of %s, %s, %s, %s", s, logLevelDebug, logLevelInfo, logLevelWarn, logLevelError)
	}
}

func logLevelString(l hybridv1.HybridCloudEnvironmentConfigurationLogLevel) string {
	switch l {
	case hybridv1.HybridCloudEnvironmentConfigurationLogLevel_HYBRID_CLOUD_ENVIRONMENT_CONFIGURATION_LOG_LEVEL_DEBUG:
		return logLevelDebug
	case hybridv1.HybridCloudEnvironmentConfigurationLogLevel_HYBRID_CLOUD_ENVIRONMENT_CONFIGURATION_LOG_LEVEL_INFO:
		return logLevelInfo
	case hybridv1.HybridCloudEnvironmentConfigurationLogLevel_HYBRID_CLOUD_ENVIRONMENT_CONFIGURATION_LOG_LEVEL_WARN:
		return logLevelWarn
	case hybridv1.HybridCloudEnvironmentConfigurationLogLevel_HYBRID_CLOUD_ENVIRONMENT_CONFIGURATION_LOG_LEVEL_ERROR:
		return logLevelError
	default:
		return ""
	}
}

const (
	serviceTypeClusterIP    = "cluster-ip"
	serviceTypeNodePort     = "node-port"
	serviceTypeLoadBalancer = "load-balancer"
)

func parseServiceType(s string) (clusterv1.ClusterServiceType, error) {
	switch s {
	case serviceTypeClusterIP:
		return clusterv1.ClusterServiceType_CLUSTER_SERVICE_TYPE_CLUSTER_IP, nil
	case serviceTypeNodePort:
		return clusterv1.ClusterServiceType_CLUSTER_SERVICE_TYPE_NODE_PORT, nil
	case serviceTypeLoadBalancer:
		return clusterv1.ClusterServiceType_CLUSTER_SERVICE_TYPE_LOAD_BALANCER, nil
	default:
		return clusterv1.ClusterServiceType_CLUSTER_SERVICE_TYPE_UNSPECIFIED,
			fmt.Errorf("invalid service type %q: must be one of %s, %s, %s", s, serviceTypeClusterIP, serviceTypeNodePort, serviceTypeLoadBalancer)
	}
}

// parseToleration parses a toleration string in one of these forms:
//
//	key=value:Effect         (operator Equal, e.g. "env=prod:NoSchedule")
//	key:Exists:Effect        (operator Exists, e.g. "env:Exists:NoSchedule")
//	key:Exists               (operator Exists, no effect filter)
//	key=value                (operator Equal, no effect filter)
func parseToleration(s string) (*clusterv1.Toleration, error) {
	tol := &clusterv1.Toleration{}

	// Determine operator: if the string contains "=", treat as Equal; if it
	// contains ":Exists", treat as Exists.
	if strings.Contains(s, "=") {
		// Format: key=value[:Effect]
		parts := strings.SplitN(s, ":", 2)
		kv := parts[0]
		key, value, _ := strings.Cut(kv, "=")
		if key == "" {
			return nil, fmt.Errorf("invalid toleration %q: key must not be empty", s)
		}
		tol.Key = &key
		tol.Value = &value
		op := clusterv1.TolerationOperator_TOLERATION_OPERATOR_EQUAL
		tol.Operator = &op
		if len(parts) == 2 {
			effect, err := parseTolerationEffect(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid toleration %q: %w", s, err)
			}
			tol.Effect = &effect
		}
	} else {
		// Format: key:Exists[:Effect] or key[:Effect]
		parts := strings.SplitN(s, ":", 3)
		key := parts[0]
		if key == "" {
			return nil, fmt.Errorf("invalid toleration %q: key must not be empty", s)
		}
		tol.Key = &key

		if len(parts) >= 2 {
			switch strings.ToLower(parts[1]) {
			case "exists":
				op := clusterv1.TolerationOperator_TOLERATION_OPERATOR_EXISTS
				tol.Operator = &op
				if len(parts) == 3 {
					effect, err := parseTolerationEffect(parts[2])
					if err != nil {
						return nil, fmt.Errorf("invalid toleration %q: %w", s, err)
					}
					tol.Effect = &effect
				}
			default:
				// Treat second part as effect (Equal implied)
				effect, err := parseTolerationEffect(parts[1])
				if err != nil {
					return nil, fmt.Errorf("invalid toleration %q: %w", s, err)
				}
				tol.Effect = &effect
			}
		}
	}

	return tol, nil
}

func parseTolerationEffect(s string) (clusterv1.TolerationEffect, error) {
	switch strings.ToLower(s) {
	case "noschedule", "no-schedule":
		return clusterv1.TolerationEffect_TOLERATION_EFFECT_NO_SCHEDULE, nil
	case "prefernoschedule", "prefer-no-schedule":
		return clusterv1.TolerationEffect_TOLERATION_EFFECT_PREFER_NO_SCHEDULE, nil
	case "noexecute", "no-execute":
		return clusterv1.TolerationEffect_TOLERATION_EFFECT_NO_EXECUTE, nil
	default:
		return clusterv1.TolerationEffect_TOLERATION_EFFECT_UNSPECIFIED,
			fmt.Errorf("invalid effect %q: must be one of NoSchedule, PreferNoSchedule, NoExecute", s)
	}
}
