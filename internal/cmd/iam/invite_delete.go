package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newInviteDeleteCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "delete <invite-id>",
				Short: "Delete an account invite",
				Args:  util.ExactArgs(1, "an invite ID"),
			}
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
			return cmd
		},
		Long: `Delete an account invite.

Cancels a pending account invite. The invited user will no longer be able to
accept or reject the invite. Requires the delete:invites permission.`,
		Example: `# Delete an invite
qcloud iam invite delete 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Delete without confirmation
qcloud iam invite delete 7b2ea926-724b-4de2-b73a-8675c42a6ebe --force`,
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			force, _ := cmd.Flags().GetBool("force")
			if !util.ConfirmAction(force, cmd.ErrOrStderr(),
				fmt.Sprintf("Delete invite %s?", args[0])) {
				fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
				return nil
			}

			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return err
			}
			accountID, err := s.AccountID()
			if err != nil {
				return err
			}

			_, err = client.Account().DeleteAccountInvite(ctx, &accountv1.DeleteAccountInviteRequest{
				AccountId: accountID,
				InviteId:  args[0],
			})
			if err != nil {
				return fmt.Errorf("failed to delete invite: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Invite %s deleted.\n", args[0])
			return nil
		},
	}.CobraCommand(s)
}
