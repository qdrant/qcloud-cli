package cluster_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestUnsuspend(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))

	env.Server.UnsuspendClusterCalls.Returns(&clusterv1.UnsuspendClusterResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "unsuspend", "cluster-123")
	require.NoError(t, err)

	req, ok := env.Server.UnsuspendClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "cluster-123", req.GetClusterId())
	assert.Contains(t, stdout, "cluster-123")
	assert.Contains(t, stdout, "unsuspending")
}

func TestUnsuspend_MissingArgs(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "cluster", "unsuspend")
	require.Error(t, err)
}
