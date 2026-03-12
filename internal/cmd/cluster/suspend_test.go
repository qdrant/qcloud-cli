package cluster_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestSuspend_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))
	t.Cleanup(env.Cleanup)

	var capturedReq *clusterv1.SuspendClusterRequest
	env.Server.SuspendClusterFunc = func(_ context.Context, req *clusterv1.SuspendClusterRequest) (*clusterv1.SuspendClusterResponse, error) {
		capturedReq = req
		return &clusterv1.SuspendClusterResponse{}, nil
	}

	stdout, _, err := testutil.Exec(t, env, "cluster", "suspend", "cluster-123", "--force")
	require.NoError(t, err)
	assert.Equal(t, "test-account-id", capturedReq.GetAccountId())
	assert.Equal(t, "cluster-123", capturedReq.GetClusterId())
	assert.Contains(t, stdout, "cluster-123")
	assert.Contains(t, stdout, "suspended")
}

func TestSuspend_MissingArgs(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	_, _, err := testutil.Exec(t, env, "cluster", "suspend")
	require.Error(t, err)
}
