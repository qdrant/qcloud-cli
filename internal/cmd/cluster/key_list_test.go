package cluster_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	clusterauthv2 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/auth/v2"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestKeyList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))
	t.Cleanup(env.Cleanup)

	expires := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)

	env.DatabaseApiKeyServer.ListDatabaseApiKeysFunc = func(_ context.Context, req *clusterauthv2.ListDatabaseApiKeysRequest) (*clusterauthv2.ListDatabaseApiKeysResponse, error) {
		assert.Equal(t, "test-account-id", req.GetAccountId())
		assert.Equal(t, "cluster-123", req.GetClusterId())
		return &clusterauthv2.ListDatabaseApiKeysResponse{
			Items: []*clusterauthv2.DatabaseApiKey{
				{
					Id:        "key-abc",
					Name:      "my-key",
					Postfix:   "xyz",
					CreatedAt: timestamppb.New(time.Now().Add(-1 * time.Hour)),
					ExpiresAt: timestamppb.New(expires),
				},
			},
		}, nil
	}

	stdout, _, err := testutil.Exec(t, env, "cluster", "key", "list", "cluster-123")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ID")
	assert.Contains(t, stdout, "NAME")
	assert.Contains(t, stdout, "POSTFIX")
	assert.Contains(t, stdout, "CREATED")
	assert.Contains(t, stdout, "EXPIRES")
	assert.Contains(t, stdout, "key-abc")
	assert.Contains(t, stdout, "my-key")
	assert.Contains(t, stdout, "xyz")
	assert.Contains(t, stdout, "ago")
	assert.Contains(t, stdout, "2027-01-01")
}

func TestKeyList_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.DatabaseApiKeyServer.ListDatabaseApiKeysFunc = func(_ context.Context, _ *clusterauthv2.ListDatabaseApiKeysRequest) (*clusterauthv2.ListDatabaseApiKeysResponse, error) {
		return &clusterauthv2.ListDatabaseApiKeysResponse{
			Items: []*clusterauthv2.DatabaseApiKey{
				{Id: "key-json", Name: "json-key"},
			},
		}, nil
	}

	stdout, _, err := testutil.Exec(t, env, "cluster", "key", "list", "cluster-123", "--json")
	require.NoError(t, err)

	var result struct {
		Items []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Items, 1)
	assert.Equal(t, "key-json", result.Items[0].ID)
	assert.Equal(t, "json-key", result.Items[0].Name)
}

func TestKeyList_EmptyResponse(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	env.DatabaseApiKeyServer.ListDatabaseApiKeysFunc = func(_ context.Context, _ *clusterauthv2.ListDatabaseApiKeysRequest) (*clusterauthv2.ListDatabaseApiKeysResponse, error) {
		return &clusterauthv2.ListDatabaseApiKeysResponse{}, nil
	}

	stdout, _, err := testutil.Exec(t, env, "cluster", "key", "list", "cluster-123")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ID")
	assert.Contains(t, stdout, "NAME")
}

func TestKeyList_ClusterIDPassedToServer(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	var capturedClusterID string
	env.DatabaseApiKeyServer.ListDatabaseApiKeysFunc = func(_ context.Context, req *clusterauthv2.ListDatabaseApiKeysRequest) (*clusterauthv2.ListDatabaseApiKeysResponse, error) {
		capturedClusterID = req.GetClusterId()
		return &clusterauthv2.ListDatabaseApiKeysResponse{}, nil
	}

	_, _, err := testutil.Exec(t, env, "cluster", "key", "list", "my-cluster-id")
	require.NoError(t, err)
	assert.Equal(t, "my-cluster-id", capturedClusterID)
}
