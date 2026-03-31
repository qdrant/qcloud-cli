package hybrid

import (
	"fmt"

	"github.com/spf13/cobra"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/qcloudapi"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newClusterDeleteCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		ValidArgsFunction: hybridClusterIDCompletion(s),
		Example: `# Delete a hybrid cloud cluster (prompts for confirmation)
qcloud hybrid cluster delete 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Delete without confirmation
qcloud hybrid cluster delete 7b2ea926-724b-4de2-b73a-8675c42a6ebe --force`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "delete <cluster-id>",
				Short: "Delete a cluster in a hybrid cloud environment",
				Args:  util.ExactArgs(1, "a cluster ID"),
			}
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			clusterID := args[0]

			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return err
			}

			resp, err := client.Cluster().GetCluster(ctx, &clusterv1.GetClusterRequest{
				AccountId: accountID,
				ClusterId: clusterID,
			})
			if err != nil {
				return fmt.Errorf("failed to get cluster: %w", err)
			}

			if resp.GetCluster().GetCloudProviderId() != qcloudapi.HybridCloudProviderID {
				return fmt.Errorf("cluster %s is not a hybrid cloud cluster; use \"qcloud cluster delete\" instead", clusterID)
			}

			force, _ := cmd.Flags().GetBool("force")
			if !util.ConfirmAction(force, cmd.ErrOrStderr(), fmt.Sprintf("Are you sure you want to delete cluster %s?", clusterID)) {
				fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
				return nil
			}

			_, err = client.Cluster().DeleteCluster(ctx, &clusterv1.DeleteClusterRequest{
				AccountId: accountID,
				ClusterId: clusterID,
			})
			if err != nil {
				return fmt.Errorf("failed to delete cluster: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Cluster %s deleted.\n", clusterID)
			return nil
		},
	}.CobraCommand(s)
}
