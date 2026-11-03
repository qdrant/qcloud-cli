package auth

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/oauth"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newDiscoverCommand(s *state.State) *cobra.Command {
	return base.DescribeCmd[oauth.Discovery]{
		Use:   "discover",
		Short: "Show the OAuth login URL inferred from the API endpoint",
		Long: `Show the OAuth API resource and login server for the current endpoint.

The gRPC endpoint (default grpc.cloud.qdrant.io:443) is turned into the API URL
https://api.cloud.qdrant.io, whose protected-resource metadata
(/.well-known/oauth-protected-resource) is read without credentials. When it
names a login server, that one is used; otherwise login.cloud.qdrant.io is
inferred from the endpoint, and the output says why the metadata was not used.
Nothing is stored and no login is needed.`,
		Example: `# Show the login server of the current endpoint
qcloud auth discover

# Show it for another endpoint
qcloud auth discover --endpoint grpc.cloud.qdrant.io:443

# Read the login server from a script
qcloud auth discover --json | jq -r .issuer`,
		Args: cobra.NoArgs,
		Fetch: func(s *state.State, cmd *cobra.Command, _ []string) (oauth.Discovery, error) {
			return s.OAuth().Discover(cmd.Context(), s.Config.Endpoint())
		},
		PrintText: func(_ *cobra.Command, out io.Writer, disc oauth.Discovery) error {
			fmt.Fprintf(out, "Endpoint:     %s\n", disc.Endpoint)
			fmt.Fprintf(out, "API resource: %s\n", disc.Resource)
			fmt.Fprintf(out, "Login URL:    %s\n", disc.Issuer)
			fmt.Fprintf(out, "Source:       %s\n", disc.Source)
			if disc.MetadataError != "" {
				fmt.Fprintf(out, "Metadata:     not used (%s)\n", disc.MetadataError)
			}

			if len(disc.Scopes) > 0 {
				fmt.Fprintf(out, "Scopes:       %s\n", strings.Join(disc.Scopes, ", "))
			}

			return nil
		},
	}.CobraCommand(s)
}
