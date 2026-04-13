package cluster_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestListClusters_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.ListClustersCalls.Returns(&clusterv1.ListClustersResponse{
		Items: []*clusterv1.Cluster{
			{
				Id:                    "cluster-1",
				Name:                  "my-cluster",
				CloudProviderId:       "aws",
				CloudProviderRegionId: "us-east-1",
				State:                 &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY},
				Configuration:         &clusterv1.ClusterConfiguration{Version: new("1.8.0")},
				CreatedAt:             timestamppb.New(time.Now().Add(-3 * time.Hour)),
			},
			{
				Id:                    "cluster-2",
				Name:                  "other-cluster",
				CloudProviderId:       "gcp",
				CloudProviderRegionId: "europe-west1",
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "list")
	require.NoError(t, err)

	assert.Contains(t, stdout, "cluster-1")
	assert.Contains(t, stdout, "my-cluster")
	assert.Contains(t, stdout, "aws")
	assert.Contains(t, stdout, "us-east-1")
	assert.Contains(t, stdout, "HEALTHY")
	assert.Contains(t, stdout, "1.8.0")

	assert.Contains(t, stdout, "ago")

	assert.Contains(t, stdout, "cluster-2")
	assert.Contains(t, stdout, "other-cluster")
	assert.Contains(t, stdout, "gcp")
	assert.Contains(t, stdout, "europe-west1")
}

func TestListClusters_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.ListClustersCalls.Returns(&clusterv1.ListClustersResponse{
		Items: []*clusterv1.Cluster{
			{
				Id:   "json-cluster",
				Name: "json-name",
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "list", "--json")
	require.NoError(t, err)

	assert.Contains(t, stdout, `"id"`)
	assert.Contains(t, stdout, `"json-cluster"`)
	assert.Contains(t, stdout, `"json-name"`)
}

func TestListClusters_EmptyResponse(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.ListClustersCalls.Returns(&clusterv1.ListClustersResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "list")
	require.NoError(t, err)

	// Table header should still be present, but no data rows.
	assert.Contains(t, stdout, "ID")
	assert.Contains(t, stdout, "NAME")
}

func TestListClusters_AuthMetadata(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAPIKey("my-secret-key"))

	env.Server.ListClustersCalls.Returns(&clusterv1.ListClustersResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "list")
	require.NoError(t, err)

	md := env.Capture.Last()
	require.NotNil(t, md)
	authValues := md.Get("authorization")
	require.Len(t, authValues, 1)
	assert.Equal(t, "apikey my-secret-key", authValues[0])
}

func TestListClusters_UserAgent(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithVersion("1.2.3"))

	env.Server.ListClustersCalls.Returns(&clusterv1.ListClustersResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "list")
	require.NoError(t, err)

	md := env.Capture.Last()
	require.NotNil(t, md)
	userAgent := md.Get("user-agent")
	require.NotEmpty(t, userAgent)
	assert.Contains(t, userAgent[0], "qcloud-cli/1.2.3")
}

func TestListClusters_AccountIDPassedToServer(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.ListClustersCalls.Returns(&clusterv1.ListClustersResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "list")
	require.NoError(t, err)

	req, ok := env.Server.ListClustersCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
}

func TestListClusters_AutoPaginateMultiplePages(t *testing.T) {
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

	stdout, _, err := testutil.Exec(t, env, "cluster", "list")
	require.NoError(t, err)

	assert.Equal(t, 2, env.Server.ListClustersCalls.Count())
	assert.Contains(t, stdout, "cluster-1")
	assert.Contains(t, stdout, "cluster-2")
	// No next page token footer when auto-paginating.
	assert.NotContains(t, stdout, "Next page token")
}

func TestListClusters_PageSizeFlagSingleRequest(t *testing.T) {
	env := testutil.NewTestEnv(t)

	token := "next-token"
	env.Server.ListClustersCalls.Returns(&clusterv1.ListClustersResponse{
		Items:         []*clusterv1.Cluster{{Id: "cluster-1"}},
		NextPageToken: &token,
	}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "list", "--page-size", "1")
	require.NoError(t, err)

	assert.Equal(t, 1, env.Server.ListClustersCalls.Count())
	req, ok := env.Server.ListClustersCalls.Last()
	require.True(t, ok)
	require.NotNil(t, req.PageSize)
	assert.Equal(t, int32(1), *req.PageSize)
}

func TestListClusters_PageTokenFlagSingleRequest(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.ListClustersCalls.Returns(&clusterv1.ListClustersResponse{
		Items: []*clusterv1.Cluster{{Id: "cluster-2"}},
	}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "list", "--page-token", "my-token")
	require.NoError(t, err)

	assert.Equal(t, 1, env.Server.ListClustersCalls.Count())
	req, ok := env.Server.ListClustersCalls.Last()
	require.True(t, ok)
	require.NotNil(t, req.PageToken)
	assert.Equal(t, "my-token", *req.PageToken)
}

func TestListClusters_CloudProviderFilter(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.ListClustersCalls.Returns(&clusterv1.ListClustersResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "list", "--cloud-provider", "aws")
	require.NoError(t, err)

	req, ok := env.Server.ListClustersCalls.Last()
	require.True(t, ok)
	require.NotNil(t, req.CloudProviderId)
	assert.Equal(t, "aws", *req.CloudProviderId)
	assert.Nil(t, req.CloudProviderRegionId)
}

func TestListClusters_CloudRegionFilter(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.ListClustersCalls.Returns(&clusterv1.ListClustersResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "list", "--cloud-provider", "aws", "--cloud-region", "us-east-1")
	require.NoError(t, err)

	req, ok := env.Server.ListClustersCalls.Last()
	require.True(t, ok)
	require.NotNil(t, req.CloudProviderId)
	assert.Equal(t, "aws", *req.CloudProviderId)
	require.NotNil(t, req.CloudProviderRegionId)
	assert.Equal(t, "us-east-1", *req.CloudProviderRegionId)
}
