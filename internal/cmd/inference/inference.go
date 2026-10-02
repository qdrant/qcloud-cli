package inference

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// NewCommand creates the "inference" parent command and registers all subcommands.
func NewCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inference",
		Short: "Manage inference resources",
		Long: `Inspect the inference models offered by Qdrant Cloud.

Inference models turn text or images into dense, sparse or multi vectors directly
inside Qdrant Cloud, so a cluster can embed documents and queries without a
separate embedding service. Which models are offered depends on the cloud
provider and region a cluster runs in.`,
		Example: `# List the inference models available in a region
qcloud inference models list --cloud-provider aws --cloud-region eu-central-1`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(newModelsCommand(s))
	return cmd
}
