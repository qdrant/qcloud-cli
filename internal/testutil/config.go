package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/qdrant/qcloud-cli/internal/state/config"
)

// WriteConfigFile marshals content as JSON and writes it to dir/config.json.
// Returns the full path for use with --config or QDRANT_CLOUD_CONFIG.
func WriteConfigFile(t *testing.T, dir string, content map[string]any) string {
	t.Helper()
	data, err := json.Marshal(content)
	require.NoError(t, err)
	path := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(path, data, 0600))
	return path
}

// WriteYAMLConfigFile marshals content as YAML and writes it to dir/config.yaml.
// Returns the full path for use with --config or QDRANT_CLOUD_CONFIG.
func WriteYAMLConfigFile(t *testing.T, dir string, content map[string]any) string {
	t.Helper()
	data, err := yaml.Marshal(content)
	require.NoError(t, err)
	path := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(path, data, 0600))
	return path
}

// WriteContextConfigFile writes a context-format YAML config file to dir/config.yaml.
// Contexts are written as a YAML slice (- name: ...) sorted by name for deterministic output.
// Returns the full path for use with --config or QDRANT_CLOUD_CONFIG.
func WriteContextConfigFile(t *testing.T, dir, currentContext string, contexts map[string]map[string]string) string {
	t.Helper()

	fd := config.File{CurrentContext: currentContext}

	names := make([]string, 0, len(contexts))
	for name := range contexts {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		vals := contexts[name]
		fd.Contexts = append(fd.Contexts, config.ContextEntry{
			Name:      name,
			Endpoint:  vals["endpoint"],
			APIKey:    vals["api_key"],
			AccountID: vals["account_id"],
		})
	}

	data, err := yaml.Marshal(fd)
	require.NoError(t, err)
	path := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(path, data, 0600))
	return path
}

// FindContextEntry reads the YAML config file at path and returns a pointer to the
// ContextEntry whose Name matches name, or nil if not found.
func FindContextEntry(t *testing.T, path, name string) *config.ContextEntry {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var fd struct {
		Contexts []config.ContextEntry `yaml:"contexts"`
	}
	require.NoError(t, yaml.Unmarshal(data, &fd))
	for i := range fd.Contexts {
		if fd.Contexts[i].Name == name {
			return &fd.Contexts[i]
		}
	}
	return nil
}
