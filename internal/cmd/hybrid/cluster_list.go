package hybrid

import (
	"fmt"
	"io"
	"strings"

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

			provider := "hybrid"
			req := &clusterv1.ListClustersRequest{
				AccountId:       accountID,
				CloudProviderId: &provider,
			}

			if cmd.Flags().Changed("env-id") {
				envID, _ := cmd.Flags().GetString("env-id")
				req.CloudProviderRegionId = &envID
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
			return &clusterv1.ListClustersResponse{Items: allItems}, nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, resp *clusterv1.ListClustersResponse) error {
			t := output.NewTable[*clusterv1.Cluster](w)
			t.AddField("ID", func(v *clusterv1.Cluster) string { return v.GetId() })
			t.AddField("NAME", func(v *clusterv1.Cluster) string { return v.GetName() })
			t.AddField("STATUS", func(v *clusterv1.Cluster) string {
				return strings.TrimPrefix(v.GetState().GetPhase().String(), "CLUSTER_PHASE_")
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
	return cmd
}
