package output

import (
	"strings"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"
)

// BackupStatus returns a concise label for a BackupStatus.
func BackupStatus(s backupv1.BackupStatus) string {
	return strings.TrimPrefix(s.String(), "BACKUP_STATUS_")
}

// BackupScheduleStatus returns a concise label for a BackupScheduleStatus.
func BackupScheduleStatus(s backupv1.BackupScheduleStatus) string {
	return strings.TrimPrefix(s.String(), "BACKUP_SCHEDULE_STATUS_")
}

// BackupRestoreStatus returns a concise label for a BackupRestoreStatus.
func BackupRestoreStatus(s backupv1.BackupRestoreStatus) string {
	return strings.TrimPrefix(s.String(), "BACKUP_RESTORE_STATUS_")
}
