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
	t.Cleanup(env.Cleanup)

	env.Server.GetClusterFunc = func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{Id: req.GetClusterId(), Name: "my-cluster"},
		}, nil
	}

	var capturedLabels map[string]string
	env.Server.UpdateClusterFunc = func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		capturedLabels = make(map[string]string)
		for _, kv := range req.GetCluster().GetLabels() {
			capturedLabels[kv.GetKey()] = kv.GetValue()
		}
		return &clusterv1.UpdateClusterResponse{
			Cluster: req.GetCluster(),
		}, nil
	}

	stdout, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--label", "env=prod",
		"--label", "team=platform",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "cluster-abc")
	assert.Contains(t, stdout, "updated successfully")
	assert.Equal(t, map[string]string{"env": "prod", "team": "platform"}, capturedLabels)
}

func TestUpdateCluster_ClearLabels(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.Server.GetClusterFunc = func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{Id: req.GetClusterId(), Name: "my-cluster"},
		}, nil
	}

	var capturedLabelCount int
	env.Server.UpdateClusterFunc = func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		capturedLabelCount = len(req.GetCluster().GetLabels())
		return &clusterv1.UpdateClusterResponse{
			Cluster: req.GetCluster(),
		}, nil
	}

	_, _, err := testutil.Exec(t, env, "cluster", "update", "cluster-abc")
	require.NoError(t, err)
	assert.Equal(t, 0, capturedLabelCount)
}

func TestUpdateCluster_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.Server.GetClusterFunc = func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{Id: req.GetClusterId()},
		}, nil
	}

	env.Server.UpdateClusterFunc = func(_ context.Context, _ *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		return nil, fmt.Errorf("internal server error")
	}

	_, _, err := testutil.Exec(t, env, "cluster", "update", "cluster-abc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update cluster")
}
