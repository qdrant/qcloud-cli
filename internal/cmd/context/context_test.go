package context_test

import (
	"os"
	"testing"

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
