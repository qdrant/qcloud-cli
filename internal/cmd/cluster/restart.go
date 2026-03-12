package cluster

import (
	"fmt"

	"github.com/spf13/cobra"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newRestartCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "restart <cluster-id>",
				Short: "Restart a cluster",
				Args:  util.ExactArgs(1, "a cluster ID"),
			}
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			clusterID := args[0]

			force, _ := cmd.Flags().GetBool("force")
			if !util.ConfirmAction(force, fmt.Sprintf("Are you sure you want to restart cluster %s?", clusterID)) {
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

			fmt.Fprintf(cmd.OutOrStdout(), "Cluster %s restarting.\n", clusterID)
			return nil
		},
	}.CobraCommand(s)
}
