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

func newBackupListCommand(s *state.State) *cobra.Command {
	cmd := base.ListCmd[*spacebackupv1.ListBackupsResponse]{
		Use:   "list",
		Short: "List backups of serverless spaces",
		Long: `List backups of serverless spaces in the current account.

Backups can be filtered by space, by the schedule that created them, and by
collection. The COLLECTION column shows "(all)" for backups of a whole space.

By default, all backups are fetched automatically across multiple pages. Use
--page-size and --page-token for manual pagination; the next page token is
included in the JSON output when more pages exist.`,
		Example: `# List all backups in the account
qcloud serverless space backup list

# List backups of a space
qcloud serverless space backup list --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# List backups of a single collection created by a schedule
qcloud serverless space backup list --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --schedule-id 3f1c2b4a-8d7e-4f6a-9b0c-1d2e3f4a5b6c --collection products`,
		Fetch: func(s *state.State, cmd *cobra.Command) (*spacebackupv1.ListBackupsResponse, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			items, next, err := util.FetchPages(cmd, func(pageSize *int32, pageToken *string) ([]*spacebackupv1.Backup, string, error) {
				req := &spacebackupv1.ListBackupsRequest{
					AccountId: accountID,
					PageSize:  pageSize,
					PageToken: pageToken,
				}
				if cmd.Flags().Changed("space-id") {
					v, _ := cmd.Flags().GetString("space-id")
					req.SpaceId = &v
				}

				if cmd.Flags().Changed("schedule-id") {
					v, _ := cmd.Flags().GetString("schedule-id")
					req.BackupScheduleId = &v
				}

				if cmd.Flags().Changed("collection") {
					v, _ := cmd.Flags().GetString("collection")
					req.CollectionName = &v
				}

				resp, err := client.ServerlessBackup().ListBackups(ctx, req)
				if err != nil {
					return nil, "", fmt.Errorf("failed to list backups: %w", err)
				}

				return resp.GetItems(), resp.GetNextPageToken(), nil
			})
			if err != nil {
				return nil, err
			}

			return &spacebackupv1.ListBackupsResponse{Items: items, NextPageToken: next}, nil
		},
		OutputTable: func(_ *cobra.Command, w io.Writer, resp *spacebackupv1.ListBackupsResponse) (output.TableRenderer, error) {
			t := output.NewTable[*spacebackupv1.Backup](w)
			t.AddField("ID", func(v *spacebackupv1.Backup) string {
				return v.GetId()
			})
			t.AddField("NAME", func(v *spacebackupv1.Backup) string {
				return v.GetName()
			})
			t.AddField("SPACE", func(v *spacebackupv1.Backup) string {
				return v.GetSpaceId()
			})
			t.AddField("COLLECTION", func(v *spacebackupv1.Backup) string {
				return formatCollection(v.CollectionName)
			})
			t.AddField("STATUS", func(v *spacebackupv1.Backup) string {
				return output.SpaceBackupStatus(v.GetStatus())
			})
			t.AddField("SIZE", func(v *spacebackupv1.Backup) string {
				if v.GetStats() == nil {
					return ""
				}

				return formatOptionalBytes(v.GetStats().SizeBytes)
			})
			t.AddField("CREATED", func(v *spacebackupv1.Backup) string {
				if v.GetCreatedAt() != nil {
					return output.HumanTime(v.GetCreatedAt().AsTime())
				}

				return ""
			})
			t.SetItems(resp.GetItems())
			return t, nil
		},
	}.CobraCommand(s)

	util.AddPaginationFlags(cmd, "backups")
	cmd.Flags().String("space-id", "", "Filter by space ID")
	cmd.Flags().String("schedule-id", "", "Filter by the ID of the backup schedule that created the backups")
	cmd.Flags().String("collection", "", "Filter by collection name")
	_ = cmd.RegisterFlagCompletionFunc("space-id", completion.SpaceIDFlagCompletion(s))
	return cmd
}
