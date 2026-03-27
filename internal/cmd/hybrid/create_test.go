package hybrid_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	hybridv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/hybrid/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestHybridCreate_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.HybridServer.CreateHybridCloudEnvironmentCalls.Returns(&hybridv1.CreateHybridCloudEnvironmentResponse{
		HybridCloudEnvironment: &hybridv1.HybridCloudEnvironment{
			Id:   "env-new",
			Name: "my-env",
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "create", "--name", "my-env")
	require.NoError(t, err)

	assert.Contains(t, stdout, "env-new")
	assert.Contains(t, stdout, "my-env")

	req, ok := env.HybridServer.CreateHybridCloudEnvironmentCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetHybridCloudEnvironment().GetAccountId())
	assert.Equal(t, "my-env", req.GetHybridCloudEnvironment().GetName())
}

func TestHybridCreate_MissingName(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "hybrid", "create")
	require.Error(t, err)
}

func TestHybridCreate_WithOptionalFlags(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.HybridServer.CreateHybridCloudEnvironmentCalls.Returns(&hybridv1.CreateHybridCloudEnvironmentResponse{
		HybridCloudEnvironment: &hybridv1.HybridCloudEnvironment{
			Id:   "env-opts",
			Name: "opts-env",
		},
	}, nil)

	_, _, err := testutil.Exec(t, env, "hybrid", "create",
		"--name", "opts-env",
		"--namespace", "qdrant-ns",
		"--database-storage-class", "fast-ssd",
		"--snapshot-storage-class", "standard",
		"--log-level", "info",
	)
	require.NoError(t, err)

	req, ok := env.HybridServer.CreateHybridCloudEnvironmentCalls.Last()
	require.True(t, ok)
	cfg := req.GetHybridCloudEnvironment().GetConfiguration()
	require.NotNil(t, cfg)
	assert.Equal(t, "qdrant-ns", cfg.GetNamespace())
	assert.Equal(t, "fast-ssd", cfg.GetDatabaseStorageClass())
	assert.Equal(t, "standard", cfg.GetSnapshotStorageClass())
	assert.Equal(t, hybridv1.HybridCloudEnvironmentConfigurationLogLevel_HYBRID_CLOUD_ENVIRONMENT_CONFIGURATION_LOG_LEVEL_INFO, cfg.GetLogLevel())
}

func TestHybridCreate_InvalidLogLevel(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "hybrid", "create", "--name", "my-env", "--log-level", "invalid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log level")
}

func TestHybridCreate_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.HybridServer.CreateHybridCloudEnvironmentCalls.Returns(nil, fmt.Errorf("quota exceeded"))

	_, _, err := testutil.Exec(t, env, "hybrid", "create", "--name", "my-env")
	require.Error(t, err)
}

func TestHybridCreate_PermissionDenied(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.HybridServer.CreateHybridCloudEnvironmentCalls.Returns(nil,
		status.Error(codes.PermissionDenied, "Account is not entitled to do the requested action"))

	_, _, err := testutil.Exec(t, env, "hybrid", "create", "--name", "my-env")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not have access")
	assert.Contains(t, err.Error(), "https://qdrant.tech/contact-us/")
}
