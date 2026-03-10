package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
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
