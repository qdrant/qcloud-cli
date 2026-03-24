package backup

import (
	"fmt"

	"github.com/spf13/cobra"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newScheduleDeleteCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "delete <schedule-id>",
				Short: "Delete a backup schedule",
				Args:  util.ExactArgs(1, "a schedule ID"),
			}
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
			cmd.Flags().Bool("delete-backups", false, "Also delete all backups created by this schedule")
			return cmd
		},
		ValidArgsFunction: scheduleIDCompletion(s),
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			scheduleID := args[0]

			force, _ := cmd.Flags().GetBool("force")
			if !util.ConfirmAction(force, cmd.ErrOrStderr(), fmt.Sprintf("Are you sure you want to delete backup schedule %s?", scheduleID)) {
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

			deleteBackups, _ := cmd.Flags().GetBool("delete-backups")

			req := &backupv1.DeleteBackupScheduleRequest{
				AccountId:        accountID,
				BackupScheduleId: scheduleID,
				DeleteBackups:    &deleteBackups,
			}

			_, err = client.Backup().DeleteBackupSchedule(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to delete backup schedule: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Backup schedule %s deleted.\n", scheduleID)
			return nil
		},
	}.CobraCommand(s)
}
