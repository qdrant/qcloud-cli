package iam

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/qcloudapi"
	"github.com/qdrant/qcloud-cli/internal/state"
)

// resolveUser looks up a user by UUID or email from the account's user list.
func resolveUser(cmd *cobra.Command, client *qcloudapi.Client, accountID, idOrEmail string) (*iamv1.User, error) {
	ctx := cmd.Context()
	resp, err := client.IAM().ListUsers(ctx, &iamv1.ListUsersRequest{AccountId: accountID})
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	for _, u := range resp.GetItems() {
		if util.IsUUID(idOrEmail) {
			if u.GetId() == idOrEmail {
				return u, nil
			}
		} else {
			if u.GetEmail() == idOrEmail {
				return u, nil
			}
		}
	}
	return nil, fmt.Errorf("user %s not found", idOrEmail)
}

// resolveRoleIDs converts a slice of role names or UUIDs to UUIDs.
// Values that already look like UUIDs are passed through unchanged.
// Non-UUID values are resolved by name via ListRoles.
func resolveRoleIDs(ctx context.Context, client *qcloudapi.Client, accountID string, namesOrIDs []string) ([]string, error) {
	if len(namesOrIDs) == 0 {
		return nil, nil
	}

	// Check whether any name resolution is needed.
	var needsLookup bool
	for _, v := range namesOrIDs {
		if !util.IsUUID(v) {
			needsLookup = true
			break
		}
	}

	var rolesByName map[string]string
	if needsLookup {
		resp, err := client.IAM().ListRoles(ctx, &iamv1.ListRolesRequest{AccountId: accountID})
		if err != nil {
			return nil, fmt.Errorf("failed to list roles: %w", err)
		}
		rolesByName = make(map[string]string, len(resp.GetItems()))
		for _, r := range resp.GetItems() {
			rolesByName[r.GetName()] = r.GetId()
		}
	}

	ids := make([]string, 0, len(namesOrIDs))
	for _, v := range namesOrIDs {
		if util.IsUUID(v) {
			ids = append(ids, v)
		} else {
			id, ok := rolesByName[v]
			if !ok {
				return nil, fmt.Errorf("role %q not found", v)
			}
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// modifyUserRoles calls AssignUserRoles with the given add/delete IDs, then
// fetches and prints the resulting role list.
func modifyUserRoles(s *state.State, cmd *cobra.Command, client *qcloudapi.Client, accountID string, user *iamv1.User, addIDs, removeIDs []string) error {
	ctx := cmd.Context()

	_, err := client.IAM().AssignUserRoles(ctx, &iamv1.AssignUserRolesRequest{
		AccountId:       accountID,
		UserId:          user.GetId(),
		RoleIdsToAdd:    addIDs,
		RoleIdsToDelete: removeIDs,
	})
	if err != nil {
		return fmt.Errorf("failed to modify roles: %w", err)
	}

	rolesResp, err := client.IAM().ListUserRoles(ctx, &iamv1.ListUserRolesRequest{
		AccountId: accountID,
		UserId:    user.GetId(),
	})
	if err != nil {
		return fmt.Errorf("failed to list user roles: %w", err)
	}

	if s.Config.JSONOutput() {
		return output.PrintJSON(cmd.OutOrStdout(), rolesResp)
	}

	w := cmd.OutOrStdout()
	fmt.Fprintf(w, "Roles for %s:\n", user.GetEmail())
	printRoles(w, rolesResp.GetRoles())
	return nil
}
