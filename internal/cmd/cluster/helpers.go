package cluster

import (
	"fmt"
	"slices"
	"strings"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"
)

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

// undiffableFlags lists flags whose restart-prompt entry can only say
// "(changed)" because they hold complex nested values with no scalar diff.
var undiffableFlags = slices.Concat(
	[]string{
		"node-selector",
		"toleration",
		"topology-spread-constraint",
		"annotation",
		"pod-label",
		"service-annotation",
	},
	storageConfigFlags,
)

// hybridConfigFlags lists hybrid-cluster flags that trigger a rolling restart.
var hybridConfigFlags = []string{
	"service-type",
	"node-selector",
	"toleration",
	"topology-spread-constraint",
	"annotation",
	"pod-label",
	"service-annotation",
	"reserved-cpu-percentage",
	"reserved-memory-percentage",
	"database-storage-class",
	"snapshot-storage-class",
	"volume-snapshot-class",
	"volume-attributes-class",
}

func formatGiB(v float64) string {
	return fmt.Sprintf("%.2f GiB", v)
}

func formatMillicores(v float64) string {
	return fmt.Sprintf("%.0fm", v)
}

const (
	diskPerfBalanced      = "balanced"
	diskPerfCostOptimised = "cost-optimised"
	diskPerfPerformance   = "performance"
)

func parseDiskPerformance(s string) (commonv1.StorageTierType, error) {
	switch s {
	case diskPerfBalanced:
		return commonv1.StorageTierType_STORAGE_TIER_TYPE_BALANCED, nil
	case diskPerfCostOptimised:
		return commonv1.StorageTierType_STORAGE_TIER_TYPE_COST_OPTIMISED, nil
	case diskPerfPerformance:
		return commonv1.StorageTierType_STORAGE_TIER_TYPE_PERFORMANCE, nil
	default:
		return commonv1.StorageTierType_STORAGE_TIER_TYPE_UNSPECIFIED, fmt.Errorf("invalid disk performance %q: must be one of %s, %s, %s", s, diskPerfBalanced, diskPerfCostOptimised, diskPerfPerformance)
	}
}

func storageTierString(t commonv1.StorageTierType) string {
	switch t {
	case commonv1.StorageTierType_STORAGE_TIER_TYPE_BALANCED:
		return diskPerfBalanced
	case commonv1.StorageTierType_STORAGE_TIER_TYPE_COST_OPTIMISED:
		return diskPerfCostOptimised
	case commonv1.StorageTierType_STORAGE_TIER_TYPE_PERFORMANCE:
		return diskPerfPerformance
	default:
		return ""
	}
}

const (
	restartModeRolling   = "rolling"
	restartModeParallel  = "parallel"
	restartModeAutomatic = "automatic"
)

func parseRestartMode(s string) (clusterv1.ClusterConfigurationRestartPolicy, error) {
	switch s {
	case restartModeRolling:
		return clusterv1.ClusterConfigurationRestartPolicy_CLUSTER_CONFIGURATION_RESTART_POLICY_ROLLING, nil
	case restartModeParallel:
		return clusterv1.ClusterConfigurationRestartPolicy_CLUSTER_CONFIGURATION_RESTART_POLICY_PARALLEL, nil
	case restartModeAutomatic:
		return clusterv1.ClusterConfigurationRestartPolicy_CLUSTER_CONFIGURATION_RESTART_POLICY_AUTOMATIC, nil
	default:
		return clusterv1.ClusterConfigurationRestartPolicy_CLUSTER_CONFIGURATION_RESTART_POLICY_UNSPECIFIED,
			fmt.Errorf("invalid restart mode %q: must be one of %s, %s, %s", s, restartModeRolling, restartModeParallel, restartModeAutomatic)
	}
}

