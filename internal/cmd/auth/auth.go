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

The login host is not a separate flag. It is derived from the gRPC
endpoint (flag --endpoint, env QDRANT_CLOUD_ENDPOINT, or the active context):

  grpc.cloud.qdrant.io:443  →  login.cloud.qdrant.io

When OAuth is enabled, the gateway also advertises the issuer at the
unauthenticated URL https://api.cloud.qdrant.io/.well-known/oauth-protected-resource.
That document wins over hostname inference.

After login, API calls send Authorization: Bearer <token>. Management API keys
keep working unchanged when --api-key / QDRANT_CLOUD_API_KEY is set.`,
		Example: `# Discover the login URL for the current endpoint
qcloud auth discover

# Browser login (authorization code + PKCE on 127.0.0.1)
qcloud auth login --client-id <native-app>

# Device code (SSH / no browser)
qcloud auth login --device --scope read-only`,
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
