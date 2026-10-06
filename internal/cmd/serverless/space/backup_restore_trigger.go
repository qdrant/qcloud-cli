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

func newBackupRestoreTriggerCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Long: `Restore a backup into its original serverless space.

The restore runs asynchronously and replaces the current data of the backed-up
collections in the space. Use "qcloud serverless space backup restore list" to
follow its progress.`,
		Example: `# Restore a backup (prompts for confirmation)
qcloud serverless space backup restore trigger 9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d

# Restore a backup without confirmation
qcloud serverless space backup restore trigger 9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d --force`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "trigger <backup-id>",
				Short: "Restore a backup into its original space",
				Args:  util.ExactArgs(1, "a backup ID"),
			}
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
			return cmd
		},
		ValidArgsFunction: completion.SpaceBackupIDCompletion(s),
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			backupID := args[0]

			force, _ := cmd.Flags().GetBool("force")
			prompt := fmt.Sprintf("Are you sure you want to restore backup %s? This replaces the current data in the space.", backupID)
			if !util.ConfirmAction(force, cmd.ErrOrStderr(), prompt) {
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

			_, err = client.ServerlessBackup().RestoreBackup(ctx, &spacebackupv1.RestoreBackupRequest{
				AccountId: accountID,
				BackupId:  backupID,
			})
			if err != nil {
				return fmt.Errorf("failed to restore backup: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Restore of backup %s started.\n", backupID)
			fmt.Fprintln(cmd.OutOrStdout(), "Run 'qcloud serverless space backup restore list' to track progress.")
			return nil
		},
	}.CobraCommand(s)
}
