package cluster_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestListPackages_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{
		Items: []*bookingv1.Package{
			{
				Id:   "pkg-123",
				Name: "starter",
				Tier: bookingv1.PackageTier_PACKAGE_TIER_STANDARD,
				ResourceConfiguration: &bookingv1.ResourceConfiguration{
					Ram:  "1GiB",
					Cpu:  "0.5",
					Disk: "10GiB",
					Gpu:  new("1000m"),
				},
				MultiAz:             true,
				UnitIntPricePerHour: 5000,
				Currency:            "USD",
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env,
		"cluster", "package", "list",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "NAME")
	assert.Contains(t, stdout, "TIER")
	assert.Contains(t, stdout, "GPU")
	assert.Contains(t, stdout, "MULTI-AZ")
	assert.Contains(t, stdout, "PRICE/HR")
	assert.Contains(t, stdout, "starter")
	assert.Contains(t, stdout, "STANDARD")
	assert.Contains(t, stdout, "1GiB")
	assert.Contains(t, stdout, "1000m")
	assert.Contains(t, stdout, "yes")
	assert.Contains(t, stdout, "0.0500 USD")
}

func TestListPackages_FreePackage(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{
		Items: []*bookingv1.Package{
			{
				Id:                  "pkg-free",
				Name:                "free",
				UnitIntPricePerHour: 0,
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env,
		"cluster", "package", "list",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "free")
}
