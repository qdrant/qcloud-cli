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
			cmd.Flags().Bool("wait", false, "Wait for the cluster to become healthy")
			cmd.Flags().Duration("wait-timeout", 10*time.Minute, "Maximum time to wait for cluster health")
			cmd.Flags().Duration("wait-poll-interval", 5*time.Second, "How often to poll for cluster health")
			_ = cmd.Flags().MarkHidden("wait-poll-interval")
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
			return clusterutil.WaitForClusterHealthy(ctx, client.Cluster(), cmd.ErrOrStderr(), accountID, created.GetId(), waitTimeout, pollInterval)
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, created *clusterv1.Cluster) {
			if ep := created.GetState().GetEndpoint(); ep != nil && ep.GetUrl() != "" {
				fmt.Fprintf(out, "Cluster %s (%s) is ready. Endpoint: %s\n", created.GetId(), created.GetName(), ep.GetUrl())
			} else {
				fmt.Fprintf(out, "Cluster %s (%s) created from backup.\n", created.GetId(), created.GetName())
			}
		},
	}.CobraCommand(s)
}
