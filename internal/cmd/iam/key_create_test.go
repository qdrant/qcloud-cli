package iam_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/auth/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestKeyCreate_PrintsIDAndKey(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))

	env.AuthServer.CreateManagementKeyCalls.Returns(&authv1.CreateManagementKeyResponse{
		ManagementKey: &authv1.ManagementKey{
			Id:  "new-key-id",
			Key: "super-secret-value",
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "key", "create")
	require.NoError(t, err)
	assert.Contains(t, stdout, "new-key-id")
	assert.Contains(t, stdout, "super-secret-value")
	assert.Contains(t, stdout, "Save this key now")

	req, ok := env.AuthServer.CreateManagementKeyCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetManagementKey().GetAccountId())
}

func TestKeyCreate_BackendError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AuthServer.CreateManagementKeyCalls.Returns(nil, fmt.Errorf("internal server error"))

	_, _, err := testutil.Exec(t, env, "iam", "key", "create")
	require.Error(t, err)
}

func TestKeyCreate_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AuthServer.CreateManagementKeyCalls.Returns(&authv1.CreateManagementKeyResponse{
		ManagementKey: &authv1.ManagementKey{
			Id:  "json-key-id",
			Key: "secret",
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "key", "create", "--json")
	require.NoError(t, err)

	var result struct {
		ID  string `json:"id"`
		Key string `json:"key"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "json-key-id", result.ID)
	assert.Equal(t, "secret", result.Key)
}
