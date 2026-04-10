package account_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestAccountUpdate_Name(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AccountServer.GetAccountCalls.Returns(&accountv1.GetAccountResponse{
		Account: &accountv1.Account{
			Id:   "acct-001",
			Name: "Old Name",
		},
	}, nil)
	env.AccountServer.UpdateAccountCalls.Always(func(_ context.Context, req *accountv1.UpdateAccountRequest) (*accountv1.UpdateAccountResponse, error) {
		return &accountv1.UpdateAccountResponse{Account: req.GetAccount()}, nil
	})

	stdout, _, err := testutil.Exec(t, env, "account", "update", "acct-001", "--name", "New Name")
	require.NoError(t, err)
	assert.Contains(t, stdout, "acct-001")
	assert.Contains(t, stdout, "New Name")
	assert.Contains(t, stdout, "updated successfully")

	req, ok := env.AccountServer.UpdateAccountCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "New Name", req.GetAccount().GetName())
}

func TestAccountUpdate_CompanyInfo(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AccountServer.GetAccountCalls.Returns(&accountv1.GetAccountResponse{
		Account: &accountv1.Account{
			Id:   "acct-001",
			Name: "My Account",
		},
	}, nil)
	env.AccountServer.UpdateAccountCalls.Always(func(_ context.Context, req *accountv1.UpdateAccountRequest) (*accountv1.UpdateAccountResponse, error) {
		return &accountv1.UpdateAccountResponse{Account: req.GetAccount()}, nil
	})

	_, _, err := testutil.Exec(t, env, "account", "update", "acct-001",
		"--company-name", "Acme Corp",
		"--company-domain", "acme.com",
	)
	require.NoError(t, err)

	req, ok := env.AccountServer.UpdateAccountCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "Acme Corp", req.GetAccount().GetCompany().GetName())
	assert.Equal(t, "acme.com", req.GetAccount().GetCompany().GetDomain())
}

func TestAccountUpdate_DefaultsToCurrentAccount(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("default-acct"))

	env.AccountServer.GetAccountCalls.Returns(&accountv1.GetAccountResponse{
		Account: &accountv1.Account{
			Id:   "default-acct",
			Name: "Default",
		},
	}, nil)
	env.AccountServer.UpdateAccountCalls.Always(func(_ context.Context, req *accountv1.UpdateAccountRequest) (*accountv1.UpdateAccountResponse, error) {
		return &accountv1.UpdateAccountResponse{Account: req.GetAccount()}, nil
	})

	_, _, err := testutil.Exec(t, env, "account", "update", "--name", "Updated")
	require.NoError(t, err)

	getReq, ok := env.AccountServer.GetAccountCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "default-acct", getReq.GetAccountId())
}

func TestAccountUpdate_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AccountServer.GetAccountCalls.Returns(&accountv1.GetAccountResponse{
		Account: &accountv1.Account{Id: "acct-001", Name: "Old"},
	}, nil)
	env.AccountServer.UpdateAccountCalls.Always(func(_ context.Context, req *accountv1.UpdateAccountRequest) (*accountv1.UpdateAccountResponse, error) {
		return &accountv1.UpdateAccountResponse{Account: req.GetAccount()}, nil
	})

	stdout, _, err := testutil.Exec(t, env, "account", "update", "acct-001", "--name", "New", "--json")
	require.NoError(t, err)

	var result struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "acct-001", result.ID)
	assert.Equal(t, "New", result.Name)
}

func TestAccountUpdate_BackendError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AccountServer.GetAccountCalls.Returns(&accountv1.GetAccountResponse{
		Account: &accountv1.Account{Id: "acct-001", Name: "X"},
	}, nil)
	env.AccountServer.UpdateAccountCalls.Returns(nil, fmt.Errorf("permission denied"))

	_, _, err := testutil.Exec(t, env, "account", "update", "acct-001", "--name", "Y")
	require.Error(t, err)
}

func TestAccountUpdate_PreservesUnchangedFields(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AccountServer.GetAccountCalls.Returns(&accountv1.GetAccountResponse{
		Account: &accountv1.Account{
			Id:   "acct-001",
			Name: "Original Name",
			Company: &accountv1.Company{
				Name:   "Original Corp",
				Domain: new("original.com"),
			},
		},
	}, nil)
	env.AccountServer.UpdateAccountCalls.Always(func(_ context.Context, req *accountv1.UpdateAccountRequest) (*accountv1.UpdateAccountResponse, error) {
		return &accountv1.UpdateAccountResponse{Account: req.GetAccount()}, nil
	})

	// Only change company-name, leave name and company-domain unchanged
	_, _, err := testutil.Exec(t, env, "account", "update", "acct-001", "--company-name", "New Corp")
	require.NoError(t, err)

	req, ok := env.AccountServer.UpdateAccountCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "Original Name", req.GetAccount().GetName())
	assert.Equal(t, "New Corp", req.GetAccount().GetCompany().GetName())
	assert.Equal(t, "original.com", req.GetAccount().GetCompany().GetDomain())
}
