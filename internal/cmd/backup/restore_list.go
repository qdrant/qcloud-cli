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

func newRestoreListCommand(s *state.State) *cobra.Command {
	cmd := base.ListCmd[*backupv1.ListBackupRestoresResponse]{
		Use:   "list",
		Short: "List backup restores",
		Fetch: func(s *state.State, cmd *cobra.Command) (*backupv1.ListBackupRestoresResponse, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			req := &backupv1.ListBackupRestoresRequest{AccountId: accountID}
			if cmd.Flags().Changed("cluster-id") {
				clusterID, _ := cmd.Flags().GetString("cluster-id")
				req.ClusterId = &clusterID
			}

			resp, err := client.Backup().ListBackupRestores(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("failed to list backup restores: %w", err)
			}
			return resp, nil
		},
		OutputTable: func(_ *cobra.Command, w io.Writer, resp *backupv1.ListBackupRestoresResponse) (output.TableRenderer, error) {
			t := output.NewTable[*backupv1.BackupRestore](w)
			t.AddField("ID", func(v *backupv1.BackupRestore) string {
				return v.GetId()
			})
			t.AddField("BACKUP", func(v *backupv1.BackupRestore) string {
				return v.GetBackupId()
			})
			t.AddField("CLUSTER", func(v *backupv1.BackupRestore) string {
				return v.GetClusterId()
			})
			t.AddField("STATUS", func(v *backupv1.BackupRestore) string {
				return output.BackupRestoreStatus(v.GetStatus())
			})
			t.AddField("CREATED", func(v *backupv1.BackupRestore) string {
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
