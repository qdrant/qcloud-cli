package cluster

import (
	"fmt"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newListCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all clusters",
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

			resp, err := client.Cluster().ListClusters(ctx, &clusterv1.ListClustersRequest{
				AccountId: accountID,
			})
			if err != nil {
				return fmt.Errorf("failed to list clusters: %w", err)
			}

			if s.Config.JSONOutput() {
				msgs := make([]proto.Message, len(resp.Items))
				for i, c := range resp.Items {
					msgs[i] = c
				}
				return output.PrintJSON(cmd.OutOrStdout(), msgs)
			}

			t := output.NewTable(cmd.OutOrStdout())
			t.AddField("ID", func(v any) string {
				return v.(*clusterv1.Cluster).GetId()
			})
			t.AddField("NAME", func(v any) string {
				return v.(*clusterv1.Cluster).GetName()
			})
			t.AddField("STATUS", func(v any) string {
				c := v.(*clusterv1.Cluster)
				if c.GetState() != nil {
					return c.GetState().GetPhase().String()
				}
				return ""
			})
			t.AddField("VERSION", func(v any) string {
				c := v.(*clusterv1.Cluster)
				if c.GetConfiguration() != nil {
					return c.GetConfiguration().GetVersion()
				}
				return ""
			})
			t.AddField("CLOUD", func(v any) string {
				return v.(*clusterv1.Cluster).GetCloudProviderId()
			})
			t.AddField("REGION", func(v any) string {
				return v.(*clusterv1.Cluster).GetCloudProviderRegionId()
			})

			items := make([]any, len(resp.Items))
			for i, c := range resp.Items {
				items[i] = c
			}
			t.Write(items)
			return nil
		},
	}
	return cmd
}
