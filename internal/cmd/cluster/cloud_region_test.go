package cluster_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	platformv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/platform/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestListCloudRegions_RequiresCloudProvider(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	_, _, err := testutil.Exec(t, env, "cluster", "cloud-region", "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cloud-provider")
}

func TestListCloudRegions_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))
	t.Cleanup(env.Cleanup)

	env.PlatformServer.ListCloudProviderRegionsFunc = func(_ context.Context, req *platformv1.ListCloudProviderRegionsRequest) (*platformv1.ListCloudProviderRegionsResponse, error) {
		assert.Equal(t, "test-account-id", req.GetAccountId())
		assert.Equal(t, "aws", req.GetCloudProviderId())
		return &platformv1.ListCloudProviderRegionsResponse{
			Items: []*platformv1.CloudProviderRegion{
				{Id: "us-east-1", Name: "US East (N. Virginia)", Provider: "aws", Available: true},
				{Id: "eu-west-1", Name: "Europe (Ireland)", Provider: "aws", Available: false},
			},
		}, nil
	}

	stdout, _, err := testutil.Exec(t, env, "cluster", "cloud-region", "list", "--cloud-provider", "aws")
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
}
