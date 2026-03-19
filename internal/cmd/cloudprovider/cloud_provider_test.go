package cloudprovider_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	platformv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/platform/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestListCloudProviders_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))

	var capturedReq *platformv1.ListCloudProvidersRequest
	env.PlatformServer.EXPECT().ListCloudProviders(mock.Anything, mock.Anything).
		Run(func(_ context.Context, req *platformv1.ListCloudProvidersRequest) { capturedReq = req }).
		Return(&platformv1.ListCloudProvidersResponse{
			Items: []*platformv1.CloudProvider{
				{Id: "aws", Name: "Amazon Web Services", Available: true},
				{Id: "gcp", Name: "Google Cloud", Available: false},
			},
		}, nil)

	stdout, _, err := testutil.Exec(t, env, "cloud-provider", "list")
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

	require.NotNil(t, capturedReq)
	assert.Equal(t, "test-account-id", capturedReq.GetAccountId())
}
