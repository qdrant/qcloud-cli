package output

import (
	"strings"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"
)

// PackageTier returns a concise label for a PackageTier.
func PackageTier(t bookingv1.PackageTier) string {
	return strings.TrimPrefix(t.String(), "PACKAGE_TIER_")
}
