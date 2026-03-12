package cluster

import (
	"context"
	"fmt"
	"regexp"
	"strings"

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

// resolvePackageByName lists active packages and returns the first matching by name.
func resolvePackageByName(ctx context.Context, booking bookingv1.BookingServiceClient,
	accountID, cloudProvider, cloudRegion, name string) (*bookingv1.Package, error) {
	resp, err := booking.ListPackages(ctx, &bookingv1.ListPackagesRequest{
		AccountId:             accountID,
		CloudProviderId:       cloudProvider,
		CloudProviderRegionId: &cloudRegion,
		Statuses:              []bookingv1.PackageStatus{bookingv1.PackageStatus_PACKAGE_STATUS_ACTIVE},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}
	for _, p := range resp.GetItems() {
		if p.GetName() == name {
			return p, nil
		}
	}
	return nil, fmt.Errorf("package %q not found for provider=%s region=%s", name, cloudProvider, cloudRegion)
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
