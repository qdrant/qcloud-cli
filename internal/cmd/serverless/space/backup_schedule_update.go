package space

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"

	spacebackupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

// pauseClockSkew is how far "--pause" back-dates the pause timestamp to
// tolerate a client clock that is behind the server's.
const pauseClockSkew = time.Minute

func newBackupScheduleUpdateCommand(s *state.State) *cobra.Command {
	cmd := base.UpdateCmd[*spacebackupv1.BackupSchedule]{
		Long: `Update a backup schedule of a serverless space.

Only the fields whose flags are given are changed. --pause stops the schedule from
creating new backups immediately and --resume starts it again; the schedule and
its existing backups are kept while paused. Pausing an already paused schedule
keeps its original pause time, while a pause scheduled for the future is brought
forward to now. The --space-id flag is required
because the API looks up schedules within a space.`,
		Example: `# Change the schedule to run every 6 hours
qcloud serverless space backup schedule update 3f1c2b4a-8d7e-4f6a-9b0c-1d2e3f4a5b6c \
  --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --schedule "0 */6 * * *"

# Pause a schedule
qcloud serverless space backup schedule update 3f1c2b4a-8d7e-4f6a-9b0c-1d2e3f4a5b6c \
  --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --pause

# Resume a schedule and change its retention
qcloud serverless space backup schedule update 3f1c2b4a-8d7e-4f6a-9b0c-1d2e3f4a5b6c \
  --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --resume --retention-days 30`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "update <schedule-id>",
				Short: "Update a backup schedule of a serverless space",
				Args:  util.ExactArgs(1, "a schedule ID"),
			}
			cmd.Flags().String("space-id", "", "ID of the space the schedule belongs to (required)")
			cmd.Flags().String("name", "", "New name of the schedule")
			cmd.Flags().String("schedule", "", "New cron schedule expression in UTC, e.g. '0 2 * * *'")
			cmd.Flags().Uint32("retention-days", 0, "New retention period of created backups in days (1-365)")
			cmd.Flags().Bool("pause", false, "Pause the schedule")
			cmd.Flags().Bool("resume", false, "Resume a paused schedule")
			_ = cmd.MarkFlagRequired("space-id")
			cmd.MarkFlagsMutuallyExclusive("pause", "resume")
			return cmd
		},
		ValidArgsFunction: backupScheduleIDCompletion(s),
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*spacebackupv1.BackupSchedule, error) {
			return getBackupSchedule(s, cmd, args[0])
		},
		Update: func(s *state.State, cmd *cobra.Command, sched *spacebackupv1.BackupSchedule) (*spacebackupv1.BackupSchedule, error) {
			if cmd.Flags().Changed("name") {
				name, _ := cmd.Flags().GetString("name")
				sched.Name = name
			}

			if cmd.Flags().Changed("schedule") {
				schedule, _ := cmd.Flags().GetString("schedule")
				sched.Schedule = schedule
			}

			if cmd.Flags().Changed("retention-days") {
				days, _ := cmd.Flags().GetUint32("retention-days")
				retention, err := retentionFromDays(days)
				if err != nil {
					return nil, err
				}

				sched.RetentionPeriod = retention
			}

			if pause, _ := cmd.Flags().GetBool("pause"); pause {
				now := time.Now()
				// Keep an existing pause that is already in effect so the
				// original pause time is preserved. A pause scheduled for the
				// future is brought forward to now.
				if sched.PausedAt == nil || sched.GetPausedAt().AsTime().After(now) {
					// Back-date slightly so the pause is already in effect on
					// the server even if the local clock runs behind it.
					sched.PausedAt = timestamppb.New(now.Add(-pauseClockSkew))
				}
			}

			if resume, _ := cmd.Flags().GetBool("resume"); resume {
				sched.PausedAt = nil
			}

			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			resp, err := client.ServerlessBackup().UpdateBackupSchedule(ctx, &spacebackupv1.UpdateBackupScheduleRequest{
				BackupSchedule: sched,
				UpdateMask:     updateMask(cmd, backupScheduleUpdatePaths),
			})
			if err != nil {
				return nil, fmt.Errorf("failed to update backup schedule: %w", err)
			}

			return resp.GetBackupSchedule(), nil
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, sched *spacebackupv1.BackupSchedule) {
			fmt.Fprintf(out, "Backup schedule %s (%s) updated.\n", sched.GetId(), sched.GetName())
		},
	}.CobraCommand(s)
	_ = cmd.RegisterFlagCompletionFunc("space-id", completion.SpaceIDFlagCompletion(s))
	return cmd
}
