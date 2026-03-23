package cluster

import (
	"fmt"
	"regexp"
	"strings"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"
	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"
)

func phaseString(phase clusterv1.ClusterPhase) string {
	return strings.TrimPrefix(phase.String(), "CLUSTER_PHASE_")
}

// isUUID returns true if s looks like a UUID.
func isUUID(s string) bool {
	matched, _ := regexp.MatchString(
		`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, s)
	return matched
}

func nodeStateString(state clusterv1.ClusterNodeState) string {
	return strings.TrimPrefix(state.String(), "CLUSTER_NODE_STATE_")
}

func packageTierString(tier bookingv1.PackageTier) string {
	return strings.TrimPrefix(tier.String(), "PACKAGE_TIER_")
}

func boolToMark(v bool) string {
	if v {
		return "yes"
	}
	return ""
}

func boolToYesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func formatGiB(v float64) string {
	return fmt.Sprintf("%.2f GiB", v)
}

func formatMillicores(v float64) string {
	return fmt.Sprintf("%.0fm", v)
}

// formatMillicents formats millicent pricing as a human-readable price string.
// 1 unit of currency = 100,000 millicents. Returns "free" for zero.
// currency should be an ISO 4217 code (e.g. "USD").
func formatMillicents(mc int32, currency string) string {
	if mc == 0 {
		return "free"
	}
	return fmt.Sprintf("%.4f %s", float64(mc)/100_000.0, currency)
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
		return clusterv1.ClusterConfigurationRestartPolicy_CLUSTER_CONFIGURATION_RESTART_POLICY_UNSPECIFIED, fmt.Errorf("invalid restart mode %q: must be one of %s, %s, %s", s, restartModeRolling, restartModeParallel, restartModeAutomatic)
	}
}

const (
	rebalanceByCount        = "by-count"
	rebalanceBySize         = "by-size"
	rebalanceByCountAndSize = "by-count-and-size"
)

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

func rebalanceStrategyString(s clusterv1.ClusterConfigurationRebalanceStrategy) string {
	switch s {
	case clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_BY_COUNT:
		return rebalanceByCount
	case clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_BY_SIZE:
		return rebalanceBySize
	case clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_BY_COUNT_AND_SIZE:
		return rebalanceByCountAndSize
	default:
		return ""
	}
}

func parseRebalanceStrategy(s string) (clusterv1.ClusterConfigurationRebalanceStrategy, error) {
	switch s {
	case rebalanceByCount:
		return clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_BY_COUNT, nil
	case rebalanceBySize:
		return clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_BY_SIZE, nil
	case rebalanceByCountAndSize:
		return clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_BY_COUNT_AND_SIZE, nil
	default:
		return clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_UNSPECIFIED, fmt.Errorf("invalid rebalance strategy %q: must be one of %s, %s, %s", s, rebalanceByCount, rebalanceBySize, rebalanceByCountAndSize)
	}
}
