package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newUserDeleteCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Long: `Delete a user from Qdrant Cloud.

Accepts either a user ID (UUID) or an email address to identify the user.
Deleting a user is permanent and cannot be undone. Deletion fails if the user
still owns any accounts; ownership of those accounts must be transferred first.

A confirmation prompt is shown unless --force is passed.`,
		Example: `# Delete a user by email (with confirmation prompt)
qcloud iam user delete user@example.com

# Delete a user by ID
qcloud iam user delete 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Delete without confirmation
qcloud iam user delete user@example.com --force`,
		ValidArgsFunction: userCompletion(s),
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "delete <user-id-or-email>",
				Short: "Delete a user",
				Args:  util.ExactArgs(1, "a user ID or email"),
			}
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			force, _ := cmd.Flags().GetBool("force")

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

			if !util.ConfirmAction(force, cmd.ErrOrStderr(), fmt.Sprintf("Delete user %s (%s)?", user.GetEmail(), user.GetId())) {
				fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
				return nil
			}

			_, err = client.IAM().DeleteUser(ctx, &iamv1.DeleteUserRequest{
				UserId: user.GetId(),
			})
			if err != nil {
				return fmt.Errorf("failed to delete user: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "User %s deleted.\n", user.GetEmail())
			return nil
		},
	}.CobraCommand(s)
}
