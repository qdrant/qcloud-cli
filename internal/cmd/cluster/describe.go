package cluster

import (
	"fmt"

	"github.com/spf13/cobra"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newDescribeCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <cluster-id>",
		Short: "Describe a cluster",
		Args:  util.ExactArgs(1, "a cluster ID"),
		RunE: func(cmd *cobra.Command, args []string) error {
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
				ClusterId: args[0],
			})
			if err != nil {
				return fmt.Errorf("failed to get cluster: %w", err)
			}

			cluster := resp.GetCluster()
			if s.Config.JSONOutput() {
				return output.PrintJSON(cluster)
			}

			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "ID:       %s\n", cluster.GetId())
			fmt.Fprintf(w, "Name:     %s\n", cluster.GetName())
			if cluster.GetState() != nil {
				fmt.Fprintf(w, "Status:   %s\n", cluster.GetState().GetPhase().String())
			}
			if cluster.GetConfiguration() != nil {
				cfg := cluster.GetConfiguration()
				fmt.Fprintf(w, "Version:  %s\n", cfg.GetVersion())
				fmt.Fprintf(w, "Nodes:    %d\n", cfg.GetNumberOfNodes())
				fmt.Fprintf(w, "Package:  %s\n", cfg.GetPackageId())
			}
			fmt.Fprintf(w, "Cloud:    %s\n", cluster.GetCloudProviderId())
			fmt.Fprintf(w, "Region:   %s\n", cluster.GetCloudProviderRegionId())
			if cluster.GetCreatedAt() != nil {
				fmt.Fprintf(w, "Created:  %s\n", cluster.GetCreatedAt().AsTime().Format("2006-01-02 15:04:05 UTC"))
			}
			return nil
		},
	}
	return cmd
}
