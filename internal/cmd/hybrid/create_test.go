package hybrid_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestHybridCreate_PrintsContactMessage(t *testing.T) {
	env := testutil.NewTestEnv(t)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "create")
	require.NoError(t, err)

	assert.Contains(t, stdout, "https://qdrant.tech/contact-us/")
}

func TestHybridCreate_RejectsFlags(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "hybrid", "create", "--name", "my-env")
	require.Error(t, err)
}
