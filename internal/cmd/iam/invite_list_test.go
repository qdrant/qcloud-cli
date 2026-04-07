package iam_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestInviteList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AccountServer.ListAccountInvitesCalls.Returns(&accountv1.ListAccountInvitesResponse{
		Items: []*accountv1.AccountInvite{
			{
				Id:        "invite-1",
				UserEmail: "alice@example.com",
				Status:    accountv1.AccountInviteStatus_ACCOUNT_INVITE_STATUS_PENDING,
				CreatedAt: timestamppb.Now(),
			},
			{
				Id:        "invite-2",
				UserEmail: "bob@example.com",
				Status:    accountv1.AccountInviteStatus_ACCOUNT_INVITE_STATUS_ACCEPTED,
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "invite", "list")
	require.NoError(t, err)

	assert.Contains(t, stdout, "invite-1")
	assert.Contains(t, stdout, "alice@example.com")
	assert.Contains(t, stdout, "PENDING")
	assert.Contains(t, stdout, "invite-2")
	assert.Contains(t, stdout, "bob@example.com")
	assert.Contains(t, stdout, "ACCEPTED")

	req, ok := env.AccountServer.ListAccountInvitesCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
}

func TestInviteList_Error(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AccountServer.ListAccountInvitesCalls.Returns(nil, fmt.Errorf("permission denied"))

	_, _, err := testutil.Exec(t, env, "iam", "invite", "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}
