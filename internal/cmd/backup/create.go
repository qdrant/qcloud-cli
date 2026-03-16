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

func newCreateCommand(s *state.State) *cobra.Command {
	cmd := base.CreateCmd[*backupv1.Backup]{
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "create",
				Short: "Create a backup for a cluster",
				Args:  cobra.NoArgs,
			}
			cmd.Flags().String("cluster-id", "", "Cluster ID to back up (required)")
			cmd.Flags().Uint32("retention-days", 0, "Retention period in days (1-365) (required)")
			_ = cmd.MarkFlagRequired("cluster-id")
			_ = cmd.MarkFlagRequired("retention-days")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) (*backupv1.Backup, error) {
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
			retentionDays, _ := cmd.Flags().GetUint32("retention-days")

			if retentionDays < 1 {
				return nil, fmt.Errorf("--retention-days must be at least 1")
			}

			d := time.Duration(retentionDays) * 24 * time.Hour
			b := &backupv1.Backup{
				AccountId:       accountID,
				ClusterId:       clusterID,
				RetentionPeriod: durationpb.New(d),
			}

			resp, err := client.Backup().CreateBackup(ctx, &backupv1.CreateBackupRequest{
				Backup: b,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create backup: %w", err)
			}
			return resp.GetBackup(), nil
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, b *backupv1.Backup) {
			fmt.Fprintf(out, "Backup %s created for cluster %s.\n", b.GetId(), b.GetClusterId())
		},
	}.CobraCommand(s)
	_ = cmd.RegisterFlagCompletionFunc("cluster-id", completion.ClusterIDCompletion(s))
	return cmd
}
