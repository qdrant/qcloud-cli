package iam_test

import (
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

const (
	testUserID = "7b2ea926-724b-4de2-b73a-8675c42a6ebe"
	testRoleID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
)

// --- user list ---

func TestUserList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{
			{
				Id:        "user-1",
				Email:     "alice@example.com",
				Status:    iamv1.UserStatus_USER_STATUS_ACTIVE,
				CreatedAt: timestamppb.New(time.Now().Add(-48 * time.Hour)),
			},
			{
				Id:     "user-2",
				Email:  "bob@example.com",
				Status: iamv1.UserStatus_USER_STATUS_BLOCKED,
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "user", "list")
	require.NoError(t, err)

	assert.Contains(t, stdout, "user-1")
	assert.Contains(t, stdout, "alice@example.com")
	assert.Contains(t, stdout, "ACTIVE")
	assert.Contains(t, stdout, "user-2")
	assert.Contains(t, stdout, "bob@example.com")
	assert.Contains(t, stdout, "BLOCKED")

	req, ok := env.IAMServer.ListUsersCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
}

func TestUserList_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{{Id: "user-1", Email: "alice@example.com"}},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "user", "list", "--json")
	require.NoError(t, err)

	assert.Contains(t, stdout, `"id"`)
	assert.Contains(t, stdout, "user-1")
}

func TestUserList_Error(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListUsersCalls.Returns(nil, fmt.Errorf("permission denied"))

	_, _, err := testutil.Exec(t, env, "iam", "user", "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

// --- user describe ---

func TestUserDescribe_ByID(t *testing.T) {
	env := testutil.NewTestEnv(t)

	userID := testUserID
	cat := "Cluster"
	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{
			{Id: userID, Email: "alice@example.com", Status: iamv1.UserStatus_USER_STATUS_ACTIVE},
		},
	}, nil)
	env.IAMServer.ListUserRolesCalls.Returns(&iamv1.ListUserRolesResponse{
		Roles: []*iamv1.Role{
			{
				Id:   "role-id-1",
				Name: "admin",
				Permissions: []*iamv1.Permission{
					{Value: "read:clusters", Category: &cat},
					{Value: "write:clusters", Category: &cat},
				},
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "user", "describe", userID)
	require.NoError(t, err)

	assert.Contains(t, stdout, userID)
	assert.Contains(t, stdout, "alice@example.com")
	assert.Contains(t, stdout, "ACTIVE")
	assert.Contains(t, stdout, "role-id-1")
	assert.Contains(t, stdout, "admin")
	assert.Contains(t, stdout, "read:clusters")
	assert.Contains(t, stdout, "write:clusters")
	assert.Contains(t, stdout, "Cluster")

	req, ok := env.IAMServer.ListUserRolesCalls.Last()
	require.True(t, ok)
	assert.Equal(t, userID, req.GetUserId())
}

func TestUserDescribe_PermissionsDeduplicatedWithRoles(t *testing.T) {
	env := testutil.NewTestEnv(t)

	userID := testUserID
	cat := "Cluster"
	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{{Id: userID, Email: "alice@example.com"}},
	}, nil)
	env.IAMServer.ListUserRolesCalls.Returns(&iamv1.ListUserRolesResponse{
		Roles: []*iamv1.Role{
			{
				Id:   "role-id-1",
				Name: "admin",
				Permissions: []*iamv1.Permission{
					{Value: "read:clusters", Category: &cat},
					{Value: "write:clusters", Category: &cat},
				},
			},
			{
				Id:   "role-id-2",
				Name: "viewer",
				Permissions: []*iamv1.Permission{
					{Value: "read:clusters", Category: &cat},
				},
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "user", "describe", userID)
	require.NoError(t, err)

	// read:clusters appears in both roles — should be listed once with both role names
	assert.Contains(t, stdout, "admin, viewer")
	// write:clusters only in admin
	assert.Contains(t, stdout, "write:clusters")
}

func TestUserDescribe_NoPermissions(t *testing.T) {
	env := testutil.NewTestEnv(t)

	userID := testUserID
	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{
			{Id: userID, Email: "alice@example.com"},
		},
	}, nil)
	env.IAMServer.ListUserRolesCalls.Returns(&iamv1.ListUserRolesResponse{
		Roles: []*iamv1.Role{
			{Id: "role-id-1", Name: "viewer"},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "user", "describe", userID)
	require.NoError(t, err)

	assert.NotContains(t, stdout, "Effective Permissions")
}

func TestUserDescribe_ByEmail(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{
			{Id: "user-id-abc", Email: "alice@example.com"},
		},
	}, nil)
	env.IAMServer.ListUserRolesCalls.Returns(&iamv1.ListUserRolesResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "user", "describe", "alice@example.com")
	require.NoError(t, err)

	assert.Contains(t, stdout, "alice@example.com")
	req, ok := env.IAMServer.ListUserRolesCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "user-id-abc", req.GetUserId())
}

func TestUserDescribe_NotFound(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{Items: nil}, nil)

	_, _, err := testutil.Exec(t, env, "iam", "user", "describe", "nobody@example.com")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// --- user assign-role ---

func TestUserAssignRole_ByRoleID(t *testing.T) {
	env := testutil.NewTestEnv(t)

	userID := testUserID
	roleID := testRoleID

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{{Id: userID, Email: "alice@example.com"}},
	}, nil)
	env.IAMServer.AssignUserRolesCalls.Returns(&iamv1.AssignUserRolesResponse{}, nil)
	env.IAMServer.ListUserRolesCalls.Returns(&iamv1.ListUserRolesResponse{
		Roles: []*iamv1.Role{{Id: roleID, Name: "admin"}},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "user", "assign-role",
		"alice@example.com", roleID)
	require.NoError(t, err)
	assert.Contains(t, stdout, "alice@example.com")
	assert.Contains(t, stdout, roleID)
	assert.Contains(t, stdout, "admin")

	req, ok := env.IAMServer.AssignUserRolesCalls.Last()
	require.True(t, ok)
	assert.Equal(t, userID, req.GetUserId())
	assert.Equal(t, []string{roleID}, req.GetRoleIdsToAdd())
}

