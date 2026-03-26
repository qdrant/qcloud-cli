package hybrid

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newClusterListCommand(s *state.State) *cobra.Command {
	cmd := base.ListCmd[*clusterv1.ListClustersResponse]{
		Use:   "list",
		Short: "List clusters in hybrid cloud environments",
		Fetch: func(s *state.State, cmd *cobra.Command) (*clusterv1.ListClustersResponse, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			envID, _ := cmd.Flags().GetString("env-id")

			req := &clusterv1.ListClustersRequest{
				AccountId: accountID,
			}

			// Auto-paginate.
			var allItems []*clusterv1.Cluster
			var nextToken *string
			for {
				if nextToken != nil {
					req.PageToken = nextToken
				}
				resp, err := client.Cluster().ListClusters(ctx, req)
				if err != nil {
					return nil, fmt.Errorf("failed to list clusters: %w", err)
				}
				allItems = append(allItems, resp.Items...)
				if resp.NextPageToken == nil || *resp.NextPageToken == "" {
					break
				}
				nextToken = resp.NextPageToken
			}

			// Filter client-side: the API requires a valid region UUID when
			// cloud_provider_id is "hybrid", so we fetch all clusters and
			// filter here instead.
			filtered := make([]*clusterv1.Cluster, 0, len(allItems))
			for _, c := range allItems {
				if c.GetCloudProviderId() != hybridCloudProviderID {
					continue
				}
				if envID != "" && c.GetCloudProviderRegionId() != envID {
					continue
				}
				filtered = append(filtered, c)
			}
			return &clusterv1.ListClustersResponse{Items: filtered}, nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, resp *clusterv1.ListClustersResponse) error {
			t := output.NewTable[*clusterv1.Cluster](w)
			t.AddField("ID", func(v *clusterv1.Cluster) string { return v.GetId() })
			t.AddField("NAME", func(v *clusterv1.Cluster) string { return v.GetName() })
			t.AddField("STATUS", func(v *clusterv1.Cluster) string {
				return output.ClusterPhase(v.GetState().GetPhase())
			})
			t.AddField("VERSION", func(v *clusterv1.Cluster) string {
				return v.GetConfiguration().GetVersion()
			})
			t.AddField("ENV", func(v *clusterv1.Cluster) string {
				return v.GetCloudProviderRegionId()
			})
			t.AddField("CREATED", func(v *clusterv1.Cluster) string {
				if v.GetCreatedAt() != nil {
					return output.HumanTime(v.GetCreatedAt().AsTime())
				}
				return ""
			})
			t.Write(resp.Items)
			return nil
		},
	}.CobraCommand(s)

	cmd.Flags().String("env-id", "", "Filter by hybrid cloud environment ID")
	_ = cmd.RegisterFlagCompletionFunc("env-id", envIDCompletion(s))
	return cmd
}
