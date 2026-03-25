package hybrid_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hybridv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/hybrid/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestHybridBootstrap_OutputsCommands(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.HybridServer.GenerateBootstrapCommandsCalls.Returns(&hybridv1.GenerateBootstrapCommandsResponse{
		Commands:         []string{"kubectl apply -f qdrant.yaml", "helm install qdrant qdrant/qdrant"},
		AccessKey:        "super-secret-key",
		RegistryUsername: "reg-user",
		RegistryPassword: "reg-pass",
	}, nil)

	stdout, stderr, err := testutil.Exec(t, env, "hybrid", "bootstrap", "env-abc")
	require.NoError(t, err)

	assert.Contains(t, stdout, "kubectl apply -f qdrant.yaml")
	assert.Contains(t, stdout, "helm install qdrant qdrant/qdrant")

	assert.Contains(t, stderr, "super-secret-key")
	assert.Contains(t, stderr, "reg-user")
	assert.Contains(t, stderr, "reg-pass")
	assert.Contains(t, stderr, "WARNING")

	req, ok := env.HybridServer.GenerateBootstrapCommandsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "env-abc", req.GetHybridCloudEnvironmentId())
}

func TestHybridBootstrap_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "hybrid", "bootstrap")
	require.Error(t, err)
}

func TestHybridBootstrap_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.HybridServer.GenerateBootstrapCommandsCalls.Returns(nil, fmt.Errorf("server error"))

	_, _, err := testutil.Exec(t, env, "hybrid", "bootstrap", "env-abc")
	require.Error(t, err)
}
