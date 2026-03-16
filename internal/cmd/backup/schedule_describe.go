package backup

import (
	"fmt"
	"io"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func nextScheduleRun(cronExpr string) (time.Time, bool) {
	s, err := cron.ParseStandard(cronExpr)
	if err != nil {
		return time.Time{}, false
	}
	return s.Next(time.Now().UTC()), true
}

func newScheduleDescribeCommand(s *state.State) *cobra.Command {
	cmd := base.DescribeCmd[*backupv1.BackupSchedule]{
		Use:               "describe <schedule-id>",
		Short:             "Describe a backup schedule",
		Args:              util.ExactArgs(1, "a schedule ID"),
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
		PrintText: func(_ *cobra.Command, w io.Writer, sched *backupv1.BackupSchedule) error {
			fmt.Fprintf(w, "ID:        %s\n", sched.GetId())
			fmt.Fprintf(w, "Cluster:   %s\n", sched.GetClusterId())
			fmt.Fprintf(w, "Schedule:  %s\n", sched.GetSchedule())
			if next, ok := nextScheduleRun(sched.GetSchedule()); ok {
				fmt.Fprintf(w, "Next Run:  %s  (%s)\n", output.HumanTime(next), output.FullDateTime(next))
			}
			fmt.Fprintf(w, "Status:    %s\n", scheduleStatusString(sched.GetStatus()))
			if sched.GetCreatedAt() != nil {
				t := sched.GetCreatedAt().AsTime()
				fmt.Fprintf(w, "Created:   %s  (%s)\n", output.HumanTime(t), output.FullDateTime(t))
			}
			if sched.GetRetentionPeriod() != nil {
				days := int64(sched.GetRetentionPeriod().AsDuration().Hours()) / 24
				fmt.Fprintf(w, "Retention: %d days\n", days)
			}
			return nil
		},
	}.CobraCommand(s)

	cmd.Flags().String("cluster-id", "", "Cluster ID (required)")
	_ = cmd.MarkFlagRequired("cluster-id")
	_ = cmd.RegisterFlagCompletionFunc("cluster-id", completion.ClusterIDCompletion(s))
	return cmd
}
