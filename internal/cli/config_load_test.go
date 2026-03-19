package cli_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

// TestConfigLoad_FlagSetsAccountID verifies that account_id from a config file
// loaded via --config reaches the gRPC request.
func TestConfigLoad_FlagSetsAccountID(t *testing.T) {
	env := testutil.NewBareTestEnv(t)

	cfgPath := testutil.WriteConfigFile(t, t.TempDir(), map[string]any{
		"account_id": "account-from-file",
	})

	env.ClusterServer.EXPECT().
		ListClusters(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, req *clusterv1.ListClustersRequest) (*clusterv1.ListClustersResponse, error) {
			assert.Equal(t, "account-from-file", req.GetAccountId())
			return &clusterv1.ListClustersResponse{}, nil
		})

	_, _, err := testutil.Exec(t, env, "--config", cfgPath, "cluster", "list")
	require.NoError(t, err)
}

// TestConfigLoad_EnvVarSetsAccountID verifies that QDRANT_CLOUD_CONFIG env var
// is respected when no --config flag is given.
func TestConfigLoad_EnvVarSetsAccountID(t *testing.T) {
	env := testutil.NewBareTestEnv(t)

	cfgPath := testutil.WriteConfigFile(t, t.TempDir(), map[string]any{
		"account_id": "account-from-envvar",
	})

	t.Setenv("QDRANT_CLOUD_CONFIG", cfgPath)

	env.ClusterServer.EXPECT().
		ListClusters(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, req *clusterv1.ListClustersRequest) (*clusterv1.ListClustersResponse, error) {
			assert.Equal(t, "account-from-envvar", req.GetAccountId())
			return &clusterv1.ListClustersResponse{}, nil
		})

	_, _, err := testutil.Exec(t, env, "cluster", "list")
	require.NoError(t, err)
}

// TestConfigLoad_FlagOverridesEnvVar verifies that --config flag takes
// precedence over QDRANT_CLOUD_CONFIG env var.
func TestConfigLoad_FlagOverridesEnvVar(t *testing.T) {
	env := testutil.NewBareTestEnv(t)

	dir := t.TempDir()
	flagCfg := testutil.WriteConfigFile(t, dir, map[string]any{
		"account_id": "account-from-flag",
	})
	envDir := t.TempDir()
	envCfg := testutil.WriteConfigFile(t, envDir, map[string]any{
		"account_id": "account-from-env",
	})

	t.Setenv("QDRANT_CLOUD_CONFIG", envCfg)

	env.ClusterServer.EXPECT().
		ListClusters(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, req *clusterv1.ListClustersRequest) (*clusterv1.ListClustersResponse, error) {
			assert.Equal(t, "account-from-flag", req.GetAccountId())
			return &clusterv1.ListClustersResponse{}, nil
		})

	_, _, err := testutil.Exec(t, env, "--config", flagCfg, "cluster", "list")
	require.NoError(t, err)
}

// TestConfigLoad_WithAccountIDTakesPrecedence verifies that WithAccountID (Set)
// takes precedence over a config file loaded via --config (Set > config file).
func TestConfigLoad_WithAccountIDTakesPrecedence(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("explicit-id"))

	cfgPath := testutil.WriteConfigFile(t, t.TempDir(), map[string]any{
		"account_id": "account-from-file",
	})

	env.ClusterServer.EXPECT().
		ListClusters(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, req *clusterv1.ListClustersRequest) (*clusterv1.ListClustersResponse, error) {
			assert.Equal(t, "explicit-id", req.GetAccountId())
			return &clusterv1.ListClustersResponse{}, nil
		})


	_, _, err := testutil.Exec(t, env, "--config", cfgPath, "cluster", "list")
	require.NoError(t, err)
}
