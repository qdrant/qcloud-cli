package cloudregion

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// NewCommand creates the "serverless cloud-region" parent command and registers all subcommands.
func NewCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloud-region",
		Short: "Explore cloud regions for serverless spaces",
		Long: `Explore the cloud regions in which serverless spaces can be created.

Serverless regions are identified by provider-agnostic IDs. Pass a region ID to
'qcloud serverless space create --cloud-region' to host a new space there. Only
regions marked as available accept new spaces.`,
		Example: `# List all serverless regions
qcloud serverless cloud-region list

# Show the details of a region
qcloud serverless cloud-region describe eu-central-1`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(
		newListCommand(s),
		newDescribeCommand(s),
	)
	return cmd
}
