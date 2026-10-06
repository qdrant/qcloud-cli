package space

import (
	"fmt"

	"github.com/spf13/cobra"

	spacev1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newDeleteCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Long: `Delete a serverless space and all of its collections.

Deletion cannot be undone. Backups of the space are kept by default so that the
data can still be restored into a new space with "create-from-backup"; pass
--delete-backups to remove them as well.`,
		Example: `# Delete a space (prompts for confirmation)
qcloud serverless space delete 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Delete a space and its backups without confirmation
qcloud serverless space delete 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --delete-backups --force`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "delete <space-id>",
				Short: "Delete a space",
				Args:  util.ExactArgs(1, "a space ID"),
			}
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
			cmd.Flags().Bool("delete-backups", false, "Also delete all backups of the space")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			spaceID := args[0]

			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return err
			}

			deleteBackups, _ := cmd.Flags().GetBool("delete-backups")
			prompt := fmt.Sprintf("Are you sure you want to delete space %s?", spaceID)
			if deleteBackups {
				prompt = fmt.Sprintf("Are you sure you want to delete space %s and all of its backups?", spaceID)
			}

			force, _ := cmd.Flags().GetBool("force")
			if !util.ConfirmAction(force, cmd.ErrOrStderr(), prompt) {
				fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
				return nil
			}

			req := &spacev1.DeleteSpaceRequest{
				AccountId: accountID,
				SpaceId:   spaceID,
			}
			if cmd.Flags().Changed("delete-backups") {
				req.DeleteBackups = &deleteBackups
			}

			if _, err := client.ServerlessSpace().DeleteSpace(ctx, req); err != nil {
				return fmt.Errorf("failed to delete space: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Space %s deleted.\n", spaceID)
			return nil
		},
		ValidArgsFunction: completion.SpaceIDCompletion(s),
	}.CobraCommand(s)
}
