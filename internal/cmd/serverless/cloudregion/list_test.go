package cloudregion_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	serverlessplatformv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/platform/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func sampleRegions() *serverlessplatformv1.ListCloudRegionsResponse {
	return &serverlessplatformv1.ListCloudRegionsResponse{
		Items: []*serverlessplatformv1.CloudRegion{
			{Id: "eu-central-1", Name: "Frankfurt", Available: true, GeographicalSubRegion: new("Western Europe"), CountryIsoCode: new("DE")},
			{Id: "us-east-1", Name: "N. Virginia", Available: false, GeographicalSubRegion: new("Northern America"), CountryIsoCode: new("US")},
		},
	}
}

func TestCloudRegionList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessPlatformServer.ListCloudRegionsCalls.Returns(sampleRegions(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "cloud-region", "list")
	require.NoError(t, err)

	for _, h := range []string{"ID", "NAME", "AVAILABLE", "SUB-REGION", "COUNTRY"} {
		assert.Contains(t, stdout, h)
	}

	assert.Contains(t, stdout, "eu-central-1")
	assert.Contains(t, stdout, "Frankfurt")
	assert.Contains(t, stdout, "Western Europe")
	assert.Contains(t, stdout, "DE")
	assert.Contains(t, stdout, "us-east-1")
	assert.Contains(t, stdout, "yes")
	assert.Contains(t, stdout, "no")

	req, ok := env.ServerlessPlatformServer.ListCloudRegionsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
}

func TestCloudRegionList_NoHeaders(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessPlatformServer.ListCloudRegionsCalls.Returns(sampleRegions(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "cloud-region", "list", "--no-headers")
	require.NoError(t, err)
	assert.NotContains(t, stdout, "SUB-REGION")
	assert.Contains(t, stdout, "eu-central-1")
}

func TestCloudRegionList_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessPlatformServer.ListCloudRegionsCalls.Returns(sampleRegions(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "cloud-region", "list", "--json")
	require.NoError(t, err)

	var result struct {
		Items []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Available bool   `json:"available"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Items, 2)
	assert.Equal(t, "eu-central-1", result.Items[0].ID)
	assert.Equal(t, "Frankfurt", result.Items[0].Name)
	assert.True(t, result.Items[0].Available)
}

func TestCloudRegionList_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessPlatformServer.ListCloudRegionsCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "cloud-region", "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list cloud regions")
}

func TestCloudRegionList_UnexpectedArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "serverless", "cloud-region", "list", "extra")
	require.Error(t, err)
	assert.Equal(t, 0, env.ServerlessPlatformServer.ListCloudRegionsCalls.Count())
}
