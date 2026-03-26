package hybrid

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	hybridv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/hybrid/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newListCommand(s *state.State) *cobra.Command {
	return base.ListCmd[*hybridv1.ListHybridCloudEnvironmentsResponse]{
		Use:   "list",
		Short: "List all hybrid cloud environments",
		Fetch: func(s *state.State, cmd *cobra.Command) (*hybridv1.ListHybridCloudEnvironmentsResponse, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			resp, err := client.Hybrid().ListHybridCloudEnvironments(ctx, &hybridv1.ListHybridCloudEnvironmentsRequest{
				AccountId: accountID,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list hybrid cloud environments: %w", err)
			}
			return resp, nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, resp *hybridv1.ListHybridCloudEnvironmentsResponse) error {
			t := output.NewTable[*hybridv1.HybridCloudEnvironment](w)
			t.AddField("ID", func(v *hybridv1.HybridCloudEnvironment) string {
				return v.GetId()
			})
			t.AddField("NAME", func(v *hybridv1.HybridCloudEnvironment) string {
				return v.GetName()
			})
			t.AddField("STATUS", func(v *hybridv1.HybridCloudEnvironment) string {
				if v.GetStatus() != nil {
					return output.HybridEnvironmentPhase(v.GetStatus().GetPhase())
				}
				return ""
			})
			t.AddField("NODES", func(v *hybridv1.HybridCloudEnvironment) string {
				if v.GetStatus() != nil {
					return fmt.Sprintf("%d", v.GetStatus().GetNumberOfNodes())
				}
				return ""
			})
			t.AddField("CREATED", func(v *hybridv1.HybridCloudEnvironment) string {
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
