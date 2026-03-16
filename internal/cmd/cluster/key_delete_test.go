package cluster_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterauthv2 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/auth/v2"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestKeyDelete_WithForce(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))

	env.DatabaseApiKeyServer.DeleteDatabaseApiKeyCalls.Returns(&clusterauthv2.DeleteDatabaseApiKeyResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "key", "delete", "cluster-123", "key-abc", "--force")
	require.NoError(t, err)
	assert.Contains(t, stdout, "key-abc")
	assert.Contains(t, stdout, "deleted")

	req, ok := env.DatabaseApiKeyServer.DeleteDatabaseApiKeyCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "cluster-123", req.GetClusterId())
	assert.Equal(t, "key-abc", req.GetDatabaseApiKeyId())
}

func TestKeyDelete_MissingArgs(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "cluster", "key", "delete", "cluster-123")
	require.Error(t, err)
}
