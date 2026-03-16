package backup

import (
	"strings"

	"github.com/spf13/cobra"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// NewCommand creates the "backup" parent command and registers all subcommands.
func NewCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Manage Qdrant Cloud backups",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(
		newListCommand(s),
		newDescribeCommand(s),
		newCreateCommand(s),
		newDeleteCommand(s),
		newRestoreCommand(s),
		newScheduleCommand(s),
	)
	return cmd
}

// backupStatusString returns a concise status label for a BackupStatus.
func backupStatusString(s backupv1.BackupStatus) string {
	return strings.TrimPrefix(s.String(), "BACKUP_STATUS_")
}

// scheduleStatusString returns a concise status label for a BackupScheduleStatus.
func scheduleStatusString(s backupv1.BackupScheduleStatus) string {
	return strings.TrimPrefix(s.String(), "BACKUP_SCHEDULE_STATUS_")
}

// restoreStatusString returns a concise status label for a BackupRestoreStatus.
func restoreStatusString(s backupv1.BackupRestoreStatus) string {
	return strings.TrimPrefix(s.String(), "BACKUP_RESTORE_STATUS_")
}
