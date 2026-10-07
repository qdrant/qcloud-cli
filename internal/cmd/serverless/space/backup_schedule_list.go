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

func newBackupScheduleListCommand(s *state.State) *cobra.Command {
	cmd := base.ListCmd[*spacebackupv1.ListBackupSchedulesResponse]{
		Use:   "list",
		Short: "List backup schedules of serverless spaces",
		Long: `List backup schedules of serverless spaces in the current account.

The PAUSED column shows "scheduled" for schedules with a pause set in the future.
By default, all schedules are fetched automatically across multiple pages. Use
--page-size and --page-token for manual pagination.`,
		Example: `# List all backup schedules in the account
qcloud serverless space backup schedule list

# List the backup schedules of a space
qcloud serverless space backup schedule list --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60`,
		Fetch: func(s *state.State, cmd *cobra.Command) (*spacebackupv1.ListBackupSchedulesResponse, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			items, next, err := util.FetchPages(cmd, func(pageSize *int32, pageToken *string) ([]*spacebackupv1.BackupSchedule, string, error) {
				req := &spacebackupv1.ListBackupSchedulesRequest{
					AccountId: accountID,
					PageSize:  pageSize,
					PageToken: pageToken,
				}
				if cmd.Flags().Changed("space-id") {
					v, _ := cmd.Flags().GetString("space-id")
					req.SpaceId = &v
				}

				resp, err := client.ServerlessBackup().ListBackupSchedules(ctx, req)
				if err != nil {
					return nil, "", fmt.Errorf("failed to list backup schedules: %w", err)
				}

				return resp.GetItems(), resp.GetNextPageToken(), nil
			})
			if err != nil {
				return nil, err
			}

			return &spacebackupv1.ListBackupSchedulesResponse{Items: items, NextPageToken: next}, nil
		},
		OutputTable: func(_ *cobra.Command, w io.Writer, resp *spacebackupv1.ListBackupSchedulesResponse) (output.TableRenderer, error) {
			t := output.NewTable[*spacebackupv1.BackupSchedule](w)
			t.AddField("ID", func(v *spacebackupv1.BackupSchedule) string {
				return v.GetId()
			})
			t.AddField("NAME", func(v *spacebackupv1.BackupSchedule) string {
				return v.GetName()
			})
			t.AddField("SPACE", func(v *spacebackupv1.BackupSchedule) string {
				return v.GetSpaceId()
			})
			t.AddField("SCHEDULE", func(v *spacebackupv1.BackupSchedule) string {
				return v.GetSchedule()
			})
			t.AddField("COLLECTION", func(v *spacebackupv1.BackupSchedule) string {
				return formatCollection(v.CollectionName)
			})
			t.AddField("STATUS", func(v *spacebackupv1.BackupSchedule) string {
				return output.SpaceBackupScheduleStatus(v.GetStatus())
			})
			t.AddField("PAUSED", schedulePauseState)
			t.AddField("RETENTION", func(v *spacebackupv1.BackupSchedule) string {
				return output.RetentionPeriod(v.GetRetentionPeriod())
			})
			t.AddField("LAST RUN", func(v *spacebackupv1.BackupSchedule) string {
				if v.GetLastFiredAt() != nil {
					return output.HumanTime(v.GetLastFiredAt().AsTime())
				}

				return ""
			})
			t.SetItems(resp.GetItems())
			return t, nil
		},
	}.CobraCommand(s)

	util.AddPaginationFlags(cmd, "schedules")
	cmd.Flags().String("space-id", "", "Filter by space ID")
	_ = cmd.RegisterFlagCompletionFunc("space-id", completion.SpaceIDFlagCompletion(s))
	return cmd
}
