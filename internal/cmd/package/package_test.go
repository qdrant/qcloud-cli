package packagecmd_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"
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
					Cpu:  "500m",
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
		"package", "list",
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
	assert.Contains(t, stdout, "0.5")
	assert.NotContains(t, stdout, "500m")
	assert.NotContains(t, stdout, "1000m")
	assert.Contains(t, stdout, "10GiB")
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
		"package", "list",
		"--cloud-provider", "aws",
		"--cloud-region", "us-east-1",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "free")
}

func TestListPackages_HybridNoRegion(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{
		Items: []*bookingv1.Package{
			{
				Id:   "pkg-hybrid",
				Name: "hybrid-starter",
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env,
		"package", "list",
		"--cloud-provider", "hybrid",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "hybrid-starter")
}

func TestListPackages_NonHybridRequiresRegion(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env,
		"package", "list",
		"--cloud-provider", "aws",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--cloud-region is required")
}

func TestListPackages_AuthenticatedRequest(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("acc-123"))

	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{
		Items: []*bookingv1.Package{
			{Id: "pkg-single", Name: "gpx1", Tier: bookingv1.PackageTier_PACKAGE_TIER_PREMIUM},
			{Id: "pkg-multi", Name: "gpx1", Tier: bookingv1.PackageTier_PACKAGE_TIER_PREMIUM, MultiAz: true},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env,
		"package", "list",
		"--cloud-provider", "aws",
		"--cloud-region", "eu-central-1",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "pkg-single")
	assert.Contains(t, stdout, "pkg-multi")
	assert.Contains(t, stdout, "PREMIUM")
	assert.Equal(t, 0, env.BookingServer.ListGlobalPackagesCalls.Count())

	req, ok := env.BookingServer.ListPackagesCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "acc-123", req.GetAccountId())
	assert.Equal(t, "aws", req.GetCloudProviderId())
	assert.Equal(t, "eu-central-1", req.GetCloudProviderRegionId())
	assert.Nil(t, req.MultiAz, "multi-AZ packages must not be filtered out")
	assert.Nil(t, req.Tier, "the account's tier must be used")
}

func TestListPackages_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{
		Items: []*bookingv1.Package{
			{Id: "pkg-multi", Name: "gpx1", MultiAz: true},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env,
		"package", "list",
		"--cloud-provider", "aws",
		"--cloud-region", "eu-central-1",
		"--json",
	)
	require.NoError(t, err)

	var resp struct {
		Items []struct {
			ID      string `json:"id"`
			MultiAz bool   `json:"multiAz"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &resp))
	require.Len(t, resp.Items, 1)
	assert.Equal(t, "pkg-multi", resp.Items[0].ID)
	assert.True(t, resp.Items[0].MultiAz)
}

func TestListPackages_UnauthenticatedUsesGlobalPackages(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAPIKey(""))

	env.BookingServer.ListGlobalPackagesCalls.Returns(&bookingv1.ListGlobalPackagesResponse{
		Items: []*bookingv1.Package{
			{Id: "pkg-public", Name: "gpx1", Tier: bookingv1.PackageTier_PACKAGE_TIER_STANDARD},
		},
	}, nil)

	stdout, stderr, err := testutil.Exec(t, env,
		"package", "list",
		"--cloud-provider", "aws",
		"--cloud-region", "eu-central-1",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "pkg-public")
	assert.NotContains(t, stderr, "no account ID configured")
	assert.Equal(t, 0, env.BookingServer.ListPackagesCalls.Count())

	req, ok := env.BookingServer.ListGlobalPackagesCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "eu-central-1", req.GetCloudProviderRegionId())
}

func TestListPackages_APIKeyWithoutAccountIDHints(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID(""))

	env.BookingServer.ListGlobalPackagesCalls.Returns(&bookingv1.ListGlobalPackagesResponse{
		Items: []*bookingv1.Package{
			{Id: "pkg-public", Name: "gpx1", Tier: bookingv1.PackageTier_PACKAGE_TIER_STANDARD},
		},
	}, nil)

	stdout, stderr, err := testutil.Exec(t, env,
		"package", "list",
		"--cloud-provider", "aws",
		"--cloud-region", "eu-central-1",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "pkg-public")
	assert.NotContains(t, stdout, "Hint")
	assert.Contains(t, stderr, "no account ID configured")
	assert.Contains(t, stderr, "QDRANT_CLOUD_ACCOUNT_ID")
	assert.Equal(t, 0, env.BookingServer.ListPackagesCalls.Count())
	assert.Equal(t, 1, env.BookingServer.ListGlobalPackagesCalls.Count())
}

func TestListPackages_BackendError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.BookingServer.ListPackagesCalls.Returns(nil, fmt.Errorf("boom"))

	_, _, err := testutil.Exec(t, env,
		"package", "list",
		"--cloud-provider", "aws",
		"--cloud-region", "eu-central-1",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list packages")
}

func TestListPackages_UnauthenticatedBackendError(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAPIKey(""))

	env.BookingServer.ListGlobalPackagesCalls.Returns(nil, fmt.Errorf("boom"))

	_, _, err := testutil.Exec(t, env,
		"package", "list",
		"--cloud-provider", "aws",
		"--cloud-region", "eu-central-1",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list packages")
	assert.Equal(t, 0, env.BookingServer.ListPackagesCalls.Count())
}

func TestListPackages_OAuthLoginRequest(t *testing.T) {
	env := testutil.NewBareTestEnv(t, testutil.WithOAuthServer(testutil.NewFakeOAuthServer()))
	env.AccountServer.ListAccountsCalls.Returns(&accountv1.ListAccountsResponse{
		Items: []*accountv1.Account{{Id: "acc-oauth", Name: "OAuth Account"}},
	}, nil)
	cfg := testutil.WriteContextConfigFile(t, t.TempDir(), "", nil)

	_, _, err := testutil.Exec(t, env, "auth", "login", "--config", cfg,
		"--endpoint", "grpc.development-cloud.qdrant.io:443", "--client-id", "cli-app")
	require.NoError(t, err)

	env.BookingServer.ListPackagesCalls.Returns(&bookingv1.ListPackagesResponse{
		Items: []*bookingv1.Package{
			{Id: "pkg-multi", Name: "gpx1", Tier: bookingv1.PackageTier_PACKAGE_TIER_PREMIUM, MultiAz: true},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env,
		"package", "list",
		"--config", cfg,
		"--cloud-provider", "aws",
		"--cloud-region", "eu-central-1",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "pkg-multi")
	assert.Equal(t, 0, env.BookingServer.ListGlobalPackagesCalls.Count())

	req, ok := env.BookingServer.ListPackagesCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "acc-oauth", req.GetAccountId())
}
