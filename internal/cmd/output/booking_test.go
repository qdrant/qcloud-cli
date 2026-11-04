package output_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
)

func TestPackageResources_Canonical(t *testing.T) {
	rc := &bookingv1.ResourceConfiguration{Cpu: "4000m", Ram: "16Gi", Disk: "64Gi", Gpu: new("1000m")}

	assert.Equal(t, "4", output.PackageCPU(rc))
	assert.Equal(t, "16GiB", output.PackageRAM(rc))
	assert.Equal(t, "64GiB", output.PackageDisk(rc))
	assert.Equal(t, "1", output.PackageGPU(rc))
}

func TestPackageResources_FractionalCPU(t *testing.T) {
	rc := &bookingv1.ResourceConfiguration{Cpu: "0.5"}

	assert.Equal(t, "0.5", output.PackageCPU(rc))
}

func TestPackageResources_NoGPU(t *testing.T) {
	assert.Equal(t, "n/a", output.PackageGPU(&bookingv1.ResourceConfiguration{}))
}

func TestPackageResources_Zero(t *testing.T) {
	rc := &bookingv1.ResourceConfiguration{Cpu: "0", Ram: "0", Disk: "0Gi", Gpu: new("0")}

	assert.Equal(t, "0", output.PackageCPU(rc))
	assert.Equal(t, "0", output.PackageRAM(rc))
	assert.Equal(t, "0", output.PackageDisk(rc))
	assert.Equal(t, "n/a", output.PackageGPU(rc))
}

func TestPackageResources_UnparsableReturnedAsIs(t *testing.T) {
	rc := &bookingv1.ResourceConfiguration{Cpu: "lots", Ram: "plenty", Disk: "huge"}

	assert.Equal(t, "lots", output.PackageCPU(rc))
	assert.Equal(t, "plenty", output.PackageRAM(rc))
	assert.Equal(t, "huge", output.PackageDisk(rc))
}

func TestPackageResources_NilConfiguration(t *testing.T) {
	assert.Empty(t, output.PackageCPU(nil))
	assert.Empty(t, output.PackageRAM(nil))
	assert.Empty(t, output.PackageDisk(nil))
	assert.Equal(t, "n/a", output.PackageGPU(nil))
}
