package iam

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newUserAssignRoleCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "assign-role <user-id-or-email>",
				Short: "Assign one or more roles to a user",
				Args:  util.ExactArgs(1, "a user ID or email"),
			}

			_ = cmd.Flags().StringSliceP("role", "r", nil, "A role ID or name")
			_ = cmd.RegisterFlagCompletionFunc("role", completion.RoleCompletion(s))
			return cmd
		},
		ValidArgsFunction: userCompletion(s),
		Long: `Assign one or more roles to a user in the account.

Accepts either a user ID (UUID) or an email address to identify the user.
Each role accepts either a role UUID or a role name, which is
resolved to an ID via the IAM service. Prints the user's resulting roles
after the assignment.`,
		Example: `# Assign a role by name
qcloud iam user assign-role user@example.com --role admin

# Assign a role by ID
qcloud iam user assign-role user@example.com --role 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Assign multiple roles at once
qcloud iam user assign-role user@example.com --role admin --role viewer

# Assign multiple roles at once using comma separated values
qcloud iam user assign-role user@example.com --role admin,viewer`,
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

			roles, _ := cmd.Flags().GetStringSlice("role")
			roleIDs, err := resolveRoleIDs(ctx, client, accountID, roles)
			if err != nil {
				return err
			}

			return modifyUserRoles(s, cmd, client, accountID, user, roleIDs, nil)
		},
	}.CobraCommand(s)
}
