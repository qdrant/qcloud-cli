package space

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

func newKeyCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key",
		Short: "Manage API keys for a space",
		Long: `Manage API keys for a serverless space.

Space API keys authenticate requests to the space endpoint. A key either grants
global access to the whole space (manage, read-only or metrics-read-only) or
per-collection access (read-only or read-write) to a set of named collections.
The secret value of a key is only returned once, when the key is created.`,
		Example: `# List API keys for a space
qcloud serverless space key list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Create an API key with manage access
qcloud serverless space key create 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --name my-key

# Delete an API key
qcloud serverless space key delete 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 a1b2c3d4-e5f6-7890-abcd-ef1234567890`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(
		newKeyListCommand(s),
		newKeyCreateCommand(s),
		newKeyDeleteCommand(s),
	)
	return cmd
}
