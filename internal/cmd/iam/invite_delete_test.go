package iam_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestInviteDelete(t *testing.T) {
	env := testutil.NewTestEnv(t)

	inviteID := testUserID
	env.AccountServer.DeleteAccountInviteCalls.Returns(&accountv1.DeleteAccountInviteResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "invite", "delete", inviteID, "--force")
	require.NoError(t, err)
	assert.Contains(t, stdout, "deleted")

	req, ok := env.AccountServer.DeleteAccountInviteCalls.Last()
	require.True(t, ok)
	assert.Equal(t, inviteID, req.GetInviteId())
}

func TestInviteDelete_Error(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AccountServer.DeleteAccountInviteCalls.Returns(nil, fmt.Errorf("not found"))

	_, _, err := testutil.Exec(t, env, "iam", "invite", "delete",
		testUserID, "--force")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
