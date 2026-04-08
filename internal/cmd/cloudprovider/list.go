package cloudprovider

import (
	"fmt"
	"io"
	"strconv"

	"github.com/spf13/cobra"

	platformv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/platform/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newListCommand(s *state.State) *cobra.Command {
	return base.ListCmd[*platformv1.ListCloudProvidersResponse]{
		Use:   "list",
		Short: "List available cloud providers",
		Fetch: func(s *state.State, cmd *cobra.Command) (*platformv1.ListCloudProvidersResponse, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			resp, err := client.Platform().ListCloudProviders(ctx, &platformv1.ListCloudProvidersRequest{
				AccountId: accountID,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list cloud providers: %w", err)
			}

			return resp, nil
		},
		OutputTable: func(_ *cobra.Command, w io.Writer, resp *platformv1.ListCloudProvidersResponse) output.Renderable {
			t := output.NewTable[*platformv1.CloudProvider](w)
			t.AddField("ID", func(v *platformv1.CloudProvider) string {
				return v.GetId()
			})
			t.AddField("NAME", func(v *platformv1.CloudProvider) string {
				return v.GetName()
			})
			t.AddField("AVAILABLE", func(v *platformv1.CloudProvider) string {
				return strconv.FormatBool(v.GetAvailable())
			})
			t.SetItems(resp.GetItems())
			return t
		},
	}.CobraCommand(s)
}
