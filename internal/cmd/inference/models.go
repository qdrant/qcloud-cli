package inference

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

func newModelsCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "models",
		Short: "Manage inference models",
		Long: `Inspect the individual inference models offered by Qdrant Cloud.

A model is identified by its name (for example "cohere/*"), and describes the
vectors it produces: the vector type, the modality of the input it accepts, and
the dimensionality of its output.`,
		Example: `# List the inference models available in a region
qcloud inference models list --cloud-provider aws --cloud-region eu-central-1`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(newModelsListCommand(s))
	return cmd
}