func TestUserAssignRole_ByRoleName(t *testing.T) {
	env := testutil.NewTestEnv(t)

	userID := testUserID
	roleID := testRoleID

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{{Id: userID, Email: "alice@example.com"}},
	}, nil)
	env.IAMServer.ListRolesCalls.Returns(&iamv1.ListRolesResponse{
		Items: []*iamv1.Role{{Id: roleID, Name: "admin"}},
	}, nil)
	env.IAMServer.AssignUserRolesCalls.Returns(&iamv1.AssignUserRolesResponse{}, nil)
	env.IAMServer.ListUserRolesCalls.Returns(&iamv1.ListUserRolesResponse{
		Roles: []*iamv1.Role{{Id: roleID, Name: "admin"}},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "user", "assign-role",
		"alice@example.com", "admin")
	require.NoError(t, err)
	assert.Contains(t, stdout, "admin")

	req, ok := env.IAMServer.AssignUserRolesCalls.Last()
	require.True(t, ok)
	assert.Equal(t, []string{roleID}, req.GetRoleIdsToAdd())
}

func TestUserAssignRole_MissingRole(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "iam", "user", "assign-role", "alice@example.com")
	require.Error(t, err)
}

// --- user remove-role ---

func TestUserRemoveRole_ByRoleID(t *testing.T) {
	env := testutil.NewTestEnv(t)

	userID := testUserID
	roleID := testRoleID

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{{Id: userID, Email: "alice@example.com"}},
	}, nil)
	env.IAMServer.AssignUserRolesCalls.Returns(&iamv1.AssignUserRolesResponse{}, nil)
	env.IAMServer.ListUserRolesCalls.Returns(&iamv1.ListUserRolesResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "iam", "user", "remove-role",
		"alice@example.com", roleID)
	require.NoError(t, err)
	assert.Contains(t, stdout, "alice@example.com")

	req, ok := env.IAMServer.AssignUserRolesCalls.Last()
	require.True(t, ok)
	assert.Equal(t, userID, req.GetUserId())
	assert.Equal(t, []string{roleID}, req.GetRoleIdsToDelete())
}

func TestUserRemoveRole_ByRoleName(t *testing.T) {
	env := testutil.NewTestEnv(t)

	userID := testUserID
	roleID := testRoleID

	env.IAMServer.ListUsersCalls.Returns(&iamv1.ListUsersResponse{
		Items: []*iamv1.User{{Id: userID, Email: "alice@example.com"}},
	}, nil)
	env.IAMServer.ListRolesCalls.Returns(&iamv1.ListRolesResponse{
		Items: []*iamv1.Role{{Id: roleID, Name: "viewer"}},
	}, nil)
	env.IAMServer.AssignUserRolesCalls.Returns(&iamv1.AssignUserRolesResponse{}, nil)
	env.IAMServer.ListUserRolesCalls.Returns(&iamv1.ListUserRolesResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "iam", "user", "remove-role",
		"alice@example.com", "viewer")
	require.NoError(t, err)

	req, ok := env.IAMServer.AssignUserRolesCalls.Last()
	require.True(t, ok)
	assert.Equal(t, []string{roleID}, req.GetRoleIdsToDelete())
}

func TestUserRemoveRole_MissingRole(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "iam", "user", "remove-role", "alice@example.com")
	require.Error(t, err)
}

// --- user invite ---

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

// --- invite list ---

func TestInviteList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.AccountServer.ListAccountInvitesCalls.Returns(&accountv1.ListAccountInvitesResponse{
		Items: []*accountv1.AccountInvite{
			{
				Id:        "invite-1",
				UserEmail: "alice@example.com",
				Status:    accountv1.AccountInviteStatus_ACCOUNT_INVITE_STATUS_PENDING,
				CreatedAt: timestamppb.New(time.Now().Add(-1 * time.Hour)),
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

// --- invite describe ---

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

// --- invite delete ---

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

