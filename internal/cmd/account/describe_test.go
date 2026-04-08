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

func TestAccountDescribe_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	now := time.Now()
	env.AccountServer.GetAccountCalls.Returns(&accountv1.GetAccountResponse{
		Account: &accountv1.Account{
			Id:         "acct-001",
			Name:       "Production",
			OwnerEmail: "owner@example.com",
			Company: &accountv1.Company{
				Name:   "Acme Corp",
				Domain: new("acme.com"),
			},
			Privileges: []string{"premium"},
			CreatedAt:  timestamppb.New(now.Add(-48 * time.Hour)),
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "account", "describe", "acct-001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "acct-001")
	assert.Contains(t, stdout, "Production")
	assert.Contains(t, stdout, "owner@example.com")
	assert.Contains(t, stdout, "Acme Corp")
	assert.Contains(t, stdout, "acme.com")
	assert.Contains(t, stdout, "premium")

	req, ok := env.AccountServer.GetAccountCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "acct-001", req.GetAccountId())
}

func TestAccountDescribe_DefaultsToCurrentAccount(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("default-acct"))

	env.AccountServer.GetAccountCalls.Returns(&accountv1.GetAccountResponse{
		Account: &accountv1.Account{
			Id:   "default-acct",
			Name: "Default",
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "account", "describe")
	require.NoError(t, err)
	assert.Contains(t, stdout, "default-acct")

	req, ok := env.AccountServer.GetAccountCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "default-acct", req.GetAccountId())
}

func TestAccountDescribe_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AccountServer.GetAccountCalls.Returns(&accountv1.GetAccountResponse{
		Account: &accountv1.Account{
			Id:         "acct-json",
			Name:       "JSON Account",
			OwnerEmail: "json@example.com",
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "account", "describe", "acct-json", "--json")
	require.NoError(t, err)

	var result struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		OwnerEmail string `json:"ownerEmail"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "acct-json", result.ID)
	assert.Equal(t, "JSON Account", result.Name)
	assert.Equal(t, "json@example.com", result.OwnerEmail)
}

func TestAccountDescribe_BackendError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AccountServer.GetAccountCalls.Returns(nil, fmt.Errorf("not found"))

	_, _, err := testutil.Exec(t, env, "account", "describe", "acct-bad")
	require.Error(t, err)
}

func TestAccountDescribe_TooManyArgs(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "account", "describe", "arg1", "arg2")
	require.Error(t, err)
}
