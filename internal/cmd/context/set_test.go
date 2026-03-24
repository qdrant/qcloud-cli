package context_test

import (
	"path"
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
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	t.Run("api-key", func(t *testing.T) {
		env := newEnv(t)
		t.Cleanup(env.Cleanup)

		_, _, err := testutil.Exec(t, env, "context", "set", "test",
			"--endpoint", "grpc.test-cloud.qdrant.io:443",
			"--account-id", "test",
		)
		require.Error(t, err)
	})

	t.Run("account-id", func(t *testing.T) {
		env := newEnv(t)
		t.Cleanup(env.Cleanup)

		_, _, err := testutil.Exec(t, env, "context", "set", "test",
			"--endpoint", "grpc.test-cloud.qdrant.io:443",
			"--api-key", "thekey",
		)
		require.Error(t, err)
	})
}

func TestContextSet_InheritsValuesFromEnvVarsOrDefaultsIfMissing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	t.Run("api-key", func(t *testing.T) {
		env := newEnv(t)
		t.Cleanup(env.Cleanup)

		t.Setenv("QDRANT_CLOUD_API_KEY", "test-key")

		_, _, err := testutil.Exec(t, env, "context", "set", "test",
			"--endpoint", "grpc.test-cloud.qdrant.io:443",
			"--account-id", "test",
		)
		require.NoError(t, err)
	})

	t.Run("account-id", func(t *testing.T) {
		env := newEnv(t)
		t.Cleanup(env.Cleanup)

		t.Setenv("QDRANT_CLOUD_ACCOUNT_ID", "3ef9543c-6cea-4ef3-b558-787b688dd03f")

		_, _, err := testutil.Exec(t, env, "context", "set", "test",
			"--endpoint", "grpc.test-cloud.qdrant.io:443",
			"--api-key", "thekey",
		)
		require.NoError(t, err)
	})

	t.Run("endpoint", func(t *testing.T) {
		env := newEnv(t)
		t.Cleanup(env.Cleanup)

		t.Setenv("QDRANT_CLOUD_ENDPOINT", "grpc.test-cloud.qdrant.io:443")
		_, _, err := testutil.Exec(t, env, "context", "set", "test",
			"--api-key", "thekey",
			"--account-id", "780c7589-f3e8-4567-808f-60a54d43ae10",
		)
		require.NoError(t, err)

		cfgPath := path.Join(dir, ".config", "qcloud", "config.yaml")
		ctxE := testutil.FindContextEntry(t, cfgPath, "test")
		require.NotNil(t, ctxE)
		// Only api_key changed; other fields preserved.
		assert.Equal(t, "thekey", ctxE.APIKey)
		assert.Equal(t, "grpc.test-cloud.qdrant.io:443", ctxE.Endpoint)
		assert.Equal(t, "780c7589-f3e8-4567-808f-60a54d43ae10", ctxE.AccountID)
	})

	t.Run("api-key-command", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("HOME", dir)

		env := newEnv(t)
		t.Cleanup(env.Cleanup)

		_, _, err := testutil.Exec(t, env, "context", "set", "test",
			"--api-key-command", "echo thekey",
			"--account-id", "780c7589-f3e8-4567-808f-60a54d43ae10",
		)
		require.NoError(t, err)

		cfgPath := path.Join(dir, ".config", "qcloud", "config.yaml")
		ctxE := testutil.FindContextEntry(t, cfgPath, "test")
		require.NotNil(t, ctxE)
		assert.Equal(t, "echo thekey", ctxE.APIKeyCommand)
		assert.Empty(t, ctxE.APIKey)
	})

	t.Run("endpoint is defaulted to hardcoded value without env var", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("HOME", dir)

		env := newEnv(t)
		t.Cleanup(env.Cleanup)

		_, _, err := testutil.Exec(t, env, "context", "set", "test",
			"--api-key", "thekey",
			"--account-id", "780c7589-f3e8-4567-808f-60a54d43ae10",
		)
		require.NoError(t, err)

		cfgPath := path.Join(dir, ".config", "qcloud", "config.yaml")
		ctxE := testutil.FindContextEntry(t, cfgPath, "test")
		require.NotNil(t, ctxE)
		// Only api_key changed; other fields preserved.
		assert.Equal(t, "thekey", ctxE.APIKey)
		assert.Equal(t, "grpc.cloud.qdrant.io:443", ctxE.Endpoint)
		assert.Equal(t, "780c7589-f3e8-4567-808f-60a54d43ae10", ctxE.AccountID)
	})
}

func TestContextSet_APIKeyHelper(t *testing.T) {
	tests := []struct {
		helper  string
		ref     string
		wantCmd string
	}{
		{"1password", "op://vault/qdrant/key", "op read op://vault/qdrant/key"},
		{"vault", "secret/qdrant", "vault kv get -field=api_key secret/qdrant"},
		{"pass", "qdrant/api-key", "pass show qdrant/api-key"},
		{"keychain", "qcloud-prod", "security find-generic-password -s qcloud-prod -w"},
		{"1password", "ls -la /etc", "op read 'ls -la /etc'"},
	}

	for _, tt := range tests {
		t.Run(tt.helper, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("HOME", dir)

			env := newEnv(t)
			t.Cleanup(env.Cleanup)

			_, _, err := testutil.Exec(t, env, "context", "set", "test",
				"--api-key-helper", tt.helper,
				"--api-key-ref", tt.ref,
				"--account-id", "acct-1",
			)
			require.NoError(t, err)

			cfgPath := path.Join(dir, ".config", "qcloud", "config.yaml")
			ctxE := testutil.FindContextEntry(t, cfgPath, "test")
			require.NotNil(t, ctxE)
			assert.Equal(t, tt.wantCmd, ctxE.APIKeyCommand)
			assert.Empty(t, ctxE.APIKey)
		})
	}
}

func TestContextSet_APIKeyHelper_UnknownHelper(t *testing.T) {
	env := newEnv(t)
	t.Cleanup(env.Cleanup)

	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_, _, err := testutil.Exec(t, env, "context", "set", "test",
		"--api-key-helper", "unknown",
		"--api-key-ref", "some-ref",
		"--account-id", "acct-1",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown api-key-helper")
}

func TestContextSet_APIKeyAndCommandMutuallyExclusive(t *testing.T) {
	env := newEnv(t)
	t.Cleanup(env.Cleanup)

	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_, _, err := testutil.Exec(t, env, "context", "set", "test",
		"--api-key", "thekey",
		"--api-key-command", "echo thekey",
		"--account-id", "acct-1",
	)
	require.Error(t, err)
}

func TestContextSet_APIKeyCommandClearsPlainKey(t *testing.T) {
	env := newEnv(t)
	t.Cleanup(env.Cleanup)

	dir := t.TempDir()
	cfgPath := testutil.WriteContextConfigFile(t, dir, "prod", map[string]map[string]string{
		"prod": {
			"endpoint":   "grpc.cloud.qdrant.io:443",
			"api_key":    "old-plain-key",
			"account_id": "acct-1",
		},
	})

	_, _, err := testutil.Exec(t, env, "--config", cfgPath, "context", "set", "prod",
		"--api-key-command", "echo new-command-key",
	)
	require.NoError(t, err)

	ctxE := testutil.FindContextEntry(t, cfgPath, "prod")
	require.NotNil(t, ctxE)
	assert.Equal(t, "echo new-command-key", ctxE.APIKeyCommand)
	assert.Empty(t, ctxE.APIKey)
}
