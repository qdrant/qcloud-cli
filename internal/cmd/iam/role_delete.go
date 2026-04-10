package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newRoleDeleteCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Long: `Delete a custom role from the account.

Only custom roles can be deleted. System roles are managed by Qdrant and cannot
be removed.`,
		Example: `# Delete a role (with confirmation prompt)
qcloud iam role delete 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Delete without confirmation
qcloud iam role delete 7b2ea926-724b-4de2-b73a-8675c42a6ebe --force`,
		ValidArgsFunction: completion.RoleIDCompletion(s),
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "delete <role-id>",
				Short: "Delete a custom role",
				Args:  util.ExactArgs(1, "a role ID"),
			}
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			force, _ := cmd.Flags().GetBool("force")

			if !util.ConfirmAction(force, cmd.ErrOrStderr(), fmt.Sprintf("Delete role %s?", args[0])) {
				fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
				return nil
			}

			client, err := s.Client(ctx)
			if err != nil {
				return err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return err
			}

			_, err = client.IAM().DeleteRole(ctx, &iamv1.DeleteRoleRequest{
				AccountId: accountID,
				RoleId:    args[0],
			})
			if err != nil {
				return fmt.Errorf("failed to delete role: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Role %s deleted.\n", args[0])
			return nil
		},
	}.CobraCommand(s)
}
