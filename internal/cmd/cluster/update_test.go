package cluster_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestUpdateCluster_SetLabels(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Always(func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{Id: req.GetClusterId(), Name: "my-cluster"},
		}, nil
	})
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		return &clusterv1.UpdateClusterResponse{
			Cluster: req.GetCluster(),
		}, nil
	})

	stdout, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--label", "env=prod",
		"--label", "team=platform",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "cluster-abc")
	assert.Contains(t, stdout, "updated successfully")

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	capturedLabels := make(map[string]string)
	for _, kv := range req.GetCluster().GetLabels() {
		capturedLabels[kv.GetKey()] = kv.GetValue()
	}
	assert.Equal(t, map[string]string{"env": "prod", "team": "platform"}, capturedLabels)
}

func TestUpdateCluster_ClearLabels(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Always(func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{Id: req.GetClusterId(), Name: "my-cluster"},
		}, nil
	})
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		return &clusterv1.UpdateClusterResponse{
			Cluster: req.GetCluster(),
		}, nil
	})

	_, _, err := testutil.Exec(t, env, "cluster", "update", "cluster-abc")
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	assert.Empty(t, req.GetCluster().GetLabels())
}

func TestUpdateCluster_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Always(func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{Id: req.GetClusterId()},
		}, nil
	})
	env.Server.UpdateClusterCalls.Returns(nil, fmt.Errorf("internal server error"))

	_, _, err := testutil.Exec(t, env, "cluster", "update", "cluster-abc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update cluster")
}
