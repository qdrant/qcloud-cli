package cluster

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"
	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
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

// gpuFlagToMillicores reads the --gpu string flag and normalizes it to a millicore string.
// Returns ("", nil) when the flag is empty or unset.
func gpuFlagToMillicores(cmd *cobra.Command) (string, error) {
	gpu, _ := cmd.Flags().GetString("gpu")
	if gpu == "" {
		return "", nil
	}
	return normalizeMillicores(gpu)
}
