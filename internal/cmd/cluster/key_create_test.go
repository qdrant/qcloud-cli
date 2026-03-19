package cluster_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	clusterauthv2 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/auth/v2"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestKeyCreate_Basic(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))

	env.DatabaseApiKeyServer.EXPECT().CreateDatabaseApiKey(mock.Anything, mock.MatchedBy(func(req *clusterauthv2.CreateDatabaseApiKeyRequest) bool {
		key := req.GetDatabaseApiKey()
		assert.Equal(t, "test-account-id", key.GetAccountId())
		assert.Equal(t, "cluster-123", key.GetClusterId())
		assert.Equal(t, "my-key", key.GetName())
		assert.Empty(t, key.GetAccessRules())
		return true
	})).Return(&clusterauthv2.CreateDatabaseApiKeyResponse{
		DatabaseApiKey: &clusterauthv2.DatabaseApiKey{
			Id:  "key-new",
			Key: "secret-key-value",
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "key", "create", "cluster-123", "--name", "my-key")
	require.NoError(t, err)
	assert.Contains(t, stdout, "key-new")
	assert.Contains(t, stdout, "secret-key-value")
	assert.Contains(t, stdout, "not be shown again")
}

func TestKeyCreate_WithManageAccessType(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.DatabaseApiKeyServer.EXPECT().CreateDatabaseApiKey(mock.Anything, mock.MatchedBy(func(req *clusterauthv2.CreateDatabaseApiKeyRequest) bool {
		rules := req.GetDatabaseApiKey().GetAccessRules()
		if assert.Len(t, rules, 1) {
			assert.Equal(t, clusterauthv2.GlobalAccessRuleAccessType_GLOBAL_ACCESS_RULE_ACCESS_TYPE_MANAGE, rules[0].GetGlobalAccess().GetAccessType())
		}
		return true
	})).Return(&clusterauthv2.CreateDatabaseApiKeyResponse{
		DatabaseApiKey: &clusterauthv2.DatabaseApiKey{Id: "key-manage"},
	}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "key", "create", "cluster-123", "--name", "manage-key", "--access-type", "manage")
	require.NoError(t, err)
}

func TestKeyCreate_WithReadOnlyAccessType(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.DatabaseApiKeyServer.EXPECT().CreateDatabaseApiKey(mock.Anything, mock.MatchedBy(func(req *clusterauthv2.CreateDatabaseApiKeyRequest) bool {
		rules := req.GetDatabaseApiKey().GetAccessRules()
		if assert.Len(t, rules, 1) {
			assert.Equal(t, clusterauthv2.GlobalAccessRuleAccessType_GLOBAL_ACCESS_RULE_ACCESS_TYPE_READ_ONLY, rules[0].GetGlobalAccess().GetAccessType())
		}
		return true
	})).Return(&clusterauthv2.CreateDatabaseApiKeyResponse{
		DatabaseApiKey: &clusterauthv2.DatabaseApiKey{Id: "key-ro"},
	}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "key", "create", "cluster-123", "--name", "ro-key", "--access-type", "read-only")
	require.NoError(t, err)
}

func TestKeyCreate_InvalidAccessType(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "cluster", "key", "create", "cluster-123", "--name", "bad-key", "--access-type", "superuser")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "superuser")
}

func TestKeyCreate_WithExpires(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.DatabaseApiKeyServer.EXPECT().CreateDatabaseApiKey(mock.Anything, mock.MatchedBy(func(req *clusterauthv2.CreateDatabaseApiKeyRequest) bool {
		expiresAt := req.GetDatabaseApiKey().GetExpiresAt()
		if assert.NotNil(t, expiresAt) {
			assert.Equal(t, "2027-06-15", expiresAt.AsTime().UTC().Format("2006-01-02"))
		}
		return true
	})).Return(&clusterauthv2.CreateDatabaseApiKeyResponse{
		DatabaseApiKey: &clusterauthv2.DatabaseApiKey{Id: "key-exp"},
	}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "key", "create", "cluster-123", "--name", "exp-key", "--expires", "2027-06-15")
	require.NoError(t, err)
}

func TestKeyCreate_InvalidExpires(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "cluster", "key", "create", "cluster-123", "--name", "bad-key", "--expires", "not-a-date")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "YYYY-MM-DD")
}

func TestKeyCreate_MissingName(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "cluster", "key", "create", "cluster-123")
	require.Error(t, err)
}
