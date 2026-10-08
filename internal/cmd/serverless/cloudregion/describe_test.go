package cloudregion_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	serverlessplatformv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/platform/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func sampleRegion() *serverlessplatformv1.GetCloudRegionResponse {
	return &serverlessplatformv1.GetCloudRegionResponse{
		Region: &serverlessplatformv1.CloudRegion{
			Id:                    "eu-central-1",
			Name:                  "Frankfurt",
			Available:             true,
			GeographicalSubRegion: new("Western Europe"),
			CountryIsoCode:        new("DE"),
		},
	}
}

func TestCloudRegionDescribe_TextOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessPlatformServer.GetCloudRegionCalls.Returns(sampleRegion(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "cloud-region", "describe", "eu-central-1")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ID:         eu-central-1")
	assert.Contains(t, stdout, "Name:       Frankfurt")
	assert.Contains(t, stdout, "Available:  yes")
	assert.Contains(t, stdout, "Sub-region: Western Europe")
	assert.Contains(t, stdout, "Country:    DE")

	req, ok := env.ServerlessPlatformServer.GetCloudRegionCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "eu-central-1", req.GetCloudRegionId())
}

func TestCloudRegionDescribe_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessPlatformServer.GetCloudRegionCalls.Returns(sampleRegion(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "cloud-region", "describe", "eu-central-1", "--json")
	require.NoError(t, err)

	var result struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		CountryIsoCode string `json:"countryIsoCode"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "eu-central-1", result.ID)
	assert.Equal(t, "Frankfurt", result.Name)
	assert.Equal(t, "DE", result.CountryIsoCode)
}

func TestCloudRegionDescribe_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessPlatformServer.GetCloudRegionCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "cloud-region", "describe", "eu-central-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get cloud region")
}

func TestCloudRegionDescribe_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "serverless", "cloud-region", "describe")
	require.Error(t, err)
	assert.Equal(t, 0, env.ServerlessPlatformServer.GetCloudRegionCalls.Count())
}

func TestCloudRegionDescribe_UnsetOptionalFields(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessPlatformServer.GetCloudRegionCalls.Returns(&serverlessplatformv1.GetCloudRegionResponse{
		Region: &serverlessplatformv1.CloudRegion{Id: "eu-central-1", Name: "Frankfurt"},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "cloud-region", "describe", "eu-central-1")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Available:  no")
	assert.Contains(t, stdout, "Sub-region: not set")
	assert.Contains(t, stdout, "Country:    not set")
}
