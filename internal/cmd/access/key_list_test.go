package access_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	authv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/auth/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestKeyList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))

	env.AuthServer.ListManagementKeysCalls.Returns(&authv1.ListManagementKeysResponse{
		Items: []*authv1.ManagementKey{
			{
				Id:        "key-abc",
				Prefix:    "abc123",
				CreatedAt: timestamppb.New(time.Now().Add(-1 * time.Hour)),
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "access", "key", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ID")
	assert.Contains(t, stdout, "PREFIX")
	assert.Contains(t, stdout, "CREATED")
	assert.Contains(t, stdout, "key-abc")
	assert.Contains(t, stdout, "abc123")
	assert.Contains(t, stdout, "ago")

	req, ok := env.AuthServer.ListManagementKeysCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
}

func TestKeyList_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AuthServer.ListManagementKeysCalls.Returns(&authv1.ListManagementKeysResponse{
		Items: []*authv1.ManagementKey{
			{Id: "key-json", Prefix: "pref"},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "access", "key", "list", "--json")
	require.NoError(t, err)

	var result struct {
		Items []struct {
			ID     string `json:"id"`
			Prefix string `json:"prefix"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Items, 1)
	assert.Equal(t, "key-json", result.Items[0].ID)
	assert.Equal(t, "pref", result.Items[0].Prefix)
}

func TestKeyList_BackendError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AuthServer.ListManagementKeysCalls.Returns(nil, fmt.Errorf("internal server error"))

	_, _, err := testutil.Exec(t, env, "access", "key", "list")
	require.Error(t, err)
}

func TestKeyList_Empty(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AuthServer.ListManagementKeysCalls.Returns(&authv1.ListManagementKeysResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "access", "key", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ID")
	assert.Contains(t, stdout, "PREFIX")
}
