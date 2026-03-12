package cluster_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	platformv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/platform/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestListCloudProviders_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))
	t.Cleanup(env.Cleanup)

	env.PlatformServer.ListCloudProvidersFunc = func(_ context.Context, req *platformv1.ListCloudProvidersRequest) (*platformv1.ListCloudProvidersResponse, error) {
		assert.Equal(t, "test-account-id", req.GetAccountId())
		return &platformv1.ListCloudProvidersResponse{
			Items: []*platformv1.CloudProvider{
				{Id: "aws", Name: "Amazon Web Services", Available: true},
				{Id: "gcp", Name: "Google Cloud", Available: false},
			},
		}, nil
	}

	stdout, _, err := testutil.Exec(t, env, "cluster", "cloud-provider", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ID")
	assert.Contains(t, stdout, "NAME")
	assert.Contains(t, stdout, "AVAILABLE")
	assert.Contains(t, stdout, "aws")
	assert.Contains(t, stdout, "Amazon Web Services")
	assert.Contains(t, stdout, "true")
	assert.Contains(t, stdout, "gcp")
	assert.Contains(t, stdout, "Google Cloud")
	assert.Contains(t, stdout, "false")
}
