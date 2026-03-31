package cluster

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"

	"github.com/spf13/cobra"
)

// Flag groups used to guard struct initialization in create/update.

var collectionFlags = []string{
	"replication-factor",
	"write-consistency-factor",
	"vectors-on-disk",
}

var performanceFlags = []string{
	"async-scorer",
	"optimizer-cpu-budget",
}

var serviceFlags = []string{
	"enable-tls",
	"api-key-secret",
	"read-only-api-key-secret",
}

var tlsFlags = []string{
	"tls-cert-secret",
	"tls-key-secret",
}

var auditLoggingFlags = []string{
	"audit-logging",
	"audit-log-rotation",
	"audit-log-max-files",
	"audit-log-trust-forwarded-headers",
}

var storageConfigFlags = []string{
	"database-storage-class",
	"snapshot-storage-class",
	"volume-snapshot-class",
	"volume-attributes-class",
}

// allDBConfigFlags lists all flags that affect DatabaseConfiguration and trigger
// a rolling restart. This includes both universal and hybrid-only DB flags.
var allDBConfigFlags = slices.Concat(
	collectionFlags,
	performanceFlags,
	serviceFlags,
	tlsFlags,
	auditLoggingFlags,
	[]string{"db-log-level"},
)

// hybridOnlyFlags lists flags that are only valid when --cloud-provider is "hybrid".
var hybridOnlyFlags = []string{
	"env-id",
	"service-type",
	"node-selector",
	"annotation",
	"pod-label",
	"service-annotation",
	"reserved-cpu-percentage",
	"reserved-memory-percentage",
	"toleration",
	"topology-spread-constraint",
	"database-storage-class",
	"snapshot-storage-class",
	"volume-snapshot-class",
	"volume-attributes-class",
	"db-log-level",
	"enable-tls",
	"api-key-secret",
	"read-only-api-key-secret",
	"tls-cert-secret",
	"tls-key-secret",
	"cost-allocation-label",
}

// Service type

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

func serviceTypeString(t clusterv1.ClusterServiceType) string {
	switch t {
	case clusterv1.ClusterServiceType_CLUSTER_SERVICE_TYPE_CLUSTER_IP:
		return serviceTypeClusterIP
	case clusterv1.ClusterServiceType_CLUSTER_SERVICE_TYPE_NODE_PORT:
		return serviceTypeNodePort
	case clusterv1.ClusterServiceType_CLUSTER_SERVICE_TYPE_LOAD_BALANCER:
		return serviceTypeLoadBalancer
	default:
		return ""
	}
}

// Toleration

