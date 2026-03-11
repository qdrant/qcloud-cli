package cluster

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newUpdateCommand(s *state.State) *cobra.Command {
	return base.UpdateCmd[*clusterv1.Cluster]{
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "update <cluster-id>",
				Short: "Update an existing cluster",
				Args:  cobra.ExactArgs(1),
			}
			cmd.Flags().StringToString("label", nil, "Label to apply to the cluster ('key=value'), can be specified multiple times; replaces all existing labels")
			return cmd
		},
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*clusterv1.Cluster, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			resp, err := client.Cluster().GetCluster(ctx, &clusterv1.GetClusterRequest{
				ClusterId: args[0],
				AccountId: accountID,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get cluster: %w", err)
			}

			return resp.GetCluster(), nil
		},
		Update: func(s *state.State, cmd *cobra.Command, cluster *clusterv1.Cluster) (*clusterv1.Cluster, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			labelMap, _ := cmd.Flags().GetStringToString("label")
			cluster.Labels = nil
			for k, v := range labelMap {
				cluster.Labels = append(cluster.Labels, &commonv1.KeyValue{Key: k, Value: v})
			}
			resp, err := client.Cluster().UpdateCluster(ctx, &clusterv1.UpdateClusterRequest{
				Cluster: cluster,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to update cluster: %w", err)
			}

			return resp.GetCluster(), nil
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, updated *clusterv1.Cluster) {
			fmt.Fprintf(out, "Cluster %s (%s) updated successfully.\n", updated.GetId(), updated.GetName())
		},
	}.CobraCommand(s)
}
