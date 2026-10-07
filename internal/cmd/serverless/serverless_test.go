package serverless_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestServerless_Help(t *testing.T) {
	env := testutil.NewTestEnv(t)

	stdout, _, err := testutil.Exec(t, env, "serverless", "--help")
	require.NoError(t, err)
	assert.Contains(t, stdout, "space")
	assert.Contains(t, stdout, "cloud-region")
}
