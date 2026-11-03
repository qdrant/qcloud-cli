package auth

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// NewCommand creates the auth command group.
func NewCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Log in to Qdrant Cloud with a browser",
		Long: `Log in to Qdrant Cloud with a browser (or a device code) instead of pasting an API key.

The login server is not configured separately: it follows from the API endpoint.
grpc.cloud.qdrant.io:443 logs in at login.cloud.qdrant.io. When the API
advertises its login server in its protected-resource metadata, that one is used.

After the login, commands send the access token instead of an API key, and
refresh it when it expires. An API key set with --api-key or QDRANT_CLOUD_API_KEY
still takes precedence over the login.`,
		Example: `# Show which login server the current endpoint uses
qcloud auth discover

# Log in with the browser
qcloud auth login

# Log in read-only on a machine without a browser
qcloud auth login --device --scope read-only

# Check whether you are logged in, and as whom
qcloud auth status`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(
		newDiscoverCommand(s),
		newLoginCommand(s),
		newLogoutCommand(s),
		newStatusCommand(s),
		newTokenCommand(s),
	)
	return cmd
}
