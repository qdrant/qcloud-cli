package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

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

// ----- context format tests -----

func TestLoad_ContextFormat_InjectsActiveContextValues(t *testing.T) {
	cfgPath := testutil.WriteContextConfigFile(t, t.TempDir(), "staging", map[string]map[string]string{
		"staging": {
			"endpoint":   "grpc.staging.qdrant.io:443",
			"api_key":    "staging-key",
			"account_id": "staging-acct",
		},
		"prod": {
			"endpoint":   "grpc.prod.qdrant.io:443",
			"api_key":    "prod-key",
			"account_id": "prod-acct",
		},
	})

	c := config.New()
	require.NoError(t, c.Load(cfgPath))

	// Active context is "staging" — its values should be resolved.
	assert.Equal(t, "staging-key", c.APIKey())
	assert.Equal(t, "staging-acct", c.AccountID())
	assert.Equal(t, "grpc.staging.qdrant.io:443", c.Endpoint())
	assert.Equal(t, "staging", c.CurrentContext())
}

func TestLoad_ContextFormat_NoCurrentContext(t *testing.T) {
	cfgPath := testutil.WriteContextConfigFile(t, t.TempDir(), "", map[string]map[string]string{
		"prod": {"api_key": "pk", "account_id": "pa"},
	})

	c := config.New()
	require.NoError(t, c.Load(cfgPath))

	// No active context — top-level values fall back to defaults/empty.
	assert.Empty(t, c.APIKey())
	assert.Empty(t, c.CurrentContext())
}

func TestContextNames_SortedList(t *testing.T) {
	cfgPath := testutil.WriteContextConfigFile(t, t.TempDir(), "staging", map[string]map[string]string{
		"staging": {"api_key": "sk"},
		"prod":    {"api_key": "pk"},
		"dev":     {"api_key": "dk"},
	})

	c := config.New()
	require.NoError(t, c.Load(cfgPath))

	assert.Equal(t, []string{"dev", "prod", "staging"}, c.ContextNames())
}

func TestContextNames_Empty(t *testing.T) {
	cfgPath := testutil.WriteContextConfigFile(t, t.TempDir(), "", nil)

	c := config.New()
	require.NoError(t, c.Load(cfgPath))

	assert.Empty(t, c.ContextNames())
}

func TestGetContext_Found(t *testing.T) {
	cfgPath := testutil.WriteContextConfigFile(t, t.TempDir(), "staging", map[string]map[string]string{
		"staging": {"api_key": "sk", "account_id": "sa"},
	})

	c := config.New()
	require.NoError(t, c.Load(cfgPath))

	entry, ok := c.GetContext("staging")
	require.True(t, ok)
	assert.Equal(t, "sk", entry.APIKey)
	assert.Equal(t, "sa", entry.AccountID)
}

func TestGetContext_NotFound(t *testing.T) {
	cfgPath := testutil.WriteContextConfigFile(t, t.TempDir(), "", nil)

	c := config.New()
	require.NoError(t, c.Load(cfgPath))

	_, ok := c.GetContext("missing")
	assert.False(t, ok)
}

func TestUpsertContext_NewContext(t *testing.T) {
	dir := t.TempDir()
	cfgPath := testutil.WriteContextConfigFile(t, dir, "staging", map[string]map[string]string{
		"staging": {"api_key": "sk"},
	})

	c := config.New()
	require.NoError(t, c.Load(cfgPath))

	c.UpsertContext(config.ContextEntry{Name: "prod", APIKey: "pk", AccountID: "pa"})
	require.NoError(t, c.WriteToFile())

	prod := testutil.FindContextEntry(t, cfgPath, "prod")
	require.NotNil(t, prod, "prod context not found in saved file")
	assert.Equal(t, "pk", prod.APIKey)
	assert.Equal(t, "pa", prod.AccountID)
	// Original context preserved.
	assert.NotNil(t, testutil.FindContextEntry(t, cfgPath, "staging"), "staging context should be preserved")
}

func TestSetCurrentContext_UpdatesFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := testutil.WriteContextConfigFile(t, dir, "staging", map[string]map[string]string{
		"staging": {"api_key": "sk"},
		"prod":    {"api_key": "pk"},
	})

	c := config.New()
	require.NoError(t, c.Load(cfgPath))

	c.SetCurrentContext("prod")
	require.NoError(t, c.WriteToFile())

	data, err := os.ReadFile(cfgPath)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, yaml.Unmarshal(data, &m))
	assert.Equal(t, "prod", m["current_context"])
}

func TestDeleteContext_RemovesContextAndClearsCurrent(t *testing.T) {
	dir := t.TempDir()
	cfgPath := testutil.WriteContextConfigFile(t, dir, "staging", map[string]map[string]string{
		"staging": {"api_key": "sk"},
		"prod":    {"api_key": "pk"},
	})

	c := config.New()
	require.NoError(t, c.Load(cfgPath))

	c.DeleteContext("staging")
	require.NoError(t, c.WriteToFile())

	data, err := os.ReadFile(cfgPath)
	require.NoError(t, err)

	assert.Nil(t, testutil.FindContextEntry(t, cfgPath, "staging"), "staging context should be removed")
	assert.NotNil(t, testutil.FindContextEntry(t, cfgPath, "prod"), "prod context should be preserved")

	var m map[string]any
	require.NoError(t, yaml.Unmarshal(data, &m))
	_, hasCurrent := m["current_context"]
	assert.False(t, hasCurrent, "current_context should be removed when the current context is deleted")
}

func TestWriteToFile_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := testutil.WriteContextConfigFile(t, dir, "staging", map[string]map[string]string{
		"staging": {"api_key": "sk"},
	})

	c := config.New()
	require.NoError(t, c.Load(cfgPath))

	c.SetCurrentContext("staging")
	require.NoError(t, c.WriteToFile())

	// File must still be valid YAML after save.
	data, err := os.ReadFile(cfgPath)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, yaml.Unmarshal(data, &m))
	assert.Equal(t, "staging", m["current_context"])
}

func TestWriteToFile_CreatesDefaultPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("QDRANT_CLOUD_CONFIG", "")

	c := config.New()
	require.NoError(t, c.Load(""))

	c.UpsertContext(config.ContextEntry{Name: "dev", APIKey: "dk"})
	require.NoError(t, c.WriteToFile())

	defaultPath := filepath.Join(home, ".config", "qcloud", "config.yaml")
	assert.FileExists(t, defaultPath)

	assert.NotNil(t, testutil.FindContextEntry(t, defaultPath, "dev"), "dev context not found in saved file")
}

// TestConfigFileTagsAreSnakeCase asserts that every field tag (mapstructure, yaml, json)
// on the exported config structs uses snake_case names.
// This prevents accidental camelCase or other styles from being introduced.
func TestConfigFileTagsAreSnakeCase(t *testing.T) {
	snakeCase := regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*$`)
	tagNames := []string{"mapstructure", "yaml", "json"}

	structs := []any{config.ContextEntry{}, config.File{}}
	for _, s := range structs {
		typ := reflect.TypeOf(s)
		for field := range typ.Fields() {
			for _, tagName := range tagNames {
				val := field.Tag.Get(tagName)
				if val == "" || val == "-" {
					continue
				}
				name, _, _ := strings.Cut(val, ",")
				assert.Truef(t, snakeCase.MatchString(name),
					"%s.%s: tag %q value %q is not snake_case", typ.Name(), field.Name, tagName, name)
			}
		}
	}
}
