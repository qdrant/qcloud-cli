package hybrid

import (
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
		Short: "List all clusters in hybrid cloud environments",
		Example: `# List all hybrid cloud clusters
qcloud hybrid cluster list

# Filter by environment
qcloud hybrid cluster list --env-id 7b2ea926-724b-4de2-b73a-8675c42a6ebe`,
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

			clusters, err := client.Cluster().ListHybridClusters(ctx, accountID)
			if err != nil {
				return nil, err
			}

			if envID != "" {
				filtered := make([]*clusterv1.Cluster, 0, len(clusters))
				for _, c := range clusters {
					if c.GetCloudProviderRegionId() == envID {
						filtered = append(filtered, c)
					}
				}
				clusters = filtered
			}
			return &clusterv1.ListClustersResponse{Items: clusters}, nil
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
