package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qdrant/qcloud-cli/internal/state/config"
	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestLoad_ExplicitPath(t *testing.T) {
	path := testutil.WriteConfigFile(t, t.TempDir(), map[string]any{
		"api_key":    "explicit-key",
		"account_id": "explicit-account",
	})

	c := config.New()
	require.NoError(t, c.Load(path))

	assert.Equal(t, "explicit-key", c.APIKey())
	assert.Equal(t, "explicit-account", c.AccountID())
}

func TestLoad_EnvVar(t *testing.T) {
	path := testutil.WriteConfigFile(t, t.TempDir(), map[string]any{
		"api_key":    "env-key",
		"account_id": "env-account",
	})

	t.Setenv("QDRANT_CLOUD_CONFIG", path)

	c := config.New()
	require.NoError(t, c.Load(""))

	assert.Equal(t, "env-key", c.APIKey())
	assert.Equal(t, "env-account", c.AccountID())
}

func TestLoad_MissingExplicitPathIsError(t *testing.T) {
	c := config.New()
	err := c.Load(filepath.Join(t.TempDir(), "nonexistent.json"))
	require.Error(t, err)
}

func TestLoad_MissingDefaultPathIsNotError(t *testing.T) {
	// When no explicit path is given and the default dir has no config file,
	// Load should succeed silently.
	t.Setenv("QDRANT_CLOUD_CONFIG", "")
	c := config.New()
	// Redirect home to an empty temp dir so the default path doesn't exist.
	t.Setenv("HOME", t.TempDir())
	require.NoError(t, c.Load(""))
}

func TestLoad_ExplicitYAMLPath(t *testing.T) {
	path := testutil.WriteYAMLConfigFile(t, t.TempDir(), map[string]any{
		"api_key":    "yaml-explicit-key",
		"account_id": "yaml-explicit-account",
	})

	c := config.New()
	require.NoError(t, c.Load(path))

	assert.Equal(t, "yaml-explicit-key", c.APIKey())
	assert.Equal(t, "yaml-explicit-account", c.AccountID())
}

func TestLoad_EnvVarYAML(t *testing.T) {
	path := testutil.WriteYAMLConfigFile(t, t.TempDir(), map[string]any{
		"api_key":    "yaml-env-key",
		"account_id": "yaml-env-account",
	})

	t.Setenv("QDRANT_CLOUD_CONFIG", path)

	c := config.New()
	require.NoError(t, c.Load(""))

	assert.Equal(t, "yaml-env-key", c.APIKey())
	assert.Equal(t, "yaml-env-account", c.AccountID())
}

func TestLoad_DefaultDirYAML(t *testing.T) {
	home := t.TempDir()
	configDir := filepath.Join(home, ".config", "qcloud")
	require.NoError(t, os.MkdirAll(configDir, 0700))
	testutil.WriteYAMLConfigFile(t, configDir, map[string]any{
		"api_key":    "yaml-default-key",
		"account_id": "yaml-default-account",
	})

	t.Setenv("QDRANT_CLOUD_CONFIG", "")
	t.Setenv("HOME", home)

	c := config.New()
	require.NoError(t, c.Load(""))

	assert.Equal(t, "yaml-default-key", c.APIKey())
	assert.Equal(t, "yaml-default-account", c.AccountID())
}
