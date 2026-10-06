package space

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

func newBackupCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Manage backups of serverless spaces",
		Long: `Manage backups of serverless spaces.

A backup captures either a whole space or a single collection of a space. Backups
are created on demand or by a backup schedule, and are kept for their retention
period (indefinitely when no retention is set). A backup can be restored in place
into its original space, or used to create a new space with
"qcloud serverless space create-from-backup".`,
		Example: `# List all backups of a space
qcloud serverless space backup list --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Back up a space and keep the backup for 30 days
qcloud serverless space backup create --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --retention-days 30

# Create a daily backup schedule
qcloud serverless space backup schedule create --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --name nightly --schedule "0 2 * * *"`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(
		newBackupListCommand(s),
		newBackupDescribeCommand(s),
		newBackupCreateCommand(s),
		newBackupDeleteCommand(s),
		newBackupRestoreCommand(s),
		newBackupScheduleCommand(s),
	)
	return cmd
}
