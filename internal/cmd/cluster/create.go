package cluster

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newCreateCommand(s *state.State) *cobra.Command {
	return base.CreateCmd[*clusterv1.Cluster]{
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "create",
				Short: "Create a new cluster",
				Args:  cobra.NoArgs,
			}
			cmd.Flags().String("name", "", "Cluster name (required)")
			cmd.Flags().String("cloud-provider", "", "Cloud provider ID (required)")
			cmd.Flags().String("cloud-region", "", "Cloud provider region ID (required)")
			cmd.Flags().String("version", "", "Qdrant version")
			cmd.Flags().Uint32("nodes", 1, "Number of nodes")
			cmd.Flags().String("package-id", "", "Booking package ID")
			_ = cmd.MarkFlagRequired("name")
			_ = cmd.MarkFlagRequired("cloud-provider")
			_ = cmd.MarkFlagRequired("cloud-region")
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
			name, _ := cmd.Flags().GetString("name")
			cloudProvider, _ := cmd.Flags().GetString("cloud-provider")
			cloudRegion, _ := cmd.Flags().GetString("cloud-region")
			version, _ := cmd.Flags().GetString("version")
			nodes, _ := cmd.Flags().GetUint32("nodes")
			packageID, _ := cmd.Flags().GetString("package-id")

			cluster := &clusterv1.Cluster{
				AccountId:             accountID,
				Name:                  name,
				CloudProviderId:       cloudProvider,
				CloudProviderRegionId: cloudRegion,
				Configuration: &clusterv1.ClusterConfiguration{
					NumberOfNodes: nodes,
				},
			}
			if version != "" {
				cluster.Configuration.Version = &version
			}
			if packageID != "" {
				cluster.Configuration.PackageId = packageID
			}

			resp, err := client.Cluster().CreateCluster(ctx, &clusterv1.CreateClusterRequest{
				Cluster: cluster,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create cluster: %w", err)
			}
			return resp.GetCluster(), nil
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, created *clusterv1.Cluster) {
			fmt.Fprintf(out, "Cluster %s (%s) created successfully.\n", created.GetId(), created.GetName())
		},
	}.CobraCommand(s)
}
