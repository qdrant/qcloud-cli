package cluster

import (
	"fmt"

	"github.com/spf13/cobra"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newUnsuspendCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Example: `# Unsuspend a cluster
qcloud cluster unsuspend 7b2ea926-724b-4de2-b73a-8675c42a6ebe`,
		BaseCobraCommand: func() *cobra.Command {
			return &cobra.Command{
				Use:   "unsuspend <cluster-id>",
				Short: "Unsuspend a cluster",
				Args:  util.ExactArgs(1, "a cluster ID"),
			}
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

			_, err = client.Cluster().UnsuspendCluster(ctx, &clusterv1.UnsuspendClusterRequest{
				AccountId: accountID,
				ClusterId: clusterID,
			})
			if err != nil {
				return fmt.Errorf("failed to unsuspend cluster: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Cluster %s unsuspending.\n", clusterID)
			return nil
		},
		ValidArgsFunction: completion.ClusterIDCompletion(s),
	}.CobraCommand(s)
}
