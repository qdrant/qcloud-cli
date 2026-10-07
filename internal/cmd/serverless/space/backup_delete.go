package space

import (
	"fmt"

	"github.com/spf13/cobra"

	spacebackupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newBackupDeleteCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Long: `Delete a backup of a serverless space.

Deletion cannot be undone; the backup can no longer be restored or used to
create a new space.`,
		Example: `# Delete a backup (prompts for confirmation)
qcloud serverless space backup delete 9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d

# Delete a backup without confirmation
qcloud serverless space backup delete 9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d --force`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "delete <backup-id>",
				Short: "Delete a backup of a serverless space",
				Args:  util.ExactArgs(1, "a backup ID"),
			}
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
			return cmd
		},
		ValidArgsFunction: completion.SpaceBackupIDCompletion(s),
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			backupID := args[0]

			force, _ := cmd.Flags().GetBool("force")
			if !util.ConfirmAction(force, cmd.ErrOrStderr(), fmt.Sprintf("Are you sure you want to delete backup %s?", backupID)) {
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

			_, err = client.ServerlessBackup().DeleteBackup(ctx, &spacebackupv1.DeleteBackupRequest{
				AccountId: accountID,
				BackupId:  backupID,
			})
			if err != nil {
				return fmt.Errorf("failed to delete backup: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Backup %s deleted.\n", backupID)
			return nil
		},
	}.CobraCommand(s)
}
