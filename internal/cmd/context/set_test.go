package context_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestContextSet_CreatesNewContext(t *testing.T) {
	env := newEnv(t)
	t.Cleanup(env.Cleanup)

	dir := t.TempDir()
	cfgPath := testutil.WriteContextConfigFile(t, dir, "staging", map[string]map[string]string{
		"staging": {"api_key": "sk"},
	})

	stdout, _, err := testutil.Exec(t, env, "--config", cfgPath, "context", "set", "prod",
		"--endpoint", "grpc.prod.qdrant.io:443",
		"--api-key", "prod-key",
		"--account-id", "prod-acct",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "prod")

	prod := testutil.FindContextEntry(t, cfgPath, "prod")
	require.NotNil(t, prod)
	assert.Equal(t, "grpc.prod.qdrant.io:443", prod.Endpoint)
	assert.Equal(t, "prod-key", prod.APIKey)
	assert.Equal(t, "prod-acct", prod.AccountID)
}

func TestContextSet_PartialUpdate(t *testing.T) {
	env := newEnv(t)
	t.Cleanup(env.Cleanup)

	dir := t.TempDir()
	cfgPath := testutil.WriteContextConfigFile(t, dir, "staging", map[string]map[string]string{
		"staging": {
			"endpoint":   "grpc.staging.qdrant.io:443",
			"api_key":    "old-key",
			"account_id": "old-acct",
		},
	})

	_, _, err := testutil.Exec(t, env, "--config", cfgPath, "context", "set", "staging",
		"--api-key", "new-key",
	)
	require.NoError(t, err)

	staging := testutil.FindContextEntry(t, cfgPath, "staging")
	require.NotNil(t, staging)
	// Only api_key changed; other fields preserved.
	assert.Equal(t, "new-key", staging.APIKey)
	assert.Equal(t, "grpc.staging.qdrant.io:443", staging.Endpoint)
	assert.Equal(t, "old-acct", staging.AccountID)
}

func TestContextSet_AutoActivatesWhenNoCurrentContext(t *testing.T) {
	env := newEnv(t)
	t.Cleanup(env.Cleanup)

	dir := t.TempDir()
	cfgPath := testutil.WriteContextConfigFile(t, dir, "", nil)

	stdout, _, err := testutil.Exec(t, env, "--config", cfgPath, "context", "set", "first",
		"--endpoint", "grpc.example.com:443",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, `"first"`)

	m := readYAML(t, cfgPath)
	assert.Equal(t, "first", m["current-context"])
}

func TestContextSet_DoesNotActivateWhenCurrentContextExists(t *testing.T) {
	env := newEnv(t)
	t.Cleanup(env.Cleanup)

	dir := t.TempDir()
	cfgPath := testutil.WriteContextConfigFile(t, dir, "existing", map[string]map[string]string{
		"existing": {"api_key": "ek"},
	})

	_, _, err := testutil.Exec(t, env, "--config", cfgPath, "context", "set", "new",
		"--endpoint", "grpc.new.com:443",
	)
	require.NoError(t, err)

	m := readYAML(t, cfgPath)
	assert.Equal(t, "existing", m["current-context"])
}
