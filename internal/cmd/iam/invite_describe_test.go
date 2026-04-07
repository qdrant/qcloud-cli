package iam_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestInviteDescribe(t *testing.T) {
	env := testutil.NewTestEnv(t)

	inviteID := testUserID
	env.AccountServer.GetAccountInviteCalls.Returns(&accountv1.GetAccountInviteResponse{
		AccountInvite: &accountv1.AccountInvite{
			Id:          inviteID,
			UserEmail:   "alice@example.com",
			Status:      accountv1.AccountInviteStatus_ACCOUNT_INVITE_STATUS_PENDING,
			UserRoleIds: []string{"role-1"},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "invite", "describe", inviteID)
	require.NoError(t, err)

	assert.Contains(t, stdout, inviteID)
	assert.Contains(t, stdout, "alice@example.com")
	assert.Contains(t, stdout, "PENDING")
	assert.Contains(t, stdout, "role-1")
}

func TestInviteDescribe_Error(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AccountServer.GetAccountInviteCalls.Returns(nil, fmt.Errorf("not found"))

	_, _, err := testutil.Exec(t, env, "iam", "invite", "describe", testUserID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
