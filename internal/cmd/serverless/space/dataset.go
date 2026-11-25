package space

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

func newDatasetCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataset",
		Short: "Load sample datasets into a space",
		Long: `Load sample datasets into a serverless space.

A sample dataset is a prebuilt dataset, with vectors already computed, that can
be loaded into a new collection of a space to try Qdrant without bringing your
own data. Only the datasets available in the region of the space are offered.
The collection size of a dataset counts against the collection size limits of
the space.

A load is asynchronous: it creates the collection and copies the data in the
background. Use "dataset load list" or "dataset load describe" to follow it, or
pass --wait to "dataset load start".`,
		Example: `# List the sample datasets that can be loaded into a space
qcloud serverless space dataset list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Show the vectors and size of a sample dataset
qcloud serverless space dataset describe 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 midjourney-styles

# Load a sample dataset into a new collection and wait until it is done
qcloud serverless space dataset load start 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 midjourney-styles --wait

# List the sample dataset loads of a space
qcloud serverless space dataset load list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(
		newDatasetListCommand(s),
		newDatasetDescribeCommand(s),
		newDatasetLoadCommand(s),
	)
	return cmd
}

func newDatasetLoadCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load",
		Short: "Manage sample dataset loads of a space",
		Long: `Manage sample dataset loads of a serverless space.

A load copies a sample dataset into a new collection of the space. It runs in
the background and ends as SUCCEEDED or FAILED; the reason of a failure (for
example the collection limit of the space) is shown by "load describe". Only
loads created in the last 30 days are listed.`,
		Example: `# Load a sample dataset into a collection with the same name as the dataset
qcloud serverless space dataset load start 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 midjourney-styles

# Load a sample dataset into a collection with a custom name
qcloud serverless space dataset load start 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 midjourney-styles --collection styles

# List the loads of a space
qcloud serverless space dataset load list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Follow a single load
qcloud serverless space dataset load describe 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 7f6e5d4c-3b2a-4190-8f7e-6d5c4b3a2918`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(
		newDatasetLoadStartCommand(s),
		newDatasetLoadListCommand(s),
		newDatasetLoadDescribeCommand(s),
	)
	return cmd
}
