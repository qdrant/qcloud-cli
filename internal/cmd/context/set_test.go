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

	stdout, _, err := testutil.Exec(t, env, "--config", cfgPath, "context", "set", "test",
		"--endpoint", "grpc.test.qdrant.io:443",
		"--api-key", "test-key",
		"--account-id", "test-acct",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "test")

	testCtx := testutil.FindContextEntry(t, cfgPath, "test")
	require.NotNil(t, testCtx)
	assert.Equal(t, "grpc.test.qdrant.io:443", testCtx.Endpoint)
	assert.Equal(t, "test-key", testCtx.APIKey)
	assert.Equal(t, "test-acct", testCtx.AccountID)
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
		"--account-id", "522a0c47-b7bf-45ef-892d-6551bc475e48",
		"--api-key", "test-key",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, `"first"`)

	m := readYAML(t, cfgPath)
	assert.Equal(t, "first", m["current_context"])
}

func TestContextSet_DoesNotActivateWhenCurrentContextExists(t *testing.T) {
	env := newEnv(t)

	dir := t.TempDir()
	cfgPath := testutil.WriteContextConfigFile(t, dir, "existing", map[string]map[string]string{
		"existing": {"api_key": "ek"},
	})

	_, _, err := testutil.Exec(t, env, "--config", cfgPath, "context", "set", "new",
		"--endpoint", "grpc.test-qdrant.cloud.io:443",
		"--account-id", "522a0c47-b7bf-45ef-892d-6551bc475e48",
		"--api-key", "test-key",
	)
	require.NoError(t, err)

	m := readYAML(t, cfgPath)
	assert.Equal(t, "existing", m["current_context"])
}

func TestContextSet_FailsWhenEntriesAreMissing(t *testing.T) {
	t.Run("api-key", func(t *testing.T) {
		env := newEnv(t)
		t.Cleanup(env.Cleanup)

		_, _, err := testutil.Exec(t, env, "context", "set", "test",
			"--endpoint", "grpc.test-cloud.qdrant.io:443",
			"--account-id", "test",
		)
		require.NoError(t, err)
	})
	
	t.Run("account-id", func(t *testing.T) {
		env := newEnv(t)
		t.Cleanup(env.Cleanup)

		_, _, err := testutil.Exec(t, env, "context", "set", "test",
			"--endpoint", "grpc.test-cloud.qdrant.io:443",
			"--api-key", "thekey",
		)
		require.NoError(t, err)
	})

	t.Run("endpoint", func(t *testing.T) {
		env := newEnv(t)
		t.Cleanup(env.Cleanup)

		_, _, err := testutil.Exec(t, env, "context", "set", "test",
			"--api-key", "thekey",
			"--account-id", "780c7589-f3e8-4567-808f-60a54d43ae10",
		)
		require.NoError(t, err)
	})
}
