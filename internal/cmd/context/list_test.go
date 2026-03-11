package context_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

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
