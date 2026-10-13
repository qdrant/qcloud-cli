package cloudregion_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestCloudRegionCompletion_Describe(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessPlatformServer.ListCloudRegionsCalls.Returns(sampleRegions(), nil)

	stdout, _, err := testutil.Exec(t, env, "__complete", "serverless", "cloud-region", "describe", "")
	require.NoError(t, err)
	assert.Contains(t, stdout, "eu-central-1\tFrankfurt")
	assert.Contains(t, stdout, "us-east-1\tN. Virginia (unavailable)")

	req, ok := env.ServerlessPlatformServer.ListCloudRegionsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
}

func TestCloudRegionCompletion_StopsAfterFirstArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	stdout, _, err := testutil.Exec(t, env, "__complete", "serverless", "cloud-region", "describe", "eu-central-1", "")
	require.NoError(t, err)
	assert.NotContains(t, stdout, "eu-central-1")
	assert.Equal(t, 0, env.ServerlessPlatformServer.ListCloudRegionsCalls.Count())
}

func TestCloudRegionCompletion_SpaceCreateFlag(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessPlatformServer.ListCloudRegionsCalls.Returns(sampleRegions(), nil)

	stdout, _, err := testutil.Exec(t, env, "__complete", "serverless", "space", "create", "--cloud-region", "")
	require.NoError(t, err)
	assert.Contains(t, stdout, "eu-central-1")
	assert.Contains(t, stdout, "us-east-1")
}

func TestCloudRegionCompletion_SpaceListFlag(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessPlatformServer.ListCloudRegionsCalls.Returns(sampleRegions(), nil)

	stdout, _, err := testutil.Exec(t, env, "__complete", "serverless", "space", "list", "--cloud-region", "")
	require.NoError(t, err)
	assert.Contains(t, stdout, "eu-central-1")
	assert.Contains(t, stdout, "us-east-1")
}
