package account_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestAccountList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AccountServer.ListAccountsCalls.Returns(&accountv1.ListAccountsResponse{
		Items: []*accountv1.Account{
			{
				Id:         "acct-001",
				Name:       "Production",
				OwnerEmail: "owner@example.com",
				CreatedAt:  timestamppb.New(time.Now().Add(-24 * time.Hour)),
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "account", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ID")
	assert.Contains(t, stdout, "NAME")
	assert.Contains(t, stdout, "OWNER EMAIL")
	assert.Contains(t, stdout, "CREATED")
	assert.Contains(t, stdout, "acct-001")
	assert.Contains(t, stdout, "Production")
	assert.Contains(t, stdout, "owner@example.com")
	assert.Contains(t, stdout, "ago")
}

func TestAccountList_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AccountServer.ListAccountsCalls.Returns(&accountv1.ListAccountsResponse{
		Items: []*accountv1.Account{
			{Id: "acct-json", Name: "Test Account"},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "account", "list", "--json")
	require.NoError(t, err)

	var result struct {
		Items []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Items, 1)
	assert.Equal(t, "acct-json", result.Items[0].ID)
	assert.Equal(t, "Test Account", result.Items[0].Name)
}

func TestAccountList_BackendError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AccountServer.ListAccountsCalls.Returns(nil, fmt.Errorf("service unavailable"))

	_, _, err := testutil.Exec(t, env, "account", "list")
	require.Error(t, err)
}

func TestAccountList_Empty(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AccountServer.ListAccountsCalls.Returns(&accountv1.ListAccountsResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "account", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ID")
	assert.Contains(t, stdout, "NAME")
}

func TestAccountList_NoHeaders(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AccountServer.ListAccountsCalls.Returns(&accountv1.ListAccountsResponse{
		Items: []*accountv1.Account{
			{Id: "acct-001", Name: "Production"},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "account", "list", "--no-headers")
	require.NoError(t, err)
	assert.NotContains(t, stdout, "ID")
	assert.NotContains(t, stdout, "NAME")
	assert.Contains(t, stdout, "acct-001")
}
