package cluster_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestListClusters_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.Server.ListClustersFunc = func(_ context.Context, req *clusterv1.ListClustersRequest) (*clusterv1.ListClustersResponse, error) {
		return &clusterv1.ListClustersResponse{
			Items: []*clusterv1.Cluster{
				{
					Id:                    "cluster-1",
					Name:                  "my-cluster",
					CloudProviderId:       "aws",
					CloudProviderRegionId: "us-east-1",
					State:                 &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY},
					Configuration:         &clusterv1.ClusterConfiguration{Version: ptrStr("1.8.0")},
				},
				{
					Id:                    "cluster-2",
					Name:                  "other-cluster",
					CloudProviderId:       "gcp",
					CloudProviderRegionId: "europe-west1",
				},
			},
		}, nil
	}

	stdout, _, err := testutil.Exec(t, env, "cluster", "list")
	require.NoError(t, err)

	assert.Contains(t, stdout, "cluster-1")
	assert.Contains(t, stdout, "my-cluster")
	assert.Contains(t, stdout, "aws")
	assert.Contains(t, stdout, "us-east-1")
	assert.Contains(t, stdout, "CLUSTER_PHASE_HEALTHY")
	assert.Contains(t, stdout, "1.8.0")

	assert.Contains(t, stdout, "cluster-2")
	assert.Contains(t, stdout, "other-cluster")
	assert.Contains(t, stdout, "gcp")
	assert.Contains(t, stdout, "europe-west1")
}

func TestListClusters_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.Server.ListClustersFunc = func(_ context.Context, req *clusterv1.ListClustersRequest) (*clusterv1.ListClustersResponse, error) {
		return &clusterv1.ListClustersResponse{
			Items: []*clusterv1.Cluster{
				{
					Id:   "json-cluster",
					Name: "json-name",
				},
			},
		}, nil
	}

	stdout, _, err := testutil.Exec(t, env, "cluster", "list", "--json")
	require.NoError(t, err)

	assert.Contains(t, stdout, `"id"`)
	assert.Contains(t, stdout, `"json-cluster"`)
	assert.Contains(t, stdout, `"json-name"`)
}

func TestListClusters_EmptyResponse(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.Server.ListClustersFunc = func(_ context.Context, req *clusterv1.ListClustersRequest) (*clusterv1.ListClustersResponse, error) {
		return &clusterv1.ListClustersResponse{}, nil
	}

	stdout, _, err := testutil.Exec(t, env, "cluster", "list")
	require.NoError(t, err)

	// Table header should still be present, but no data rows.
	assert.Contains(t, stdout, "ID")
	assert.Contains(t, stdout, "NAME")
}

func TestListClusters_AuthMetadata(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAPIKey("my-secret-key"))
	t.Cleanup(env.Cleanup)

	env.Server.ListClustersFunc = func(_ context.Context, req *clusterv1.ListClustersRequest) (*clusterv1.ListClustersResponse, error) {
		return &clusterv1.ListClustersResponse{}, nil
	}

	_, _, err := testutil.Exec(t, env, "cluster", "list")
	require.NoError(t, err)

	md := env.Capture.Last()
	require.NotNil(t, md)
	authValues := md.Get("authorization")
	require.Len(t, authValues, 1)
	assert.Equal(t, "apikey my-secret-key", authValues[0])
}

func TestListClusters_AccountIDPassedToServer(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	var capturedAccountID string
	env.Server.ListClustersFunc = func(_ context.Context, req *clusterv1.ListClustersRequest) (*clusterv1.ListClustersResponse, error) {
		capturedAccountID = req.GetAccountId()
		return &clusterv1.ListClustersResponse{}, nil
	}

	_, _, err := testutil.Exec(t, env, "--account-id", "acct-42", "cluster", "list")
	require.NoError(t, err)
	assert.Equal(t, "acct-42", capturedAccountID)
}

func ptrStr(s string) *string {
	return &s
}
