package space

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	spacebackupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newBackupScheduleDescribeCommand(s *state.State) *cobra.Command {
	cmd := base.DescribeCmd[*spacebackupv1.BackupSchedule]{
		Use:   "describe <schedule-id>",
		Short: "Describe a backup schedule of a serverless space",
		Long: `Describe a backup schedule of a serverless space.

Shows the cron expression together with the next time it fires, the retention
period applied to created backups, and whether the schedule is paused. The
--space-id flag is required because the API looks up schedules within a space.`,
		Example: `# Describe a backup schedule
qcloud serverless space backup schedule describe 3f1c2b4a-8d7e-4f6a-9b0c-1d2e3f4a5b6c \
  --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60`,
		Args:              util.ExactArgs(1, "a schedule ID"),
		ValidArgsFunction: backupScheduleIDCompletion(s),
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*spacebackupv1.BackupSchedule, error) {
			return getBackupSchedule(s, cmd, args[0])
		},
		PrintText: func(_ *cobra.Command, w io.Writer, sched *spacebackupv1.BackupSchedule) error {
			fmt.Fprintf(w, "ID:          %s\n", sched.GetId())
			fmt.Fprintf(w, "Name:        %s\n", sched.GetName())
			fmt.Fprintf(w, "Space:       %s\n", sched.GetSpaceId())
			fmt.Fprintf(w, "Collection:  %s\n", formatCollection(sched.CollectionName))
			fmt.Fprintf(w, "Schedule:    %s\n", sched.GetSchedule())
			fmt.Fprintf(w, "Status:      %s\n", output.SpaceBackupScheduleStatus(sched.GetStatus()))

			paused := schedulePauseState(sched)
			if sched.PausedAt != nil {
				t := sched.GetPausedAt().AsTime()
				fmt.Fprintf(w, "Paused:      %s  (%s)\n", paused, output.FullDateTime(t))
			} else {
				fmt.Fprintf(w, "Paused:      %s\n", paused)
			}

			if paused != "yes" {
				if next, ok := util.NextCronRun(sched.GetSchedule()); ok {
					fmt.Fprintf(w, "Next Run:    %s  (%s)\n", output.HumanTime(next), output.FullDateTime(next))
				}
			}

			if sched.GetLastFiredAt() != nil {
				t := sched.GetLastFiredAt().AsTime()
				fmt.Fprintf(w, "Last Run:    %s  (%s)\n", output.HumanTime(t), output.FullDateTime(t))
			}

			fmt.Fprintf(w, "Retention:   %s\n", output.RetentionPeriod(sched.GetRetentionPeriod()))

			if sched.GetCreatedAt() != nil {
				t := sched.GetCreatedAt().AsTime()
				fmt.Fprintf(w, "Created:     %s  (%s)\n", output.HumanTime(t), output.FullDateTime(t))
			}

			return nil
		},
	}.CobraCommand(s)

	cmd.Flags().String("space-id", "", "ID of the space the schedule belongs to (required)")
	_ = cmd.MarkFlagRequired("space-id")
	_ = cmd.RegisterFlagCompletionFunc("space-id", completion.SpaceIDFlagCompletion(s))
	return cmd
}

// getBackupSchedule fetches a backup schedule by ID within the space given by --space-id.
func getBackupSchedule(s *state.State, cmd *cobra.Command, scheduleID string) (*spacebackupv1.BackupSchedule, error) {
	ctx := cmd.Context()
	client, err := s.Client(ctx)
	if err != nil {
		return nil, err
	}

	accountID, err := s.AccountID()
	if err != nil {
		return nil, err
	}

	spaceID, _ := cmd.Flags().GetString("space-id")

	resp, err := client.ServerlessBackup().GetBackupSchedule(ctx, &spacebackupv1.GetBackupScheduleRequest{
		AccountId:        accountID,
		SpaceId:          spaceID,
		BackupScheduleId: scheduleID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get backup schedule: %w", err)
	}

	return resp.GetBackupSchedule(), nil
}
