package serverless

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/serverless/cloudregion"
	"github.com/qdrant/qcloud-cli/internal/cmd/serverless/quota"
	"github.com/qdrant/qcloud-cli/internal/cmd/serverless/space"
	"github.com/qdrant/qcloud-cli/internal/state"
)

// NewCommand creates the "serverless" parent command and registers all subcommands.
func NewCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serverless",
		Short: "Manage Qdrant Cloud Serverless resources",
		Long: `Manage Qdrant Cloud Serverless resources.

Qdrant Cloud Serverless runs collections in spaces instead of dedicated clusters.
A space is hosted in a single cloud region and scales its search workers
automatically, so there are no nodes, packages or disks to size. Use the
commands in this group to manage spaces together with their API keys and
backups, to explore the regions spaces can be created in, and to inspect the
account's serverless quota.`,
		Example: `# List the regions in which spaces can be created
qcloud serverless cloud-region list

# Show how many spaces the account can still create
qcloud serverless quota

# List all serverless spaces
qcloud serverless space list

# Create a space in a region and wait until it is ready
qcloud serverless space create --cloud-region eu-central-1 --wait`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(
		space.NewCommand(s),
		cloudregion.NewCommand(s),
		quota.NewCommand(s),
	)
	return cmd
}
