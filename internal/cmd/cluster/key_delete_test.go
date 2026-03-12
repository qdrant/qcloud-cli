package cluster_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterauthv2 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/auth/v2"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestKeyDelete_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))
	t.Cleanup(env.Cleanup)

	var capturedReq *clusterauthv2.DeleteDatabaseApiKeyRequest
	env.DatabaseApiKeyServer.DeleteDatabaseApiKeyFunc = func(_ context.Context, req *clusterauthv2.DeleteDatabaseApiKeyRequest) (*clusterauthv2.DeleteDatabaseApiKeyResponse, error) {
		capturedReq = req
		return &clusterauthv2.DeleteDatabaseApiKeyResponse{}, nil
	}

	stdout, _, err := testutil.Exec(t, env, "cluster", "key", "delete", "cluster-123", "key-abc", "--force")
	require.NoError(t, err)
	assert.Equal(t, "test-account-id", capturedReq.GetAccountId())
	assert.Equal(t, "cluster-123", capturedReq.GetClusterId())
	assert.Equal(t, "key-abc", capturedReq.GetDatabaseApiKeyId())
	assert.Contains(t, stdout, "key-abc")
	assert.Contains(t, stdout, "deleted")
}

func TestKeyDelete_MissingArgs(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	_, _, err := testutil.Exec(t, env, "cluster", "key", "delete", "cluster-123")
	require.Error(t, err)
}
