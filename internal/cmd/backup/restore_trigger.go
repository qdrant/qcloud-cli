package backup

import (
	"fmt"

	"github.com/spf13/cobra"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newRestoreTriggerCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "trigger <backup-id>",
				Short: "Trigger a restore from a backup",
				Args:  util.ExactArgs(1, "a backup ID"),
			}
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
			return cmd
		},
		ValidArgsFunction: completion.BackupIDCompletion(s),
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			backupID := args[0]

			force, _ := cmd.Flags().GetBool("force")
			if !util.ConfirmAction(force, fmt.Sprintf("Are you sure you want to restore backup %s?", backupID)) {
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

			_, err = client.Backup().RestoreBackup(ctx, &backupv1.RestoreBackupRequest{
				AccountId: accountID,
				BackupId:  backupID,
			})
			if err != nil {
				return fmt.Errorf("failed to restore backup: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Restore of backup %s started.\n", backupID)
			fmt.Fprintln(cmd.OutOrStdout(), "Run 'qcloud backup restore list' to track progress.")
			return nil
		},
	}.CobraCommand(s)
}
