package cluster

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"

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
			cmd.Flags().String("package", "", "Booking package name or ID (see 'cluster package list')")
			cmd.Flags().StringToString("label", nil, "Label to apply to the cluster ('key=value'), can be specified multiple times")
			cmd.Flags().Bool("wait", false, "Wait for the cluster to become healthy")
			cmd.Flags().Duration("wait-timeout", 10*time.Minute, "Maximum time to wait for cluster health")
			cmd.Flags().Duration("wait-poll-interval", 5*time.Second, "How often to poll for cluster health")
			_ = cmd.Flags().MarkHidden("wait-poll-interval")
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
			packageValue, _ := cmd.Flags().GetString("package")
			labelMap, _ := cmd.Flags().GetStringToString("label")

			var packageID string
			if packageValue != "" {
				if isUUID(packageValue) {
					packageID = packageValue
				} else {
					pkg, err := resolvePackageByName(ctx, client.Booking(), accountID, cloudProvider, cloudRegion, packageValue)
					if err != nil {
						return nil, err
					}
					packageID = pkg.GetId()
				}
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
			for k, v := range labelMap {
				cluster.Labels = append(cluster.Labels, &commonv1.KeyValue{Key: k, Value: v})
			}

			resp, err := client.Cluster().CreateCluster(ctx, &clusterv1.CreateClusterRequest{
				Cluster: cluster,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create cluster: %w", err)
			}

			created := resp.GetCluster()

			wait, _ := cmd.Flags().GetBool("wait")
			if !wait {
				return created, nil
			}

			waitTimeout, _ := cmd.Flags().GetDuration("wait-timeout")
			pollInterval, _ := cmd.Flags().GetDuration("wait-poll-interval")
			fmt.Fprintf(cmd.ErrOrStderr(), "Cluster %s created, waiting for it to become healthy...\n", created.GetId())
			return waitForHealthyWithInterval(ctx, client.Cluster(), cmd.ErrOrStderr(), accountID, created.GetId(), waitTimeout, pollInterval)
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, created *clusterv1.Cluster) {
			if ep := created.GetState().GetEndpoint(); ep != nil && ep.GetUrl() != "" {
				fmt.Fprintf(out, "Cluster %s (%s) is ready. Endpoint: %s\n", created.GetId(), created.GetName(), ep.GetUrl())
			} else {
				fmt.Fprintf(out, "Cluster %s (%s) created successfully.\n", created.GetId(), created.GetName())
			}
		},
	}.CobraCommand(s)
}
