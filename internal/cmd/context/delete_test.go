package context_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

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
