package cluster

import (
	"fmt"

	"github.com/spf13/cobra"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newCreateCommand(s *state.State) *cobra.Command {
	var (
		name          string
		cloudProvider string
		cloudRegion   string
		version       string
		nodes         uint32
		packageID     string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new cluster",
		Args:  cobra.NoArgs,
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
				return fmt.Errorf("failed to create cluster: %w", err)
			}

			created := resp.GetCluster()
			if s.Config.JSONOutput() {
				return output.PrintJSON(created)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Cluster %s (%s) created successfully.\n", created.GetId(), created.GetName())
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Cluster name (required)")
	cmd.Flags().StringVar(&cloudProvider, "cloud-provider", "", "Cloud provider ID (required)")
	cmd.Flags().StringVar(&cloudRegion, "cloud-region", "", "Cloud provider region ID (required)")
	cmd.Flags().StringVar(&version, "version", "", "Qdrant version")
	cmd.Flags().Uint32Var(&nodes, "nodes", 1, "Number of nodes")
	cmd.Flags().StringVar(&packageID, "package-id", "", "Booking package ID")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("cloud-provider")
	_ = cmd.MarkFlagRequired("cloud-region")

	return cmd
}
