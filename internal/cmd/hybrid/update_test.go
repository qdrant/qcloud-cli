package hybrid_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hybridv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/hybrid/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func setupHybridUpdateHandlers(env *testutil.TestEnv) {
	env.HybridServer.GetHybridCloudEnvironmentCalls.Always(func(_ context.Context, req *hybridv1.GetHybridCloudEnvironmentRequest) (*hybridv1.GetHybridCloudEnvironmentResponse, error) {
		return &hybridv1.GetHybridCloudEnvironmentResponse{
			HybridCloudEnvironment: &hybridv1.HybridCloudEnvironment{
				Id:   req.GetHybridCloudEnvironmentId(),
				Name: "existing-env",
			},
		}, nil
	})
	env.HybridServer.UpdateHybridCloudEnvironmentCalls.Always(func(_ context.Context, req *hybridv1.UpdateHybridCloudEnvironmentRequest) (*hybridv1.UpdateHybridCloudEnvironmentResponse, error) {
		return &hybridv1.UpdateHybridCloudEnvironmentResponse{
			HybridCloudEnvironment: req.GetHybridCloudEnvironment(),
		}, nil
	})
}

func TestHybridUpdate_Name(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupHybridUpdateHandlers(env)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "update", "env-abc", "--name", "new-name")
	require.NoError(t, err)
	assert.Contains(t, stdout, "updated")

	req, ok := env.HybridServer.UpdateHybridCloudEnvironmentCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "new-name", req.GetHybridCloudEnvironment().GetName())
}

func TestHybridUpdate_Configuration(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupHybridUpdateHandlers(env)

	_, _, err := testutil.Exec(t, env, "hybrid", "update", "env-abc",
		"--namespace", "new-ns",
		"--log-level", "debug",
	)
	require.NoError(t, err)

	req, ok := env.HybridServer.UpdateHybridCloudEnvironmentCalls.Last()
	require.True(t, ok)
	cfg := req.GetHybridCloudEnvironment().GetConfiguration()
	require.NotNil(t, cfg)
	assert.Equal(t, "new-ns", cfg.GetNamespace())
	assert.Equal(t, hybridv1.HybridCloudEnvironmentConfigurationLogLevel_HYBRID_CLOUD_ENVIRONMENT_CONFIGURATION_LOG_LEVEL_DEBUG, cfg.GetLogLevel())
}

func TestHybridUpdate_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "hybrid", "update")
	require.Error(t, err)
}

func TestHybridUpdate_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.HybridServer.GetHybridCloudEnvironmentCalls.Returns(&hybridv1.GetHybridCloudEnvironmentResponse{
		HybridCloudEnvironment: &hybridv1.HybridCloudEnvironment{Id: "env-abc", Name: "my-env"},
	}, nil)
	env.HybridServer.UpdateHybridCloudEnvironmentCalls.Returns(nil, fmt.Errorf("internal server error"))

	_, _, err := testutil.Exec(t, env, "hybrid", "update", "env-abc", "--name", "new-name")
	require.Error(t, err)
}
