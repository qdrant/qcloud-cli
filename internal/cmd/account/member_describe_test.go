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
	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestMemberDescribe_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))

	env.AccountServer.GetAccountMemberCalls.Returns(&accountv1.GetAccountMemberResponse{
		AccountMember: &accountv1.AccountMember{
			AccountMember: &iamv1.User{
				Id:        "user-001",
				Email:     "owner@example.com",
				CreatedAt: timestamppb.New(time.Now().Add(-48 * time.Hour)),
			},
			IsOwner: true,
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "account", "member", "describe", "user-001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "user-001")
	assert.Contains(t, stdout, "owner@example.com")
	assert.Contains(t, stdout, "yes")

	req, ok := env.AccountServer.GetAccountMemberCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "user-001", req.GetUserId())
}

func TestMemberDescribe_NonOwner(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AccountServer.GetAccountMemberCalls.Returns(&accountv1.GetAccountMemberResponse{
		AccountMember: &accountv1.AccountMember{
			AccountMember: &iamv1.User{
				Id:    "user-002",
				Email: "member@example.com",
			},
			IsOwner: false,
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "account", "member", "describe", "user-002")
	require.NoError(t, err)
	assert.Contains(t, stdout, "user-002")
	assert.Contains(t, stdout, "no")
}

func TestMemberDescribe_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AccountServer.GetAccountMemberCalls.Returns(&accountv1.GetAccountMemberResponse{
		AccountMember: &accountv1.AccountMember{
			AccountMember: &iamv1.User{Id: "user-json", Email: "json@example.com"},
			IsOwner:       true,
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "account", "member", "describe", "user-json", "--json")
	require.NoError(t, err)

	var result struct {
		AccountMember struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"accountMember"`
		IsOwner bool `json:"isOwner"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "user-json", result.AccountMember.ID)
	assert.True(t, result.IsOwner)
}

func TestMemberDescribe_BackendError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AccountServer.GetAccountMemberCalls.Returns(nil, fmt.Errorf("not found"))

	_, _, err := testutil.Exec(t, env, "account", "member", "describe", "user-bad")
	require.Error(t, err)
}

func TestMemberDescribe_MissingArgs(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "account", "member", "describe")
	require.Error(t, err)
}
