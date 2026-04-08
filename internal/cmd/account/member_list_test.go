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

func TestMemberList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))

	env.AccountServer.ListAccountMembersCalls.Returns(&accountv1.ListAccountMembersResponse{
		Items: []*accountv1.AccountMember{
			{
				AccountMember: &iamv1.User{
					Id:        "user-001",
					Email:     "owner@example.com",
					CreatedAt: timestamppb.New(time.Now().Add(-72 * time.Hour)),
				},
				IsOwner: true,
			},
			{
				AccountMember: &iamv1.User{
					Id:        "user-002",
					Email:     "member@example.com",
					CreatedAt: timestamppb.New(time.Now().Add(-24 * time.Hour)),
				},
				IsOwner: false,
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "account", "member", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ID")
	assert.Contains(t, stdout, "EMAIL")
	assert.Contains(t, stdout, "OWNER")
	assert.Contains(t, stdout, "CREATED")
	assert.Contains(t, stdout, "user-001")
	assert.Contains(t, stdout, "owner@example.com")
	assert.Contains(t, stdout, "yes")
	assert.Contains(t, stdout, "user-002")
	assert.Contains(t, stdout, "member@example.com")

	req, ok := env.AccountServer.ListAccountMembersCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
}

func TestMemberList_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AccountServer.ListAccountMembersCalls.Returns(&accountv1.ListAccountMembersResponse{
		Items: []*accountv1.AccountMember{
			{
				AccountMember: &iamv1.User{Id: "user-json", Email: "json@example.com"},
				IsOwner:       true,
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "account", "member", "list", "--json")
	require.NoError(t, err)

	var result struct {
		Items []struct {
			AccountMember struct {
				ID    string `json:"id"`
				Email string `json:"email"`
			} `json:"accountMember"`
			IsOwner bool `json:"isOwner"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Items, 1)
	assert.Equal(t, "user-json", result.Items[0].AccountMember.ID)
	assert.True(t, result.Items[0].IsOwner)
}

func TestMemberList_BackendError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AccountServer.ListAccountMembersCalls.Returns(nil, fmt.Errorf("internal error"))

	_, _, err := testutil.Exec(t, env, "account", "member", "list")
	require.Error(t, err)
}

func TestMemberList_Empty(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AccountServer.ListAccountMembersCalls.Returns(&accountv1.ListAccountMembersResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "account", "member", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ID")
	assert.Contains(t, stdout, "EMAIL")
}
