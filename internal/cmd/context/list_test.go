package context_test

import (
	"encoding/json"
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
	assert.Contains(t, stdout, "CREDENTIAL")
	assert.Contains(t, stdout, "staging")
	assert.Contains(t, stdout, "prod")
	assert.Contains(t, stdout, "*")
}

func TestContextList_CredentialColumn(t *testing.T) {
	env := newEnv(t)
	t.Cleanup(env.Cleanup)

	cfgPath := testutil.WriteContextConfigFile(t, t.TempDir(), "prod", map[string]map[string]string{
		"prod":           {"endpoint": "grpc.cloud.qdrant.io:443", "api_key": "pk", "account_id": "pa"},
		"dev-my-account": {"endpoint": "grpc.development-cloud.qdrant.io:443", "account_id": "acc-1", "auth": "oauth"},
	})

	stdout, _, err := testutil.Exec(t, env, "--config", cfgPath, "context", "list")
	require.NoError(t, err)

	assert.Regexp(t, `CURRENT\s+NAME\s+CREDENTIAL`, stdout)
	assert.Regexp(t, `dev-my-account\s+oauth`, stdout)
	assert.Regexp(t, `\*\s+prod\s+api key`, stdout)
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

func TestContextList_JSONIncludesTheCredentialOfEachContext(t *testing.T) {
	env := newEnv(t)
	t.Cleanup(env.Cleanup)

	cfgPath := testutil.WriteContextConfigFile(t, t.TempDir(), "prod", map[string]map[string]string{
		"prod":           {"api_key": "pk"},
		"dev-my-account": {"endpoint": "grpc.development-cloud.qdrant.io:443", "account_id": "acc-1", "auth": "oauth"},
	})

	stdout, _, err := testutil.Exec(t, env, "--config", cfgPath, "--json", "context", "list")
	require.NoError(t, err)

	var got struct {
		Contexts []string          `json:"contexts"`
		Current  string            `json:"current"`
		Auth     map[string]string `json:"auth"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &got))
	assert.Equal(t, []string{"dev-my-account", "prod"}, got.Contexts)
	assert.Equal(t, "prod", got.Current)
	assert.Equal(t, map[string]string{"dev-my-account": "oauth", "prod": "api key"}, got.Auth)
}
