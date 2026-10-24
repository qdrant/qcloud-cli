package output

import (
	"fmt"
	"strings"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"

	"github.com/qdrant/qcloud-cli/internal/resource"
)

// PackageTier returns a concise label for a PackageTier.
func PackageTier(t bookingv1.PackageTier) string {
	return strings.TrimPrefix(t.String(), "PACKAGE_TIER_")
}

// FormatMillicents formats millicent pricing as a human-readable price string.
// 1 unit of currency = 100,000 millicents. Returns "free" for zero.
// currency should be an ISO 4217 code (e.g. "USD").
func FormatMillicents(mc int32, currency string) string {
	if mc == 0 {
		return "free"
	}

	return fmt.Sprintf("%.4f %s", float64(mc)/100_000.0, currency)
}

// VectorType returns a concise label for a VectorType.
func VectorType(t bookingv1.VectorType) string {
	return strings.TrimPrefix(t.String(), "VECTOR_TYPE_")
}

// ModelModality returns a concise label for a ModelModality.
func ModelModality(m bookingv1.ModelModality) string {
	return strings.TrimPrefix(m.String(), "MODEL_MODALITY_")
}

// PackageCPU returns a package's CPU as a (fractional) number of cores (e.g.
// "0.5", "4"). A value that cannot be parsed is returned as is.
func PackageCPU(rc *bookingv1.ResourceConfiguration) string {
	return Cores(rc.GetCpu())
}

// PackageRAM returns a package's RAM in canonical form (e.g. "16GiB").
// A value that cannot be parsed is returned as is.
func PackageRAM(rc *bookingv1.ResourceConfiguration) string {
	return byteQuantity(rc.GetRam(), "0")
}

// PackageDisk returns a package's included disk in canonical form (e.g. "64GiB").
// A value that cannot be parsed is returned as is.
func PackageDisk(rc *bookingv1.ResourceConfiguration) string {
	return byteQuantity(rc.GetDisk(), "0")
}

// PackageGPU returns a package's GPU as a (fractional) number of GPUs (e.g.
// "1"), or "n/a" for a package without a GPU. A value that cannot be parsed is
// returned as is.
func PackageGPU(rc *bookingv1.ResourceConfiguration) string {
	gpu := rc.GetGpu()
	if m, err := resource.ParseMillicores(gpu); gpu == "" || (err == nil && m == 0) {
		return "n/a"
	}

	return Cores(gpu)
}

// byteQuantity formats s canonically, returning zero for a value of zero, since
// ByteQuantity.String renders zero as "".
func byteQuantity(s, zero string) string {
	v, err := resource.ParseByteQuantity(s)
	if err != nil {
		return s
	}

	if v == 0 {
		return zero
	}

	return v.String()
}
