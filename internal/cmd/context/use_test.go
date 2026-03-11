package context_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

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
	assert.Equal(t, "prod", m["current_context"])
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
