package cluster_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	platformv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/platform/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestClusterIDCompletion(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.Server.ListClustersFunc = func(_ context.Context, _ *clusterv1.ListClustersRequest) (*clusterv1.ListClustersResponse, error) {
		return &clusterv1.ListClustersResponse{
			Items: []*clusterv1.Cluster{
				{Id: "cluster-abc", Name: "my-cluster"},
				{Id: "cluster-xyz", Name: "other-cluster"},
			},
		}, nil
	}

	stdout, _, err := testutil.Exec(t, env, "__complete", "cluster", "describe", "")
	require.NoError(t, err)
	assert.Contains(t, stdout, "cluster-abc")
	assert.Contains(t, stdout, "my-cluster")
	assert.Contains(t, stdout, "cluster-xyz")
}

func TestCloudProviderCompletion(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.PlatformServer.ListCloudProvidersFunc = func(_ context.Context, _ *platformv1.ListCloudProvidersRequest) (*platformv1.ListCloudProvidersResponse, error) {
		return &platformv1.ListCloudProvidersResponse{
			Items: []*platformv1.CloudProvider{
				{Id: "aws", Name: "Amazon Web Services"},
				{Id: "gcp", Name: "Google Cloud"},
			},
		}, nil
	}

	stdout, _, err := testutil.Exec(t, env, "__complete", "cluster", "list", "--cloud-provider", "")
	require.NoError(t, err)
	assert.Contains(t, stdout, "aws")
	assert.Contains(t, stdout, "Amazon Web Services")
	assert.Contains(t, stdout, "gcp")
}

func TestCloudRegionCompletion(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.PlatformServer.ListCloudProviderRegionsFunc = func(_ context.Context, req *platformv1.ListCloudProviderRegionsRequest) (*platformv1.ListCloudProviderRegionsResponse, error) {
		return &platformv1.ListCloudProviderRegionsResponse{
			Items: []*platformv1.CloudProviderRegion{
				{Id: "us-east-1", Name: "US East (N. Virginia)"},
			},
		}, nil
	}

	stdout, _, err := testutil.Exec(t, env, "__complete", "cluster", "list", "--cloud-provider", "aws", "--cloud-region", "")
	require.NoError(t, err)
	assert.Contains(t, stdout, "us-east-1")
	assert.Contains(t, stdout, "US East")
}

func TestCloudRegionCompletion_NoProvider(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	stdout, _, err := testutil.Exec(t, env, "__complete", "cluster", "list", "--cloud-region", "")
	require.NoError(t, err)
	// No completions when provider is not set.
	assert.NotContains(t, stdout, "us-east-1")
}
