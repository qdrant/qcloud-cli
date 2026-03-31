package clusterutil_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/clusterutil"
	"github.com/qdrant/qcloud-cli/internal/resource"
)

func TestCalculateAdditionalDisk_LargerThanPackage(t *testing.T) {
	pkg := &bookingv1.Package{
		ResourceConfiguration: &bookingv1.ResourceConfiguration{Disk: "100GiB"},
	}
	requested, err := resource.ParseByteQuantity("200GiB")
	require.NoError(t, err)

	additional, err := clusterutil.CalculateAdditionalDisk(requested, pkg)
	require.NoError(t, err)
	assert.Equal(t, uint32(100), additional)
}

func TestCalculateAdditionalDisk_SmallerThanPackage(t *testing.T) {
	pkg := &bookingv1.Package{
		ResourceConfiguration: &bookingv1.ResourceConfiguration{Disk: "100GiB"},
	}
	requested, err := resource.ParseByteQuantity("50GiB")
	require.NoError(t, err)

	additional, err := clusterutil.CalculateAdditionalDisk(requested, pkg)
	require.NoError(t, err)
	assert.Equal(t, uint32(0), additional)
}

func TestCalculateAdditionalDisk_EqualToPackage(t *testing.T) {
	pkg := &bookingv1.Package{
		ResourceConfiguration: &bookingv1.ResourceConfiguration{Disk: "100GiB"},
	}
	requested, err := resource.ParseByteQuantity("100GiB")
	require.NoError(t, err)

	additional, err := clusterutil.CalculateAdditionalDisk(requested, pkg)
	require.NoError(t, err)
	assert.Equal(t, uint32(0), additional)
}

func TestCalculateAdditionalDisk_EmptyPackageDisk(t *testing.T) {
	pkg := &bookingv1.Package{
		ResourceConfiguration: &bookingv1.ResourceConfiguration{},
	}
	requested, err := resource.ParseByteQuantity("50GiB")
	require.NoError(t, err)

	additional, err := clusterutil.CalculateAdditionalDisk(requested, pkg)
	require.NoError(t, err)
	assert.Equal(t, uint32(0), additional)
}
