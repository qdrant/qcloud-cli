package space_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spacev1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestSpaceSuggestName_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.SuggestSpaceNameCalls.Returns(&spacev1.SuggestSpaceNameResponse{Name: "brave-falcon"}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "suggest-name")
	require.NoError(t, err)
	assert.Equal(t, "brave-falcon\n", stdout)

	req, ok := env.ServerlessSpaceServer.SuggestSpaceNameCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
}

func TestSpaceSuggestName_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.SuggestSpaceNameCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "suggest-name")
	require.Error(t, err)
}

func TestSpaceSuggestName_RejectsArgs(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "suggest-name", "extra")
	require.Error(t, err)
	assert.Equal(t, 0, env.ServerlessSpaceServer.SuggestSpaceNameCalls.Count())
}
