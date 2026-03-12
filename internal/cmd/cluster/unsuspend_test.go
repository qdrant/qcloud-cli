package cluster_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestUnsuspend(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))
	t.Cleanup(env.Cleanup)

	var capturedReq *clusterv1.UnsuspendClusterRequest
	env.Server.UnsuspendClusterFunc = func(_ context.Context, req *clusterv1.UnsuspendClusterRequest) (*clusterv1.UnsuspendClusterResponse, error) {
		capturedReq = req
		return &clusterv1.UnsuspendClusterResponse{}, nil
	}

	stdout, _, err := testutil.Exec(t, env, "cluster", "unsuspend", "cluster-123")
	require.NoError(t, err)
	assert.Equal(t, "test-account-id", capturedReq.GetAccountId())
	assert.Equal(t, "cluster-123", capturedReq.GetClusterId())
	assert.Contains(t, stdout, "cluster-123")
	assert.Contains(t, stdout, "unsuspending")
}

func TestUnsuspend_MissingArgs(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	_, _, err := testutil.Exec(t, env, "cluster", "unsuspend")
	require.Error(t, err)
}
