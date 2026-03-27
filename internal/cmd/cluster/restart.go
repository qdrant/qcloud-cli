package cluster

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/clusterutil"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newRestartCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Example: `# Restart a cluster (prompts for confirmation)
qcloud cluster restart 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Restart without confirmation and wait for healthy status
qcloud cluster restart 7b2ea926-724b-4de2-b73a-8675c42a6ebe --force --wait`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "restart <cluster-id>",
				Short: "Restart a cluster",
				Args:  util.ExactArgs(1, "a cluster ID"),
			}
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
			cmd.Flags().Bool("wait", false, "Wait for the cluster to restart to a healthy status")
			cmd.Flags().Duration("wait-timeout", 10*time.Minute, "Maximum time to wait for cluster the cluster to restart to healthy status")
			cmd.Flags().Duration("wait-poll-interval", 5*time.Second, "How often to poll for the cluster to restart to healthy status")
			_ = cmd.Flags().MarkHidden("wait-poll-interval")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			clusterID := args[0]

			force, _ := cmd.Flags().GetBool("force")
			if !util.ConfirmAction(force, cmd.ErrOrStderr(), fmt.Sprintf("Are you sure you want to restart cluster %s?", clusterID)) {
				fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
				return nil
			}

			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return err
			}

			_, err = client.Cluster().RestartCluster(ctx, &clusterv1.RestartClusterRequest{
				AccountId: accountID,
				ClusterId: clusterID,
			})
			if err != nil {
				return fmt.Errorf("failed to restart cluster: %w", err)
			}

			wait, _ := cmd.Flags().GetBool("wait")
			if !wait {
				fmt.Fprintf(cmd.OutOrStdout(), "Cluster %s restarting.\n", clusterID)
				return nil
			}

			waitTimeout, _ := cmd.Flags().GetDuration("wait-timeout")
			waitPollInterval, _ := cmd.Flags().GetDuration("wait-poll-interval")
			fmt.Fprintf(cmd.ErrOrStderr(), "Cluster %s restarting, waiting for it to become healthy...\n", clusterID)
			cluster, err := clusterutil.WaitForClusterHealthy(ctx, client.Cluster(), cmd.ErrOrStderr(), accountID, clusterID, waitTimeout, waitPollInterval)
			if err != nil {
				return err
			}
			if ep := cluster.GetState().GetEndpoint(); ep != nil && ep.GetUrl() != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Cluster %s (%s) is ready. Endpoint: %s\n", cluster.GetId(), cluster.GetName(), ep.GetUrl())
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Cluster %s (%s) is ready.\n", cluster.GetId(), cluster.GetName())
			}
			return nil
		},
		ValidArgsFunction: completion.ClusterIDCompletion(s),
	}.CobraCommand(s)
}
