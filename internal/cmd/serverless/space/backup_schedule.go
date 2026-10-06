package space

import (
	"time"

	"github.com/spf13/cobra"

	spacebackupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/state"
)

func newBackupScheduleCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schedule",
		Short: "Manage backup schedules of serverless spaces",
		Long: `Manage backup schedules of serverless spaces.

A backup schedule creates backups of a space, or of a single collection, on a
cron schedule (in UTC). Every backup created by a schedule inherits its retention
period. A schedule can be paused and resumed without deleting it.`,
		Example: `# List the backup schedules of a space
qcloud serverless space backup schedule list --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Create a nightly schedule that keeps backups for 14 days
qcloud serverless space backup schedule create --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --name nightly --schedule "0 2 * * *" --retention-days 14

# Pause a schedule
qcloud serverless space backup schedule update 3f1c2b4a-8d7e-4f6a-9b0c-1d2e3f4a5b6c \
  --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --pause`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(
		newBackupScheduleListCommand(s),
		newBackupScheduleDescribeCommand(s),
		newBackupScheduleCreateCommand(s),
		newBackupScheduleUpdateCommand(s),
		newBackupScheduleDeleteCommand(s),
	)
	return cmd
}

// schedulePauseState reports whether a schedule is paused: "no", "yes", or
// "scheduled" when the pause takes effect in the future.
func schedulePauseState(sched *spacebackupv1.BackupSchedule) string {
	if sched.PausedAt == nil {
		return "no"
	}

	if sched.GetPausedAt().AsTime().After(time.Now()) {
		return "scheduled"
	}

	return "yes"
}
