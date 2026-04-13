package backup

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newScheduleListCommand(s *state.State) *cobra.Command {
	cmd := base.ListCmd[*backupv1.ListBackupSchedulesResponse]{
		Use:   "list",
		Short: "List backup schedules",
		Fetch: func(s *state.State, cmd *cobra.Command) (*backupv1.ListBackupSchedulesResponse, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			req := &backupv1.ListBackupSchedulesRequest{AccountId: accountID}
			if cmd.Flags().Changed("cluster-id") {
				clusterID, _ := cmd.Flags().GetString("cluster-id")
				req.ClusterId = &clusterID
			}

			resp, err := client.Backup().ListBackupSchedules(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("failed to list backup schedules: %w", err)
			}
			return resp, nil
		},
		OutputTable: func(_ *cobra.Command, w io.Writer, resp *backupv1.ListBackupSchedulesResponse) (output.TableRenderer, error) {
			t := output.NewTable[*backupv1.BackupSchedule](w)
			t.AddField("ID", func(v *backupv1.BackupSchedule) string {
				return v.GetId()
			})
			t.AddField("CLUSTER", func(v *backupv1.BackupSchedule) string {
				return v.GetClusterId()
			})
			t.AddField("SCHEDULE", func(v *backupv1.BackupSchedule) string {
				return v.GetSchedule()
			})
			t.AddField("STATUS", func(v *backupv1.BackupSchedule) string {
				return output.BackupScheduleStatus(v.GetStatus())
			})
			t.AddField("NEXT RUN", func(v *backupv1.BackupSchedule) string {
				if next, ok := nextScheduleRun(v.GetSchedule()); ok {
					return output.HumanTime(next)
				}
				return ""
			})
			t.AddField("CREATED", func(v *backupv1.BackupSchedule) string {
				if v.GetCreatedAt() != nil {
					return output.HumanTime(v.GetCreatedAt().AsTime())
				}
				return ""
			})
			t.SetItems(resp.GetItems())
			return t, nil
		},
	}.CobraCommand(s)

	cmd.Flags().String("cluster-id", "", "Filter by cluster ID")
	_ = cmd.RegisterFlagCompletionFunc("cluster-id", completion.ClusterIDCompletion(s))
	return cmd
}
