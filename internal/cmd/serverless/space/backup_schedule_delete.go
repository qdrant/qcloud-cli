package space

import (
	"fmt"

	"github.com/spf13/cobra"

	spacebackupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newBackupScheduleDeleteCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Long: `Delete a backup schedule of a serverless space.

The schedule stops creating backups. Backups it already created are kept by
default; pass --delete-backups to remove them as well.`,
		Example: `# Delete a schedule (prompts for confirmation)
qcloud serverless space backup schedule delete 3f1c2b4a-8d7e-4f6a-9b0c-1d2e3f4a5b6c

# Delete a schedule and all backups it created, without confirmation
qcloud serverless space backup schedule delete 3f1c2b4a-8d7e-4f6a-9b0c-1d2e3f4a5b6c --delete-backups --force`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "delete <schedule-id>",
				Short: "Delete a backup schedule of a serverless space",
				Args:  util.ExactArgs(1, "a schedule ID"),
			}
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
			cmd.Flags().Bool("delete-backups", false, "Also delete all backups created by this schedule")
			return cmd
		},
		ValidArgsFunction: backupScheduleIDCompletion(s),
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			scheduleID := args[0]

			deleteBackups, _ := cmd.Flags().GetBool("delete-backups")
			prompt := fmt.Sprintf("Are you sure you want to delete backup schedule %s?", scheduleID)
			if deleteBackups {
				prompt = fmt.Sprintf("Are you sure you want to delete backup schedule %s and all backups it created?", scheduleID)
			}

			force, _ := cmd.Flags().GetBool("force")
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

			req := &spacebackupv1.DeleteBackupScheduleRequest{
				AccountId:        accountID,
				BackupScheduleId: scheduleID,
			}
			if cmd.Flags().Changed("delete-backups") {
				req.DeleteBackups = &deleteBackups
			}

			if _, err := client.ServerlessBackup().DeleteBackupSchedule(ctx, req); err != nil {
				return fmt.Errorf("failed to delete backup schedule: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Backup schedule %s deleted.\n", scheduleID)
			return nil
		},
	}.CobraCommand(s)
}
