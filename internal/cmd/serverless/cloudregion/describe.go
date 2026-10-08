package cloudregion

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	serverlessplatformv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/platform/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newDescribeCommand(s *state.State) *cobra.Command {
	return base.DescribeCmd[*serverlessplatformv1.CloudRegion]{
		Use:   "describe <region-id>",
		Short: "Describe a cloud region for serverless spaces",
		Long: `Describe a cloud region in which serverless spaces can be created.

Shows the region's display name, whether it currently accepts new spaces, and
its geographical location.`,
		Example: `# Describe a region
qcloud serverless cloud-region describe eu-central-1

# Output as JSON
qcloud serverless cloud-region describe eu-central-1 --json`,
		Args: util.ExactArgs(1, "a cloud region ID"),
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*serverlessplatformv1.CloudRegion, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			resp, err := client.ServerlessPlatform().GetCloudRegion(ctx, &serverlessplatformv1.GetCloudRegionRequest{
				AccountId:     accountID,
				CloudRegionId: args[0],
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get cloud region: %w", err)
			}

			return resp.GetRegion(), nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, r *serverlessplatformv1.CloudRegion) error {
			fmt.Fprintf(w, "ID:         %s\n", r.GetId())
			fmt.Fprintf(w, "Name:       %s\n", r.GetName())
			fmt.Fprintf(w, "Available:  %s\n", output.BoolYesNo(r.GetAvailable()))
			fmt.Fprintf(w, "Sub-region: %s\n", output.OptionalValue(r.GeographicalSubRegion, "not set"))
			fmt.Fprintf(w, "Country:    %s\n", output.OptionalValue(r.CountryIsoCode, "not set"))
			return nil
		},
		ValidArgsFunction: completion.ServerlessCloudRegionCompletion(s),
	}.CobraCommand(s)
}
