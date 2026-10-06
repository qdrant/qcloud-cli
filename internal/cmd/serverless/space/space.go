package space

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// NewCommand creates the "space" parent command and registers all subcommands.
func NewCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "space",
		Short: "Manage serverless spaces",
		Long: `Manage serverless spaces.

A space is the unit of deployment in Qdrant Cloud Serverless. It lives in a single
cloud region, exposes one endpoint for the REST and gRPC APIs, and holds any number
of collections up to the account's quota. Spaces are configured with network
restrictions (allowed IP ranges and browser origins), per-collection size limits
and search-worker settings.`,
		Example: `# List all spaces
qcloud serverless space list

# Create a space with a generated name
qcloud serverless space create --cloud-region aws-eu-central-1

# Show the details of a space
qcloud serverless space describe 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(
		newListCommand(s),
		newDescribeCommand(s),
		newCreateCommand(s),
		newCreateFromBackupCommand(s),
		newUpdateCommand(s),
		newDeleteCommand(s),
		newWaitCommand(s),
		newSuggestNameCommand(s),
		newKeyCommand(s),
	)
	return cmd
}
