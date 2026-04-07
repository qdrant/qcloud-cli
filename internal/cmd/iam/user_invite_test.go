package iam_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"
	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestUserInvite(t *testing.T) {
	env := testutil.NewTestEnv(t)

	inviteID := "invite-id-123"
	env.AccountServer.CreateAccountInviteCalls.Returns(&accountv1.CreateAccountInviteResponse{
		AccountInvite: &accountv1.AccountInvite{Id: inviteID, UserEmail: "bob@example.com"},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "user", "invite",
		"--email", "bob@example.com")
	require.NoError(t, err)
	assert.Contains(t, stdout, inviteID)
	assert.Contains(t, stdout, "bob@example.com")

	req, ok := env.AccountServer.CreateAccountInviteCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "bob@example.com", req.GetAccountInvite().GetUserEmail())
	assert.Equal(t, "test-account-id", req.GetAccountInvite().GetAccountId())
}

func TestUserInvite_WithRole(t *testing.T) {
	env := testutil.NewTestEnv(t)

	roleID := testRoleID
	env.IAMServer.ListRolesCalls.Returns(&iamv1.ListRolesResponse{
		Items: []*iamv1.Role{{Id: roleID, Name: "viewer"}},
	}, nil)
	env.AccountServer.CreateAccountInviteCalls.Returns(&accountv1.CreateAccountInviteResponse{
		AccountInvite: &accountv1.AccountInvite{Id: "invite-id", UserEmail: "bob@example.com"},
	}, nil)

	_, _, err := testutil.Exec(t, env, "iam", "user", "invite",
		"--email", "bob@example.com", "--role", "viewer")
	require.NoError(t, err)

	req, ok := env.AccountServer.CreateAccountInviteCalls.Last()
	require.True(t, ok)
	assert.Equal(t, []string{roleID}, req.GetAccountInvite().GetUserRoleIds())
}

func TestUserInvite_MissingEmail(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "iam", "user", "invite")
	require.Error(t, err)
}
