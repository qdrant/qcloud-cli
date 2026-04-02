package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newUserRemoveRoleCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		BaseCobraCommand: func() *cobra.Command {
			return &cobra.Command{
				Use:   "remove-role <user-id-or-email> <role> [<role>...]",
				Short: "Remove one or more roles from a user",
				Args:  cobra.MinimumNArgs(2),
			}
		},
		Long: `Remove one or more roles from a user in the account.

Accepts either a user ID (UUID) or an email address to identify the user.
Each role argument accepts either a role UUID or a role name, which is
resolved to an ID via the IAM service. Prints the user's resulting roles
after the removal.`,
		Example: `# Remove a role by name
qcloud iam user remove-role user@example.com admin

# Remove a role by ID
qcloud iam user remove-role user@example.com 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Remove multiple roles at once
qcloud iam user remove-role user@example.com admin viewer`,
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return err
			}
			accountID, err := s.AccountID()
			if err != nil {
				return err
			}

			user, err := resolveUser(cmd, client, accountID, args[0])
			if err != nil {
				return err
			}

			roleIDs, err := resolveRoleIDs(ctx, client, accountID, args[1:])
			if err != nil {
				return err
			}

			_, err = client.IAM().AssignUserRoles(ctx, &iamv1.AssignUserRolesRequest{
				AccountId:       accountID,
				UserId:          user.GetId(),
				RoleIdsToDelete: roleIDs,
			})
			if err != nil {
				return fmt.Errorf("failed to remove roles: %w", err)
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
		},
	}.CobraCommand(s)
}
