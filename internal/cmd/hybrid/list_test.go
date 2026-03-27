package hybrid_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hybridv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/hybrid/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestHybridList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.HybridServer.ListHybridCloudEnvironmentsCalls.Returns(&hybridv1.ListHybridCloudEnvironmentsResponse{
		Items: []*hybridv1.HybridCloudEnvironment{
			{
				Id:   "env-1",
				Name: "my-env",
				Status: &hybridv1.HybridCloudEnvironmentStatus{
					Phase:         hybridv1.HybridCloudEnvironmentStatusPhase_HYBRID_CLOUD_ENVIRONMENT_STATUS_PHASE_READY,
					NumberOfNodes: 3,
				},
			},
			{
				Id:   "env-2",
				Name: "other-env",
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "list")
	require.NoError(t, err)

	assert.Contains(t, stdout, "env-1")
	assert.Contains(t, stdout, "my-env")
	assert.Contains(t, stdout, "READY")
	assert.Contains(t, stdout, "3")
	assert.Contains(t, stdout, "env-2")
	assert.Contains(t, stdout, "other-env")
}

func TestHybridList_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.HybridServer.ListHybridCloudEnvironmentsCalls.Returns(&hybridv1.ListHybridCloudEnvironmentsResponse{
		Items: []*hybridv1.HybridCloudEnvironment{
			{Id: "env-json", Name: "json-env"},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "list", "--json")
	require.NoError(t, err)

	assert.Contains(t, stdout, `"id"`)
	assert.Contains(t, stdout, `"env-json"`)
	assert.Contains(t, stdout, `"json-env"`)
}

func TestHybridList_Empty(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.HybridServer.ListHybridCloudEnvironmentsCalls.Returns(&hybridv1.ListHybridCloudEnvironmentsResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "list")
	require.NoError(t, err)

	assert.Contains(t, stdout, "ID")
	assert.Contains(t, stdout, "NAME")
}

func TestHybridList_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.HybridServer.ListHybridCloudEnvironmentsCalls.Returns(nil, fmt.Errorf("server error"))

	_, _, err := testutil.Exec(t, env, "hybrid", "list")
	require.Error(t, err)
}