func restartPolicyString(p clusterv1.ClusterConfigurationRestartPolicy) string {
	switch p {
	case clusterv1.ClusterConfigurationRestartPolicy_CLUSTER_CONFIGURATION_RESTART_POLICY_ROLLING:
		return restartModeRolling
	case clusterv1.ClusterConfigurationRestartPolicy_CLUSTER_CONFIGURATION_RESTART_POLICY_PARALLEL:
		return restartModeParallel
	case clusterv1.ClusterConfigurationRestartPolicy_CLUSTER_CONFIGURATION_RESTART_POLICY_AUTOMATIC:
		return restartModeAutomatic
	default:
		return ""
	}
}

const (
	rebalanceStrategyByCount        = "by-count"
	rebalanceStrategyBySize         = "by-size"
	rebalanceStrategyByCountAndSize = "by-count-and-size"
)

func parseRebalanceStrategy(s string) (clusterv1.ClusterConfigurationRebalanceStrategy, error) {
	switch s {
	case rebalanceStrategyByCount:
		return clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_BY_COUNT, nil
	case rebalanceStrategyBySize:
		return clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_BY_SIZE, nil
	case rebalanceStrategyByCountAndSize:
		return clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_BY_COUNT_AND_SIZE, nil
	default:
		return clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_UNSPECIFIED,
			fmt.Errorf("invalid rebalance strategy %q: must be one of %s, %s, %s", s, rebalanceStrategyByCount, rebalanceStrategyBySize, rebalanceStrategyByCountAndSize)
	}
}

func rebalanceStrategyString(s clusterv1.ClusterConfigurationRebalanceStrategy) string {
	switch s {
	case clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_BY_COUNT:
		return rebalanceStrategyByCount
	case clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_BY_SIZE:
		return rebalanceStrategyBySize
	case clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_BY_COUNT_AND_SIZE:
		return rebalanceStrategyByCountAndSize
	default:
		return ""
	}
}

// applyTolerations applies toleration changes to an existing slice. A raw string
// of the form "key-" (no colon or equals) removes all tolerations with that key;
// any other format appends a new toleration. The input slice is not modified.
func applyTolerations(existing []*clusterv1.Toleration, raw []string) ([]*clusterv1.Toleration, error) {
	removeKeys := map[string]bool{}
	var adds []*clusterv1.Toleration

	for _, r := range raw {
		if strings.HasSuffix(r, "-") && !strings.ContainsAny(r, ":=") {
			removeKeys[strings.TrimSuffix(r, "-")] = true
		} else {
			tol, err := parseToleration(r)
			if err != nil {
				return nil, err
			}
			adds = append(adds, tol)
		}
	}

	result := make([]*clusterv1.Toleration, 0, len(existing)+len(adds))
	for _, t := range existing {
		if !removeKeys[t.GetKey()] {
			result = append(result, t)
		}
	}
	return append(result, adds...), nil
}

// applyTopologySpreadConstraints applies TSC changes to an existing slice. A raw
// string of the form "topologyKey-" (no colon) removes the constraint with that
// topologyKey; any other format adds or replaces (by topologyKey) a constraint.
// The input slice is not modified.
func applyTopologySpreadConstraints(existing []*commonv1.TopologySpreadConstraint, raw []string) ([]*commonv1.TopologySpreadConstraint, error) {
	removeKeys := map[string]bool{}
	var adds []*commonv1.TopologySpreadConstraint

	for _, r := range raw {
		if strings.HasSuffix(r, "-") && !strings.Contains(r, ":") {
			removeKeys[strings.TrimSuffix(r, "-")] = true
		} else {
			tsc, err := parseTopologySpreadConstraint(r)
			if err != nil {
				return nil, err
			}
			adds = append(adds, tsc)
		}
	}

	// Build set of topologyKeys being replaced so we don't keep the old entry.
	addKeys := map[string]bool{}
	for _, tsc := range adds {
		addKeys[tsc.TopologyKey] = true
	}

	result := make([]*commonv1.TopologySpreadConstraint, 0, len(existing)+len(adds))
	for _, t := range existing {
		if !removeKeys[t.TopologyKey] && !addKeys[t.TopologyKey] {
			result = append(result, t)
		}
	}
	return append(result, adds...), nil
}
