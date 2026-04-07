package iam

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newUserAssignRoleCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		BaseCobraCommand: func() *cobra.Command {
			return &cobra.Command{
				Use:   "assign-role <user-id-or-email> <role> [<role>...]",
				Short: "Assign one or more roles to a user",
				Args:  cobra.MinimumNArgs(2),
			}
		},
		Long: `Assign one or more roles to a user in the account.

Accepts either a user ID (UUID) or an email address to identify the user.
Each role argument accepts either a role UUID or a role name, which is
resolved to an ID via the IAM service. Prints the user's resulting roles
after the assignment.`,
		Example: `# Assign a role by name
qcloud iam user assign-role user@example.com admin

# Assign a role by ID
qcloud iam user assign-role user@example.com 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Assign multiple roles at once
qcloud iam user assign-role user@example.com admin viewer`,
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
			return modifyUserRoles(s, cmd, client, accountID, user, roleIDs, nil, "assign")
		},
	}.CobraCommand(s)
}
