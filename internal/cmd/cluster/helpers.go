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