// parseToleration parses a toleration string in one of these forms:
//
//	key=value:Effect         (operator Equal, e.g. "env=prod:NoSchedule")
//	key:Exists:Effect        (operator Exists, e.g. "env:Exists:NoSchedule")
//	key:Exists               (operator Exists, no effect filter)
//	key=value                (operator Equal, no effect filter)
func parseToleration(s string) (*clusterv1.Toleration, error) {
	tol := &clusterv1.Toleration{}

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

// GPU type

const (
	gpuTypeNvidia = "nvidia"
	gpuTypeAMD    = "amd"
)

func parseGpuType(s string) (clusterv1.ClusterConfigurationGpuType, error) {
	switch s {
	case gpuTypeNvidia:
		return clusterv1.ClusterConfigurationGpuType_CLUSTER_CONFIGURATION_GPU_TYPE_NVIDIA, nil
	case gpuTypeAMD:
		return clusterv1.ClusterConfigurationGpuType_CLUSTER_CONFIGURATION_GPU_TYPE_AMD, nil
	default:
		return clusterv1.ClusterConfigurationGpuType_CLUSTER_CONFIGURATION_GPU_TYPE_UNSPECIFIED,
			fmt.Errorf("invalid GPU type %q: must be one of %s, %s", s, gpuTypeNvidia, gpuTypeAMD)
	}
}

func gpuTypeString(t clusterv1.ClusterConfigurationGpuType) string {
	switch t {
	case clusterv1.ClusterConfigurationGpuType_CLUSTER_CONFIGURATION_GPU_TYPE_NVIDIA:
		return gpuTypeNvidia
	case clusterv1.ClusterConfigurationGpuType_CLUSTER_CONFIGURATION_GPU_TYPE_AMD:
		return gpuTypeAMD
	default:
		return ""
	}
}

// Database log level

const (
	dbLogLevelTrace = "trace"
	dbLogLevelDebug = "debug"
	dbLogLevelInfo  = "info"
	dbLogLevelWarn  = "warn"
	dbLogLevelError = "error"
	dbLogLevelOff   = "off"
)

func parseDBLogLevel(s string) (clusterv1.DatabaseConfigurationLogLevel, error) {
	switch s {
	case dbLogLevelTrace:
		return clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_TRACE, nil
	case dbLogLevelDebug:
		return clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_DEBUG, nil
	case dbLogLevelInfo:
		return clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_INFO, nil
	case dbLogLevelWarn:
		return clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_WARN, nil
	case dbLogLevelError:
		return clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_ERROR, nil
	case dbLogLevelOff:
		return clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_OFF, nil
	default:
		return clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_UNSPECIFIED,
			fmt.Errorf("invalid DB log level %q: must be one of %s, %s, %s, %s, %s, %s", s, dbLogLevelTrace, dbLogLevelDebug, dbLogLevelInfo, dbLogLevelWarn, dbLogLevelError, dbLogLevelOff)
	}
}

func dbLogLevelString(l clusterv1.DatabaseConfigurationLogLevel) string {
	switch l {
	case clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_TRACE:
		return dbLogLevelTrace
	case clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_DEBUG:
		return dbLogLevelDebug
	case clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_INFO:
		return dbLogLevelInfo
	case clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_WARN:
		return dbLogLevelWarn
	case clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_ERROR:
		return dbLogLevelError
	case clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_OFF:
		return dbLogLevelOff
	default:
		return ""
	}
}

// Audit log rotation

const (
	auditLogRotationDaily  = "daily"
	auditLogRotationHourly = "hourly"
)

func parseAuditLogRotation(s string) (clusterv1.AuditLogRotation, error) {
	switch s {
	case auditLogRotationDaily:
		return clusterv1.AuditLogRotation_AUDIT_LOG_ROTATION_DAILY, nil
	case auditLogRotationHourly:
		return clusterv1.AuditLogRotation_AUDIT_LOG_ROTATION_HOURLY, nil
	default:
		return clusterv1.AuditLogRotation_AUDIT_LOG_ROTATION_UNSPECIFIED,
			fmt.Errorf("invalid audit log rotation %q: must be one of %s, %s", s, auditLogRotationDaily, auditLogRotationHourly)
	}
}

func auditLogRotationString(r clusterv1.AuditLogRotation) string {
	switch r {
	case clusterv1.AuditLogRotation_AUDIT_LOG_ROTATION_DAILY:
		return auditLogRotationDaily
	case clusterv1.AuditLogRotation_AUDIT_LOG_ROTATION_HOURLY:
		return auditLogRotationHourly
	default:
		return ""
	}
}

// Topology spread constraint
// Format: topologyKey[:maxSkew[:whenUnsatisfiable]]

func parseTopologySpreadConstraint(s string) (*commonv1.TopologySpreadConstraint, error) {
	parts := strings.SplitN(s, ":", 3)
	if parts[0] == "" {
		return nil, fmt.Errorf("invalid topology spread constraint %q: topologyKey must not be empty", s)
	}

	tsc := &commonv1.TopologySpreadConstraint{
		TopologyKey: parts[0],
	}

	if len(parts) >= 2 {
		skew, err := strconv.ParseInt(parts[1], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid topology spread constraint %q: maxSkew must be an integer", s)
		}
		v := int32(skew)
		tsc.MaxSkew = &v
	}

	if len(parts) == 3 {
		wu, err := parseWhenUnsatisfiable(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid topology spread constraint %q: %w", s, err)
		}
		tsc.WhenUnsatisfiable = wu.Enum()
	}

	return tsc, nil
}

func parseWhenUnsatisfiable(s string) (commonv1.TopologySpreadConstraintWhenUnsatisfiable, error) {
	switch strings.ToLower(s) {
	case "donotschedule", "do-not-schedule":
		return commonv1.TopologySpreadConstraintWhenUnsatisfiable_TOPOLOGY_SPREAD_CONSTRAINT_WHEN_UNSATISFIABLE_DO_NOT_SCHEDULE, nil
	case "scheduleanyway", "schedule-anyway":
		return commonv1.TopologySpreadConstraintWhenUnsatisfiable_TOPOLOGY_SPREAD_CONSTRAINT_WHEN_UNSATISFIABLE_SCHEDULE_ANYWAY, nil
	default:
		return commonv1.TopologySpreadConstraintWhenUnsatisfiable_TOPOLOGY_SPREAD_CONSTRAINT_WHEN_UNSATISFIABLE_UNSPECIFIED,
			fmt.Errorf("invalid whenUnsatisfiable %q: must be one of DoNotSchedule, ScheduleAnyway", s)
	}
}

// SecretKeyRef
// Format: secretName:key

func parseSecretKeyRef(s string) (*commonv1.SecretKeyRef, error) {
	name, key, ok := strings.Cut(s, ":")
	if !ok || name == "" || key == "" {
		return nil, fmt.Errorf("invalid secret key ref %q: must be in format 'secretName:key'", s)
	}
	return &commonv1.SecretKeyRef{Name: name, Key: key}, nil
}

func secretKeyRefString(ref *commonv1.SecretKeyRef) string {
	if ref == nil {
		return ""
	}
	return ref.GetName() + ":" + ref.GetKey()
}

// Completion functions for hybrid-specific flags

func serviceTypeCompletion() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{serviceTypeClusterIP, serviceTypeNodePort, serviceTypeLoadBalancer}, cobra.ShellCompDirectiveNoFileComp
	}
}

func dbLogLevelCompletion() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{dbLogLevelTrace, dbLogLevelDebug, dbLogLevelInfo, dbLogLevelWarn, dbLogLevelError, dbLogLevelOff}, cobra.ShellCompDirectiveNoFileComp
	}
}

func auditLogRotationCompletion() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{auditLogRotationDaily, auditLogRotationHourly}, cobra.ShellCompDirectiveNoFileComp
	}
}

func gpuTypeCompletion() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{gpuTypeNvidia, gpuTypeAMD}, cobra.ShellCompDirectiveNoFileComp
	}
}
