package cluster

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newListCommand(s *state.State) *cobra.Command {
	cmd := base.ListCmd[*clusterv1.ListClustersResponse]{
		Use:   "list",
		Short: "List all clusters",
		Example: `# List all clusters
qcloud cluster list

# List clusters in JSON format
qcloud cluster list --json

# Filter by cloud provider and region
qcloud cluster list --cloud-provider aws --cloud-region eu-central-1

# Manual pagination
qcloud cluster list --page-size 10`,
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

			pageSizeChanged := cmd.Flags().Changed("page-size")
			pageTokenChanged := cmd.Flags().Changed("page-token")
			cloudProviderChanged := cmd.Flags().Changed("cloud-provider")
			cloudRegionChanged := cmd.Flags().Changed("cloud-region")

			var cloudProvider, cloudRegion string
			if cloudProviderChanged {
				cloudProvider, _ = cmd.Flags().GetString("cloud-provider")
			}
			if cloudRegionChanged {
				cloudRegion, _ = cmd.Flags().GetString("cloud-region")
			}

			if !pageSizeChanged && !pageTokenChanged {
				// Auto-paginate: fetch all pages and return combined results.
				var allItems []*clusterv1.Cluster
				var nextToken *string
				for {
					req := &clusterv1.ListClustersRequest{AccountId: accountID}
					if nextToken != nil {
						req.PageToken = nextToken
					}
					if cloudProviderChanged {
						req.CloudProviderId = &cloudProvider
					}
					if cloudRegionChanged {
						req.CloudProviderRegionId = &cloudRegion
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
			}

			// Manual mode: single request with provided flags.
			req := &clusterv1.ListClustersRequest{AccountId: accountID}
			if pageSizeChanged {
				ps, _ := cmd.Flags().GetInt32("page-size")
				req.PageSize = &ps
			}
			if pageTokenChanged {
				pt, _ := cmd.Flags().GetString("page-token")
				req.PageToken = &pt
			}
			if cloudProviderChanged {
				req.CloudProviderId = &cloudProvider
			}
			if cloudRegionChanged {
				req.CloudProviderRegionId = &cloudRegion
			}
			resp, err := client.Cluster().ListClusters(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("failed to list clusters: %w", err)
			}
			return resp, nil
		},
		OutputTable: func(_ *cobra.Command, w io.Writer, resp *clusterv1.ListClustersResponse) (output.TableRenderer, error) {
			t := output.NewTable[*clusterv1.Cluster](w)
			t.AddField("ID", func(v *clusterv1.Cluster) string {
				return v.GetId()
			})
			t.AddField("NAME", func(v *clusterv1.Cluster) string {
				return v.GetName()
			})
			t.AddField("STATUS", func(v *clusterv1.Cluster) string {
				if v.GetState() != nil {
					return output.ClusterPhase(v.GetState().GetPhase())
				}
				return ""
			})
			t.AddField("VERSION", func(v *clusterv1.Cluster) string {
				if v.GetConfiguration() != nil {
					return v.GetConfiguration().GetVersion()
				}
				return ""
			})
			t.AddField("CLOUD", func(v *clusterv1.Cluster) string {
				return v.GetCloudProviderId()
			})
			t.AddField("REGION / ENV", func(v *clusterv1.Cluster) string {
				return v.GetCloudProviderRegionId()
			})
			t.AddField("CREATED", func(v *clusterv1.Cluster) string {
				if v.GetCreatedAt() != nil {
					return output.HumanTime(v.GetCreatedAt().AsTime())
				}
				return ""
			})
			t.SetItems(resp.Items)
			return t, nil
		},
	}.CobraCommand(s)

	cmd.Long = `List all clusters in the current account.

By default, all clusters are fetched automatically across multiple pages.

Use --page-size and --page-token for manual pagination:
  --page-size limits how many clusters are returned per call.
  --page-token resumes from a specific page (token is printed when more pages exist).
  If --page-token is omitted, listing starts from the beginning.

Use --cloud-provider and --cloud-region to filter results server-side:
  --cloud-provider filters clusters by cloud provider ID (e.g. aws, gcp).
  --cloud-region filters clusters by cloud provider region ID (e.g. us-east-1).`

	cmd.Flags().Int32("page-size", 0, "Maximum number of clusters to return per page (manual pagination mode)")
	cmd.Flags().String("page-token", "", "Page token from a previous response to resume from (manual pagination mode)")
	cmd.Flags().String("cloud-provider", "", "Filter by cloud provider ID")
	cmd.Flags().String("cloud-region", "", "Filter by cloud provider region ID")

	_ = cmd.RegisterFlagCompletionFunc("cloud-provider", completion.CloudProviderCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("cloud-region", completion.CloudRegionCompletion(s))

	return cmd
}
