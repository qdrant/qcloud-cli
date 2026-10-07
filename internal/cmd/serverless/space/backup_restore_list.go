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

func newBackupRestoreListCommand(s *state.State) *cobra.Command {
	cmd := base.ListCmd[*spacebackupv1.ListBackupRestoresResponse]{
		Use:   "list",
		Short: "List restores of serverless space backups",
		Long: `List restores of serverless space backups in the current account.

The PROGRESS column is reported while a restore is running and may remain after
it completes. By default, all restores are fetched automatically across multiple
pages. Use --page-size and --page-token for manual pagination.`,
		Example: `# List all restores in the account
qcloud serverless space backup restore list

# List restores of a space
qcloud serverless space backup restore list --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60`,
		Fetch: func(s *state.State, cmd *cobra.Command) (*spacebackupv1.ListBackupRestoresResponse, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			items, next, err := util.FetchPages(cmd, func(pageSize *int32, pageToken *string) ([]*spacebackupv1.BackupRestore, string, error) {
				req := &spacebackupv1.ListBackupRestoresRequest{
					AccountId: accountID,
					PageSize:  pageSize,
					PageToken: pageToken,
				}
				if cmd.Flags().Changed("space-id") {
					v, _ := cmd.Flags().GetString("space-id")
					req.SpaceId = &v
				}

				resp, err := client.ServerlessBackup().ListBackupRestores(ctx, req)
				if err != nil {
					return nil, "", fmt.Errorf("failed to list backup restores: %w", err)
				}

				return resp.GetItems(), resp.GetNextPageToken(), nil
			})
			if err != nil {
				return nil, err
			}

			return &spacebackupv1.ListBackupRestoresResponse{Items: items, NextPageToken: next}, nil
		},
		OutputTable: func(_ *cobra.Command, w io.Writer, resp *spacebackupv1.ListBackupRestoresResponse) (output.TableRenderer, error) {
			t := output.NewTable[*spacebackupv1.BackupRestore](w)
			t.AddField("ID", func(v *spacebackupv1.BackupRestore) string {
				return v.GetId()
			})
			t.AddField("BACKUP", func(v *spacebackupv1.BackupRestore) string {
				return v.GetBackupId()
			})
			t.AddField("SPACE", func(v *spacebackupv1.BackupRestore) string {
				return v.GetSpaceId()
			})
			t.AddField("STATUS", func(v *spacebackupv1.BackupRestore) string {
				return output.SpaceBackupRestoreStatus(v.GetStatus())
			})
			t.AddField("PROGRESS", func(v *spacebackupv1.BackupRestore) string {
				return v.GetStats().GetProgress()
			})
			t.AddField("DURATION", func(v *spacebackupv1.BackupRestore) string {
				return formatOptionalDuration(v.GetStats().GetDuration())
			})
			t.AddField("CREATED", func(v *spacebackupv1.BackupRestore) string {
				if v.GetCreatedAt() != nil {
					return output.HumanTime(v.GetCreatedAt().AsTime())
				}

				return ""
			})
			t.SetItems(resp.GetItems())
			return t, nil
		},
	}.CobraCommand(s)

	util.AddPaginationFlags(cmd, "restores")
	cmd.Flags().String("space-id", "", "Filter by space ID")
	_ = cmd.RegisterFlagCompletionFunc("space-id", completion.SpaceIDFlagCompletion(s))
	return cmd
}
