package backup

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/durationpb"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newScheduleUpdateCommand(s *state.State) *cobra.Command {
	cmd := base.UpdateCmd[*backupv1.BackupSchedule]{
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "update <schedule-id>",
				Short: "Update a backup schedule",
				Args:  util.ExactArgs(1, "a schedule ID"),
			}
			cmd.Flags().String("cluster-id", "", "Cluster ID (required)")
			cmd.Flags().String("schedule", "", "New cron schedule expression in UTC, e.g. '0 2 * * *'")
			cmd.Flags().Uint32("retention-days", 0, "New retention period in days (1-365)")
			_ = cmd.MarkFlagRequired("cluster-id")
			return cmd
		},
		ValidArgsFunction: scheduleIDCompletion(s),
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*backupv1.BackupSchedule, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			clusterID, _ := cmd.Flags().GetString("cluster-id")

			resp, err := client.Backup().GetBackupSchedule(ctx, &backupv1.GetBackupScheduleRequest{
				AccountId:        accountID,
				ClusterId:        clusterID,
				BackupScheduleId: args[0],
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get backup schedule: %w", err)
			}
			return resp.GetBackupSchedule(), nil
		},
		Update: func(s *state.State, cmd *cobra.Command, sched *backupv1.BackupSchedule) (*backupv1.BackupSchedule, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			if cmd.Flags().Changed("schedule") {
				schedule, _ := cmd.Flags().GetString("schedule")
				sched.Schedule = schedule
			}

			if cmd.Flags().Changed("retention-days") {
				retentionDays, _ := cmd.Flags().GetUint32("retention-days")
				if retentionDays < 1 {
					return nil, fmt.Errorf("--retention-days must be at least 1")
				}
				d := time.Duration(retentionDays) * 24 * time.Hour
				sched.RetentionPeriod = durationpb.New(d)
			}

			resp, err := client.Backup().UpdateBackupSchedule(ctx, &backupv1.UpdateBackupScheduleRequest{
				BackupSchedule: sched,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to update backup schedule: %w", err)
			}
			return resp.GetBackupSchedule(), nil
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, sched *backupv1.BackupSchedule) {
			fmt.Fprintf(out, "Backup schedule %s updated.\n", sched.GetId())
		},
	}.CobraCommand(s)
	_ = cmd.RegisterFlagCompletionFunc("cluster-id", completion.ClusterIDCompletion(s))
	return cmd
}
