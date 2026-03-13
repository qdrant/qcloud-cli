package cluster

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newWaitCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "wait <cluster-id>",
				Short: "Wait for a cluster to become healthy",
				Args:  util.ExactArgs(1, "a cluster ID"),
			}
			cmd.Flags().Duration("timeout", 10*time.Minute, "Maximum time to wait for the cluster to become healthy")
			cmd.Flags().Duration("poll-interval", 5*time.Second, "How often to poll for cluster health")
			_ = cmd.Flags().MarkHidden("poll-interval")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return err
			}

			timeout, _ := cmd.Flags().GetDuration("timeout")
			pollInterval, _ := cmd.Flags().GetDuration("poll-interval")
			cluster, err := waitForHealthyWithInterval(
				ctx, client.Cluster(), cmd.ErrOrStderr(),
				accountID, args[0], timeout, pollInterval,
			)
			if err != nil {
				return err
			}

			if ep := cluster.GetState().GetEndpoint(); ep != nil && ep.GetUrl() != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Cluster %s (%s) is ready. Endpoint: %s\n",
					cluster.GetId(), cluster.GetName(), ep.GetUrl())
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Cluster %s (%s) is ready.\n",
					cluster.GetId(), cluster.GetName())
			}
			return nil
		},
		ValidArgsFunction: completion.ClusterIDCompletion(s),
	}.CobraCommand(s)
}
