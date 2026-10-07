package cluster

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/clusterutil"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newCreateFromBackupCommand(s *state.State) *cobra.Command {
	return base.CreateCmd[*clusterv1.Cluster]{
		Example: `# Create a cluster from a backup
qcloud cluster create-from-backup --backup-id <backup-id> --name my-restored-cluster

# Create a cluster from a backup and wait until it is healthy
qcloud cluster create-from-backup --backup-id <backup-id> --name my-restored-cluster --wait

# Create with a custom wait timeout
qcloud cluster create-from-backup --backup-id <backup-id> --name my-restored-cluster --wait --wait-timeout 20m`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "create-from-backup",
				Short: "Create a new cluster from a backup",
				Long: `Create a new Qdrant Cloud cluster seeded with the data from an existing backup.

The new cluster is provisioned using the same configuration as the original cluster
at the time the backup was taken. The backup must belong to the current account.`,
				Args: cobra.NoArgs,
			}
			cmd.Flags().String("backup-id", "", "ID of the backup to restore from (required)")
			cmd.Flags().String("name", "", "Name for the new cluster (required)")
			util.AddWaitFlags(cmd, "the cluster to become healthy", 10*time.Minute, 5*time.Second)
			_ = cmd.MarkFlagRequired("backup-id")
			_ = cmd.MarkFlagRequired("name")
			_ = cmd.RegisterFlagCompletionFunc("backup-id", completion.BackupIDCompletion(s))
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) (*clusterv1.Cluster, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			backupID, _ := cmd.Flags().GetString("backup-id")
			name, _ := cmd.Flags().GetString("name")

			resp, err := client.Cluster().CreateClusterFromBackup(ctx, &clusterv1.CreateClusterFromBackupRequest{
				AccountId:   accountID,
				BackupId:    backupID,
				ClusterName: name,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create cluster from backup: %w", err)
			}

			created := resp.GetCluster()

			wait, _ := cmd.Flags().GetBool("wait")
			if !wait {
				return created, nil
			}

			waitTimeout, _ := cmd.Flags().GetDuration("wait-timeout")
			pollInterval, _ := cmd.Flags().GetDuration("wait-poll-interval")
			fmt.Fprintf(cmd.ErrOrStderr(), "Cluster %s created, waiting for it to become healthy...\n", created.GetId())
			healthy, err := clusterutil.WaitForClusterHealthy(ctx, client.Cluster(), cmd.ErrOrStderr(), accountID, created.GetId(), waitTimeout, pollInterval)
			if err != nil {
				if s.Config.JSONOutput() {
					_ = output.PrintJSON(cmd.OutOrStdout(), created)
				} else {
					fmt.Fprint(cmd.OutOrStdout(), clusterResultMessage(created, "created from backup"))
				}

				return nil, fmt.Errorf("cluster %s was created but did not become healthy: %w", created.GetId(), err)
			}

			return healthy, nil
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, created *clusterv1.Cluster) {
			fmt.Fprint(out, clusterResultMessage(created, "created from backup"))
		},
	}.CobraCommand(s)
}
