package hybrid_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hybridv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/hybrid/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestHybridDescribe_FullOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	ns := "qdrant-ns"
	dbClass := "fast-ssd"
	env.HybridServer.GetHybridCloudEnvironmentCalls.Returns(&hybridv1.GetHybridCloudEnvironmentResponse{
		HybridCloudEnvironment: &hybridv1.HybridCloudEnvironment{
			Id:   "env-abc",
			Name: "my-env",
			Status: &hybridv1.HybridCloudEnvironmentStatus{
				Phase:         hybridv1.HybridCloudEnvironmentStatusPhase_HYBRID_CLOUD_ENVIRONMENT_STATUS_PHASE_READY,
				NumberOfNodes: 5,
			},
			Configuration: &hybridv1.HybridCloudEnvironmentConfiguration{
				Namespace:            ns,
				DatabaseStorageClass: &dbClass,
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "describe", "env-abc")
	require.NoError(t, err)

	assert.Contains(t, stdout, "env-abc")
	assert.Contains(t, stdout, "my-env")
	assert.Contains(t, stdout, "READY")
	assert.Contains(t, stdout, "qdrant-ns")
	assert.Contains(t, stdout, "fast-ssd")
	assert.Contains(t, stdout, "5")
}

func TestHybridDescribe_MinimalOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.HybridServer.GetHybridCloudEnvironmentCalls.Returns(&hybridv1.GetHybridCloudEnvironmentResponse{
		HybridCloudEnvironment: &hybridv1.HybridCloudEnvironment{
			Id:   "env-min",
			Name: "minimal-env",
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "describe", "env-min")
	require.NoError(t, err)

	assert.Contains(t, stdout, "env-min")
	assert.Contains(t, stdout, "minimal-env")
}

func TestHybridDescribe_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "hybrid", "describe")
	require.Error(t, err)
}

func TestHybridDescribe_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.HybridServer.GetHybridCloudEnvironmentCalls.Returns(nil, fmt.Errorf("not found"))

	_, _, err := testutil.Exec(t, env, "hybrid", "describe", "env-abc")
	require.Error(t, err)
}
