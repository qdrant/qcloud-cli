package cloudregion

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	serverlessplatformv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/platform/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newListCommand(s *state.State) *cobra.Command {
	return base.ListCmd[*serverlessplatformv1.ListCloudRegionsResponse]{
		Use:   "list",
		Short: "List cloud regions for serverless spaces",
		Long: `List the cloud regions in which serverless spaces can be created.

The AVAILABLE column shows whether a region currently accepts new spaces.`,
		Example: `# List all serverless regions
qcloud serverless cloud-region list

# List the IDs of the available regions
qcloud serverless cloud-region list --json | jq -r '.items[] | select(.available) | .id'`,
		Fetch: func(s *state.State, cmd *cobra.Command) (*serverlessplatformv1.ListCloudRegionsResponse, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			resp, err := client.ServerlessPlatform().ListCloudRegions(ctx, &serverlessplatformv1.ListCloudRegionsRequest{
				AccountId: accountID,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list cloud regions: %w", err)
			}

			return resp, nil
		},
		OutputTable: func(_ *cobra.Command, w io.Writer, resp *serverlessplatformv1.ListCloudRegionsResponse) (output.TableRenderer, error) {
			t := output.NewTable[*serverlessplatformv1.CloudRegion](w)
			t.AddField("ID", func(v *serverlessplatformv1.CloudRegion) string {
				return v.GetId()
			})
			t.AddField("NAME", func(v *serverlessplatformv1.CloudRegion) string {
				return v.GetName()
			})
			t.AddField("AVAILABLE", func(v *serverlessplatformv1.CloudRegion) string {
				return output.BoolYesNo(v.GetAvailable())
			})
			t.AddField("SUB-REGION", func(v *serverlessplatformv1.CloudRegion) string {
				return output.OptionalValue(v.GeographicalSubRegion, "-")
			})
			t.AddField("COUNTRY", func(v *serverlessplatformv1.CloudRegion) string {
				return output.OptionalValue(v.CountryIsoCode, "-")
			})
			t.SetItems(resp.GetItems())
			return t, nil
		},
	}.CobraCommand(s)
}
