package quota_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	serverlessquotav1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/quota/v1"

	"github.com/qdrant/qcloud-cli/internal/resource"
	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func sampleQuota() *serverlessquotav1.GetQuotasResponse {
	return &serverlessquotav1.GetQuotasResponse{
		Quota: &serverlessquotav1.AccountQuota{
			AccountId:                       "test-account-id",
			MaxSpaces:                       5,
			UsedSpaces:                      2,
			MaxCollectionsPerSpace:          10,
			PlatformMaxSizePerCollection:    uint64(50 * resource.GiB),
			PlatformMaxWorkersPerCollection: 4,
		},
	}
}

func TestQuota_TextOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessQuotaServer.GetQuotasCalls.Returns(sampleQuota(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "quota")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Spaces:                            2 / 5")
	assert.Contains(t, stdout, "Max Collections per Space:         10")
	assert.Contains(t, stdout, "Max Size per Collection:           50GiB")
	assert.Contains(t, stdout, "Max Search Workers per Collection: 4")

	req, ok := env.ServerlessQuotaServer.GetQuotasCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
}

func TestQuota_Unlimited(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessQuotaServer.GetQuotasCalls.Returns(&serverlessquotav1.GetQuotasResponse{
		Quota: &serverlessquotav1.AccountQuota{AccountId: "test-account-id", UsedSpaces: 3},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "quota")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Spaces:                            3 / unlimited")
	assert.Contains(t, stdout, "Max Collections per Space:         unlimited")
	assert.Contains(t, stdout, "Max Size per Collection:           unlimited")
	assert.Contains(t, stdout, "Max Search Workers per Collection: unlimited")
}

func TestQuota_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessQuotaServer.GetQuotasCalls.Returns(sampleQuota(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "quota", "--json")
	require.NoError(t, err)

	var result struct {
		AccountID  string `json:"accountId"`
		MaxSpaces  string `json:"maxSpaces"`
		UsedSpaces string `json:"usedSpaces"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "test-account-id", result.AccountID)
	assert.Equal(t, "5", result.MaxSpaces)
	assert.Equal(t, "2", result.UsedSpaces)
}

func TestQuota_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessQuotaServer.GetQuotasCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "quota")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get serverless quota")
}

func TestQuota_UnexpectedArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "serverless", "quota", "extra")
	require.Error(t, err)
	assert.Equal(t, 0, env.ServerlessQuotaServer.GetQuotasCalls.Count())
}
