package space

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	spacedatasetv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/dataset/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newDatasetLoadDescribeCommand(s *state.State) *cobra.Command {
	return base.DescribeCmd[*spacedatasetv1.SampleDatasetLoad]{
		Use:   "describe <space-id> <load-id>",
		Short: "Describe a sample dataset load",
		Long: `Describe a sample dataset load of a serverless space.

Shows the status of the load, its progress while it is running and how long it
took once it finished. When a load failed, Reason explains why, for example
because the space reached its collection limit.`,
		Example: `# Describe a sample dataset load
qcloud serverless space dataset load describe 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 7f6e5d4c-3b2a-4190-8f7e-6d5c4b3a2918

# Output as JSON
qcloud serverless space dataset load describe 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 7f6e5d4c-3b2a-4190-8f7e-6d5c4b3a2918 --json`,
		Args:              util.ExactArgs(2, "a space ID and a sample dataset load ID"),
		ValidArgsFunction: completion.SpaceSampleDatasetLoadIDCompletion(s),
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*spacedatasetv1.SampleDatasetLoad, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			resp, err := client.ServerlessSampleDataset().GetSampleDatasetLoad(ctx, &spacedatasetv1.GetSampleDatasetLoadRequest{
				AccountId:           accountID,
				SpaceId:             args[0],
				SampleDatasetLoadId: args[1],
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get sample dataset load: %w", err)
			}

			return resp.GetSampleDatasetLoad(), nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, l *spacedatasetv1.SampleDatasetLoad) error {
			printSampleDatasetLoad(w, l)
			return nil
		},
	}.CobraCommand(s)
}
