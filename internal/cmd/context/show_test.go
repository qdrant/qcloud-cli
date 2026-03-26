package context_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestContextShow_ResolvedValues(t *testing.T) {
	env := newEnv(t)
	t.Cleanup(env.Cleanup)

	cfgPath := testutil.WriteContextConfigFile(t, t.TempDir(), "staging", map[string]map[string]string{
		"staging": {
			"endpoint":   "grpc.staging.qdrant.io:443",
			"account_id": "stage-acct",
			"api_key":    "sk",
		},
	})

	stdout, _, err := testutil.Exec(t, env, "--config", cfgPath, "context", "show")
	require.NoError(t, err)

	assert.Contains(t, stdout, "staging")
	assert.Contains(t, stdout, "grpc.staging.qdrant.io:443")
	assert.Contains(t, stdout, "stage-acct")
	// API key must NOT appear in output.
	assert.NotContains(t, stdout, "sk")
}

func TestContextShow_BackendURL(t *testing.T) {
	env := newEnv(t)
	t.Cleanup(env.Cleanup)

	cfgPath := testutil.WriteContextConfigFile(t, t.TempDir(), "staging", map[string]map[string]string{
		"staging": {
			"endpoint":    "grpc.staging.qdrant.io:443",
			"backend_url": "https://staging.cloud.qdrant.io",
			"account_id":  "stage-acct",
			"api_key":     "sk",
		},
	})

	stdout, _, err := testutil.Exec(t, env, "--config", cfgPath, "context", "show")
	require.NoError(t, err)

	assert.Contains(t, stdout, "https://staging.cloud.qdrant.io")
}

func TestContextShow_FlagOverridesCurrentContext(t *testing.T) {
	env := newEnv(t)
	t.Cleanup(env.Cleanup)

	cfgPath := testutil.WriteContextConfigFile(t, t.TempDir(), "staging", map[string]map[string]string{
		"staging": {"endpoint": "grpc.staging.qdrant.io:443", "account_id": "stage-acct", "api_key": "sk"},
		"prod":    {"endpoint": "grpc.prod.qdrant.io:443", "account_id": "prod-acct", "api_key": "pk"},
	})

	stdout, _, err := testutil.Exec(t, env, "--config", cfgPath, "--context", "prod", "context", "show")
	require.NoError(t, err)

	assert.Contains(t, stdout, "prod")
	assert.Contains(t, stdout, "grpc.prod.qdrant.io:443")
	assert.Contains(t, stdout, "prod-acct")
}

func TestContextShow_JSON(t *testing.T) {
	env := newEnv(t)
	t.Cleanup(env.Cleanup)

	cfgPath := testutil.WriteContextConfigFile(t, t.TempDir(), "staging", map[string]map[string]string{
		"staging": {"endpoint": "grpc.staging.qdrant.io:443", "account_id": "sa", "api_key": "sk"},
	})

	stdout, _, err := testutil.Exec(t, env, "--config", cfgPath, "--json", "context", "show")
	require.NoError(t, err)

	assert.Contains(t, stdout, `"context"`)
	assert.Contains(t, stdout, `"endpoint"`)
	assert.Contains(t, stdout, `"account_id"`)
	// API key must NOT appear in JSON output either.
	assert.NotContains(t, stdout, `"api_key"`)
}
