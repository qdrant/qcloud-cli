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
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newScheduleCreateCommand(s *state.State) *cobra.Command {
	cmd := base.CreateCmd[*backupv1.BackupSchedule]{
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "create",
				Short: "Create a backup schedule for a cluster",
				Args:  cobra.NoArgs,
			}
			cmd.Flags().String("cluster-id", "", "Cluster ID (required)")
			cmd.Flags().String("schedule", "", "Cron schedule expression in UTC (required), e.g. '0 2 * * *'")
			cmd.Flags().Uint32("retention-days", 0, "Retention period in days (1-365) (required)")
			_ = cmd.MarkFlagRequired("cluster-id")
			_ = cmd.MarkFlagRequired("schedule")
			_ = cmd.MarkFlagRequired("retention-days")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) (*backupv1.BackupSchedule, error) {
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
			schedule, _ := cmd.Flags().GetString("schedule")
			retentionDays, _ := cmd.Flags().GetUint32("retention-days")

			if retentionDays < 1 {
				return nil, fmt.Errorf("--retention-days must be at least 1")
			}

			d := time.Duration(retentionDays) * 24 * time.Hour
			sched := &backupv1.BackupSchedule{
				AccountId:       accountID,
				ClusterId:       clusterID,
				Schedule:        schedule,
				RetentionPeriod: durationpb.New(d),
			}

			resp, err := client.Backup().CreateBackupSchedule(ctx, &backupv1.CreateBackupScheduleRequest{
				BackupSchedule: sched,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create backup schedule: %w", err)
			}
			return resp.GetBackupSchedule(), nil
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, sched *backupv1.BackupSchedule) {
			fmt.Fprintf(out, "Backup schedule %s created for cluster %s.\n", sched.GetId(), sched.GetClusterId())
		},
	}.CobraCommand(s)
	_ = cmd.RegisterFlagCompletionFunc("cluster-id", completion.ClusterIDCompletion(s))
	return cmd
}
