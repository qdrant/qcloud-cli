package cloudregion_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	platformv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/platform/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestListCloudRegions_RequiresCloudProvider(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "cloud-region", "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cloud-provider")
}

func TestListCloudRegions_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))

	var capturedReq *platformv1.ListCloudProviderRegionsRequest
	env.PlatformServer.EXPECT().ListCloudProviderRegions(mock.Anything, mock.Anything).
		Run(func(_ context.Context, req *platformv1.ListCloudProviderRegionsRequest) { capturedReq = req }).
		Return(&platformv1.ListCloudProviderRegionsResponse{
			Items: []*platformv1.CloudProviderRegion{
				{Id: "us-east-1", Name: "US East (N. Virginia)", Provider: "aws", Available: true},
				{Id: "eu-west-1", Name: "Europe (Ireland)", Provider: "aws", Available: false},
			},
		}, nil)

	stdout, _, err := testutil.Exec(t, env, "cloud-region", "list", "--cloud-provider", "aws")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ID")
	assert.Contains(t, stdout, "NAME")
	assert.Contains(t, stdout, "PROVIDER")
	assert.Contains(t, stdout, "AVAILABLE")
	assert.Contains(t, stdout, "us-east-1")
	assert.Contains(t, stdout, "US East (N. Virginia)")
	assert.Contains(t, stdout, "eu-west-1")
	assert.Contains(t, stdout, "Europe (Ireland)")
	assert.Contains(t, stdout, "true")
	assert.Contains(t, stdout, "false")

	require.NotNil(t, capturedReq)
	assert.Equal(t, "test-account-id", capturedReq.GetAccountId())
	assert.Equal(t, "aws", capturedReq.GetCloudProviderId())
}
