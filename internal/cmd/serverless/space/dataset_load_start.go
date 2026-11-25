package space

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	spacedatasetv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/dataset/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newDatasetLoadStartCommand(s *state.State) *cobra.Command {
	return base.CreateCmd[*spacedatasetv1.SampleDatasetLoad]{
		Long: `Load a sample dataset into a new collection of a serverless space.

The collection is created by the load, so no collection with that name may
exist in the space. Without --collection, the collection gets the ID of the
dataset as its name.

The load is asynchronous: the command returns as soon as the load is accepted.
With --wait it follows the load until it succeeds, and fails when the load
fails (for example because the space reached its collection limit). The load
is printed before the error, so its ID is not lost.`,
		Example: `# Load a sample dataset into a collection named after the dataset
qcloud serverless space dataset load start 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 midjourney-styles

# Load a sample dataset into a collection named "styles" and wait until it is done
qcloud serverless space dataset load start 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 midjourney-styles \
  --collection styles --wait

# Start a load and keep its ID for scripting
qcloud serverless space dataset load start 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 midjourney-styles --json | jq -r .id`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "start <space-id> <dataset-id>",
				Short: "Load a sample dataset into a new collection",
				Args:  util.ExactArgs(2, "a space ID and a sample dataset ID"),
			}
			cmd.Flags().String("collection", "", "Name of the new collection (default: the dataset ID)")
			util.AddWaitFlags(cmd, "the sample dataset load to finish", 30*time.Minute, 5*time.Second)
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) (*spacedatasetv1.SampleDatasetLoad, error) {
			spaceID, datasetID := args[0], args[1]

			collection := datasetID
			if cmd.Flags().Changed("collection") {
				collection, _ = cmd.Flags().GetString("collection")
				if collection == "" {
					return nil, fmt.Errorf("--collection must not be empty")
				}
			}

			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			resp, err := client.ServerlessSampleDataset().LoadSampleDataset(ctx, &spacedatasetv1.LoadSampleDatasetRequest{
				AccountId:       accountID,
				SpaceId:         spaceID,
				SampleDatasetId: datasetID,
				CollectionName:  collection,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to load sample dataset: %w", err)
			}

			started := resp.GetSampleDatasetLoad()

			wait, _ := cmd.Flags().GetBool("wait")
			if !wait {
				return started, nil
			}

			waitTimeout, _ := cmd.Flags().GetDuration("wait-timeout")
			pollInterval, _ := cmd.Flags().GetDuration("wait-poll-interval")

			fmt.Fprintf(cmd.ErrOrStderr(), "Sample dataset load %s started, waiting for it to finish...\n", started.GetId())
			done, err := waitForSampleDatasetLoad(ctx, client.ServerlessSampleDataset(), cmd.ErrOrStderr(),
				accountID, spaceID, started.GetId(), waitTimeout, pollInterval)
			if err != nil {
				// The load exists even though it did not succeed: print it so
				// that its ID and last status are not lost.
				last := done
				if last == nil {
					last = started
				}

				if s.Config.JSONOutput() {
					_ = output.PrintJSON(cmd.OutOrStdout(), last)
				} else {
					printSampleDatasetLoad(cmd.OutOrStdout(), last)
				}

				return nil, fmt.Errorf("sample dataset load %s did not succeed: %w", started.GetId(), err)
			}

			return done, nil
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, l *spacedatasetv1.SampleDatasetLoad) {
			if l.GetStatus() == spacedatasetv1.SampleDatasetLoadStatus_SAMPLE_DATASET_LOAD_STATUS_SUCCEEDED {
				fmt.Fprintf(out, "Sample dataset %s loaded into collection %s (load %s).\n",
					l.GetSampleDatasetId(), l.GetCollectionName(), l.GetId())
				return
			}

			fmt.Fprintf(out, "Sample dataset load %s started: loading %s into collection %s.\n",
				l.GetId(), l.GetSampleDatasetId(), l.GetCollectionName())
			fmt.Fprintf(out, "Follow it with: qcloud serverless space dataset load describe %s %s\n", l.GetSpaceId(), l.GetId())
		},
		ValidArgsFunction: completion.SpaceSampleDatasetIDCompletion(s),
	}.CobraCommand(s)
}
