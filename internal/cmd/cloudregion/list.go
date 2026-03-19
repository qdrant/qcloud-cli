package cloudregion

import (
	"fmt"
	"io"
	"strconv"

	"github.com/spf13/cobra"

	platformv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/platform/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newListCommand(s *state.State) *cobra.Command {
	cmd := base.ListCmd[*platformv1.ListCloudProviderRegionsResponse]{
		Use:   "list",
		Short: "List available cloud regions for a cloud provider",
		Fetch: func(s *state.State, cmd *cobra.Command) (*platformv1.ListCloudProviderRegionsResponse, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			cloudProvider, _ := cmd.Flags().GetString("cloud-provider")

			resp, err := client.Platform().ListCloudProviderRegions(ctx, &platformv1.ListCloudProviderRegionsRequest{
				AccountId:       accountID,
				CloudProviderId: cloudProvider,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list cloud regions: %w", err)
			}

			return resp, nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, resp *platformv1.ListCloudProviderRegionsResponse) error {
			t := output.NewTable[*platformv1.CloudProviderRegion](w)
			t.AddField("ID", func(v *platformv1.CloudProviderRegion) string {
				return v.GetId()
			})
			t.AddField("NAME", func(v *platformv1.CloudProviderRegion) string {
				return v.GetName()
			})
			t.AddField("PROVIDER", func(v *platformv1.CloudProviderRegion) string {
				return v.GetProvider()
			})
			t.AddField("AVAILABLE", func(v *platformv1.CloudProviderRegion) string {
				return strconv.FormatBool(v.GetAvailable())
			})
			t.Write(resp.GetItems())
			return nil
		},
	}.CobraCommand(s)

	cmd.Flags().String("cloud-provider", "", "Cloud provider ID (required)")
	_ = cmd.MarkFlagRequired("cloud-provider")
	_ = cmd.RegisterFlagCompletionFunc("cloud-provider", completion.CloudProviderCompletion(s))
	return cmd
}
