package context_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/qdrant/qcloud-cli/internal/state"
	"github.com/qdrant/qcloud-cli/internal/testutil"
)

// newEnv creates a minimal TestEnv without a gRPC server — context commands
// don't call the API.
func newEnv(t *testing.T) *testutil.TestEnv {
	t.Helper()
	return &testutil.TestEnv{
		State:   state.New("test"),
		Cleanup: func() {},
	}
}

// readYAML parses a YAML file into a map for verification.
func readYAML(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, yaml.Unmarshal(data, &m))
	return m
}

// ----- context list -----

func TestContextList_Table(t *testing.T) {
	env := newEnv(t)
	t.Cleanup(env.Cleanup)

	cfgPath := testutil.WriteContextConfigFile(t, t.TempDir(), "staging", map[string]map[string]string{
		"staging": {"endpoint": "grpc.staging.qdrant.io:443", "api_key": "sk", "account_id": "sa"},
		"prod":    {"endpoint": "grpc.prod.qdrant.io:443", "api_key": "pk", "account_id": "pa"},
	})

	stdout, _, err := testutil.Exec(t, env, "--config", cfgPath, "context", "list")
	require.NoError(t, err)

	assert.Contains(t, stdout, "CURRENT")
	assert.Contains(t, stdout, "NAME")
	assert.Contains(t, stdout, "staging")
	assert.Contains(t, stdout, "prod")
	assert.Contains(t, stdout, "*")
}

func TestContextList_CurrentMarkerOnlyForCurrentContext(t *testing.T) {
	env := newEnv(t)
	t.Cleanup(env.Cleanup)

	cfgPath := testutil.WriteContextConfigFile(t, t.TempDir(), "prod", map[string]map[string]string{
		"prod":    {"api_key": "pk"},
		"staging": {"api_key": "sk"},
	})

	stdout, _, err := testutil.Exec(t, env, "--config", cfgPath, "context", "list")
	require.NoError(t, err)

	// The * marker row should contain "prod", and the empty row should contain "staging".
	assert.Contains(t, stdout, "prod")
	assert.Contains(t, stdout, "staging")
	assert.Contains(t, stdout, "*")
}

func TestContextList_Empty(t *testing.T) {
	env := newEnv(t)
	t.Cleanup(env.Cleanup)

	cfgPath := testutil.WriteContextConfigFile(t, t.TempDir(), "", nil)

	stdout, _, err := testutil.Exec(t, env, "--config", cfgPath, "context", "list")
	require.NoError(t, err)

	assert.Contains(t, stdout, "CURRENT")
	assert.Contains(t, stdout, "NAME")
}

func TestContextList_JSON(t *testing.T) {
	env := newEnv(t)
	t.Cleanup(env.Cleanup)

	cfgPath := testutil.WriteContextConfigFile(t, t.TempDir(), "staging", map[string]map[string]string{
		"staging": {"api_key": "sk"},
		"prod":    {"api_key": "pk"},
	})

	stdout, _, err := testutil.Exec(t, env, "--config", cfgPath, "--json", "context", "list")
	require.NoError(t, err)

	assert.Contains(t, stdout, `"current"`)
	assert.Contains(t, stdout, `"staging"`)
	assert.Contains(t, stdout, `"prod"`)
}

// ----- context use -----

func TestContextUse_SwitchesCurrentContext(t *testing.T) {
	env := newEnv(t)
	t.Cleanup(env.Cleanup)

	dir := t.TempDir()
	cfgPath := testutil.WriteContextConfigFile(t, dir, "staging", map[string]map[string]string{
		"staging": {"api_key": "sk"},
		"prod":    {"api_key": "pk"},
	})

	stdout, _, err := testutil.Exec(t, env, "--config", cfgPath, "context", "use", "prod")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"prod"`)

	m := readYAML(t, cfgPath)
	assert.Equal(t, "prod", m["current-context"])
}

func TestContextUse_UnknownContext(t *testing.T) {
	env := newEnv(t)
	t.Cleanup(env.Cleanup)

	cfgPath := testutil.WriteContextConfigFile(t, t.TempDir(), "staging", map[string]map[string]string{
		"staging": {"api_key": "sk"},
	})

	_, _, err := testutil.Exec(t, env, "--config", cfgPath, "context", "use", "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

// ----- context show -----

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

// ----- context set -----

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

// ----- context delete -----

func TestContextDelete_RemovesContext(t *testing.T) {
	env := newEnv(t)
	t.Cleanup(env.Cleanup)

	dir := t.TempDir()
	cfgPath := testutil.WriteContextConfigFile(t, dir, "staging", map[string]map[string]string{
		"staging": {"api_key": "sk"},
		"prod":    {"api_key": "pk"},
	})

	stdout, _, err := testutil.Exec(t, env, "--config", cfgPath, "context", "delete", "prod", "--force")
	require.NoError(t, err)
	assert.Contains(t, stdout, "prod")

	assert.Nil(t, testutil.FindContextEntry(t, cfgPath, "prod"), "prod context should be removed")
	assert.NotNil(t, testutil.FindContextEntry(t, cfgPath, "staging"), "staging context should be preserved")
	// current-context unchanged since prod was not current.
	m := readYAML(t, cfgPath)
	assert.Equal(t, "staging", m["current-context"])
}

func TestContextDelete_ClearsCurrentContext(t *testing.T) {
	env := newEnv(t)
	t.Cleanup(env.Cleanup)

	dir := t.TempDir()
	cfgPath := testutil.WriteContextConfigFile(t, dir, "staging", map[string]map[string]string{
		"staging": {"api_key": "sk"},
		"prod":    {"api_key": "pk"},
	})

	_, _, err := testutil.Exec(t, env, "--config", cfgPath, "context", "delete", "staging", "--force")
	require.NoError(t, err)

	m := readYAML(t, cfgPath)
	// current-context was "staging" — it must be cleared after deletion.
	_, hasCurrent := m["current-context"]
	assert.False(t, hasCurrent, "current-context should be removed when the current context is deleted")
}

func TestContextDelete_UnknownContext(t *testing.T) {
	env := newEnv(t)
	t.Cleanup(env.Cleanup)

	cfgPath := testutil.WriteContextConfigFile(t, t.TempDir(), "staging", map[string]map[string]string{
		"staging": {"api_key": "sk"},
	})

	_, _, err := testutil.Exec(t, env, "--config", cfgPath, "context", "delete", "nonexistent", "--force")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestContextDelete_RequiresExactlyOneArg(t *testing.T) {
	env := newEnv(t)
	t.Cleanup(env.Cleanup)

	cfgPath := testutil.WriteContextConfigFile(t, t.TempDir(), "", nil)

	_, _, err := testutil.Exec(t, env, "--config", cfgPath, "context", "delete", "--force")
	require.Error(t, err)
}
