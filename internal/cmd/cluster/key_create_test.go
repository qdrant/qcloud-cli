package cluster_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterauthv2 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/auth/v2"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestKeyCreate_Basic(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))
	t.Cleanup(env.Cleanup)

	var capturedKey *clusterauthv2.DatabaseApiKey
	env.DatabaseApiKeyServer.CreateDatabaseApiKeyFunc = func(_ context.Context, req *clusterauthv2.CreateDatabaseApiKeyRequest) (*clusterauthv2.CreateDatabaseApiKeyResponse, error) {
		capturedKey = req.GetDatabaseApiKey()
		return &clusterauthv2.CreateDatabaseApiKeyResponse{
			DatabaseApiKey: &clusterauthv2.DatabaseApiKey{
				Id:   "key-new",
				Name: req.GetDatabaseApiKey().GetName(),
				Key:  "secret-key-value",
			},
		}, nil
	}

	stdout, _, err := testutil.Exec(t, env, "cluster", "key", "create", "cluster-123", "--name", "my-key")
	require.NoError(t, err)
	assert.Equal(t, "test-account-id", capturedKey.GetAccountId())
	assert.Equal(t, "cluster-123", capturedKey.GetClusterId())
	assert.Equal(t, "my-key", capturedKey.GetName())
	assert.Empty(t, capturedKey.GetAccessRules())
	assert.Contains(t, stdout, "key-new")
	assert.Contains(t, stdout, "secret-key-value")
	assert.Contains(t, stdout, "not be shown again")
}

func TestKeyCreate_WithManageAccessType(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	var capturedKey *clusterauthv2.DatabaseApiKey
	env.DatabaseApiKeyServer.CreateDatabaseApiKeyFunc = func(_ context.Context, req *clusterauthv2.CreateDatabaseApiKeyRequest) (*clusterauthv2.CreateDatabaseApiKeyResponse, error) {
		capturedKey = req.GetDatabaseApiKey()
		return &clusterauthv2.CreateDatabaseApiKeyResponse{
			DatabaseApiKey: &clusterauthv2.DatabaseApiKey{Id: "key-manage"},
		}, nil
	}

	_, _, err := testutil.Exec(t, env, "cluster", "key", "create", "cluster-123", "--name", "manage-key", "--access-type", "manage")
	require.NoError(t, err)
	require.Len(t, capturedKey.GetAccessRules(), 1)
	globalAccess := capturedKey.GetAccessRules()[0].GetGlobalAccess()
	require.NotNil(t, globalAccess)
	assert.Equal(t, clusterauthv2.GlobalAccessRuleAccessType_GLOBAL_ACCESS_RULE_ACCESS_TYPE_MANAGE, globalAccess.GetAccessType())
}

func TestKeyCreate_WithReadOnlyAccessType(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	var capturedKey *clusterauthv2.DatabaseApiKey
	env.DatabaseApiKeyServer.CreateDatabaseApiKeyFunc = func(_ context.Context, req *clusterauthv2.CreateDatabaseApiKeyRequest) (*clusterauthv2.CreateDatabaseApiKeyResponse, error) {
		capturedKey = req.GetDatabaseApiKey()
		return &clusterauthv2.CreateDatabaseApiKeyResponse{
			DatabaseApiKey: &clusterauthv2.DatabaseApiKey{Id: "key-ro"},
		}, nil
	}

	_, _, err := testutil.Exec(t, env, "cluster", "key", "create", "cluster-123", "--name", "ro-key", "--access-type", "read-only")
	require.NoError(t, err)
	require.Len(t, capturedKey.GetAccessRules(), 1)
	globalAccess := capturedKey.GetAccessRules()[0].GetGlobalAccess()
	require.NotNil(t, globalAccess)
	assert.Equal(t, clusterauthv2.GlobalAccessRuleAccessType_GLOBAL_ACCESS_RULE_ACCESS_TYPE_READ_ONLY, globalAccess.GetAccessType())
}

func TestKeyCreate_InvalidAccessType(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	_, _, err := testutil.Exec(t, env, "cluster", "key", "create", "cluster-123", "--name", "bad-key", "--access-type", "superuser")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "superuser")
}

func TestKeyCreate_WithExpires(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	var capturedKey *clusterauthv2.DatabaseApiKey
	env.DatabaseApiKeyServer.CreateDatabaseApiKeyFunc = func(_ context.Context, req *clusterauthv2.CreateDatabaseApiKeyRequest) (*clusterauthv2.CreateDatabaseApiKeyResponse, error) {
		capturedKey = req.GetDatabaseApiKey()
		return &clusterauthv2.CreateDatabaseApiKeyResponse{
			DatabaseApiKey: &clusterauthv2.DatabaseApiKey{Id: "key-exp"},
		}, nil
	}

	_, _, err := testutil.Exec(t, env, "cluster", "key", "create", "cluster-123", "--name", "exp-key", "--expires", "2027-06-15")
	require.NoError(t, err)
	require.NotNil(t, capturedKey.GetExpiresAt())
	assert.Equal(t, "2027-06-15", capturedKey.GetExpiresAt().AsTime().UTC().Format("2006-01-02"))
}

func TestKeyCreate_InvalidExpires(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	_, _, err := testutil.Exec(t, env, "cluster", "key", "create", "cluster-123", "--name", "bad-key", "--expires", "not-a-date")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "YYYY-MM-DD")
}

func TestKeyCreate_MissingName(t *testing.T) {
	env := testutil.NewTestEnv(t)
	t.Cleanup(env.Cleanup)

	_, _, err := testutil.Exec(t, env, "cluster", "key", "create", "cluster-123")
	require.Error(t, err)
}
