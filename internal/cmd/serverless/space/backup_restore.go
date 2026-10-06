package space

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

func newBackupRestoreCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Manage restores of serverless space backups",
		Long: `Manage restores of serverless space backups.

A restore writes the data of a backup back into the space it was taken from,
replacing the current data of the backed-up collections. To restore into a new
space instead, use "qcloud serverless space create-from-backup".`,
		Example: `# Restore a backup into its original space
qcloud serverless space backup restore trigger 9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d

# Track the progress of restores for a space
qcloud serverless space backup restore list --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(
		newBackupRestoreListCommand(s),
		newBackupRestoreTriggerCommand(s),
	)
	return cmd
}
