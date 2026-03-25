package hybrid_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestHybridClusterList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	version := "1.8.0"
	env.Server.ListClustersCalls.Returns(&clusterv1.ListClustersResponse{
		Items: []*clusterv1.Cluster{
			{
				Id:                    "cluster-1",
				Name:                  "my-cluster",
				CloudProviderRegionId: "env-123",
				State:                 &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY},
				Configuration:         &clusterv1.ClusterConfiguration{Version: &version},
			},
			{
				Id:   "cluster-2",
				Name: "other-cluster",
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "cluster", "list")
	require.NoError(t, err)

	assert.Contains(t, stdout, "cluster-1")
	assert.Contains(t, stdout, "my-cluster")
	assert.Contains(t, stdout, "HEALTHY")
	assert.Contains(t, stdout, "1.8.0")
	assert.Contains(t, stdout, "env-123")
	assert.Contains(t, stdout, "cluster-2")
	assert.Contains(t, stdout, "other-cluster")
}

func TestHybridClusterList_WithEnvID(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.ListClustersCalls.Returns(&clusterv1.ListClustersResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "hybrid", "cluster", "list", "--env-id", "env-123")
	require.NoError(t, err)

	req, ok := env.Server.ListClustersCalls.Last()
	require.True(t, ok)
	require.NotNil(t, req.CloudProviderRegionId)
	assert.Equal(t, "env-123", *req.CloudProviderRegionId)
	require.NotNil(t, req.CloudProviderId)
	assert.Equal(t, "hybrid", *req.CloudProviderId)
}

func TestHybridClusterList_AutoPaginate(t *testing.T) {
	env := testutil.NewTestEnv(t)

	token := "page-2-token"
	env.Server.ListClustersCalls.
		OnCall(0, func(_ context.Context, _ *clusterv1.ListClustersRequest) (*clusterv1.ListClustersResponse, error) {
			return &clusterv1.ListClustersResponse{
				Items:         []*clusterv1.Cluster{{Id: "cluster-1", Name: "first"}},
				NextPageToken: &token,
			}, nil
		}).
		OnCall(1, func(_ context.Context, _ *clusterv1.ListClustersRequest) (*clusterv1.ListClustersResponse, error) {
			return &clusterv1.ListClustersResponse{
				Items: []*clusterv1.Cluster{{Id: "cluster-2", Name: "second"}},
			}, nil
		})

	stdout, _, err := testutil.Exec(t, env, "hybrid", "cluster", "list")
	require.NoError(t, err)

	assert.Equal(t, 2, env.Server.ListClustersCalls.Count())
	assert.Contains(t, stdout, "cluster-1")
	assert.Contains(t, stdout, "cluster-2")
}

func TestHybridClusterList_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.ListClustersCalls.Returns(&clusterv1.ListClustersResponse{
		Items: []*clusterv1.Cluster{
			{Id: "json-cluster", Name: "json-name"},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "hybrid", "cluster", "list", "--json")
	require.NoError(t, err)

	assert.Contains(t, stdout, `"id"`)
	assert.Contains(t, stdout, `"json-cluster"`)
}
