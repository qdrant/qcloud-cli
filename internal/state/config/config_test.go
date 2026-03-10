package config_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qdrant/qcloud-cli/internal/state/config"
	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestLoad_ExplicitPath(t *testing.T) {
	path := testutil.WriteConfigFile(t, t.TempDir(), map[string]any{
		"management_key": "explicit-key",
		"account_id":     "explicit-account",
	})

	c := config.New()
	require.NoError(t, c.Load(path))

	assert.Equal(t, "explicit-key", c.APIKey())
	assert.Equal(t, "explicit-account", c.AccountID())
}

func TestLoad_EnvVar(t *testing.T) {
	path := testutil.WriteConfigFile(t, t.TempDir(), map[string]any{
		"management_key": "env-key",
		"account_id":     "env-account",
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

