package cluster

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newListCommand(s *state.State) *cobra.Command {
	return base.ListCmd[*clusterv1.ListClustersResponse]{
		Use:   "list",
		Short: "List all clusters",
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

			resp, err := client.Cluster().ListClusters(ctx, &clusterv1.ListClustersRequest{
				AccountId: accountID,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list clusters: %w", err)
			}

			return resp, nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, resp *clusterv1.ListClustersResponse) error {
			t := output.NewTable[*clusterv1.Cluster](w)
			t.AddField("ID", func(v *clusterv1.Cluster) string {
				return v.GetId()
			})
			t.AddField("NAME", func(v *clusterv1.Cluster) string {
				return v.GetName()
			})
			t.AddField("STATUS", func(v *clusterv1.Cluster) string {
				if v.GetState() != nil {
					return phaseString(v.GetState().GetPhase())
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
			t.AddField("REGION", func(v *clusterv1.Cluster) string {
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
}
