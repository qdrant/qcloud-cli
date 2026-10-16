package auth

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newDiscoverCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Long: `Print the OAuth API resource and authorization-server issuer for the current endpoint.

The gRPC endpoint (default grpc.cloud.qdrant.io:443) is turned into
https://api.cloud.qdrant.io. The CLI then GETs
/.well-known/oauth-protected-resource with no credentials. If the gateway
returns authorization_servers, that issuer is used; otherwise
login.cloud.qdrant.io is inferred. No client id is required.`,
		Example: `qcloud auth discover`,
		BaseCobraCommand: func() *cobra.Command {
			return &cobra.Command{
				Use:   "discover",
				Short: "Show the OAuth login URL inferred from the API endpoint",
				Args:  cobra.NoArgs,
			}
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			disc, err := s.OAuthDiscoverer().Discover(cmd.Context(), s.Config.Endpoint())
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Endpoint:     %s\n", disc.Hosts.GRPCEndpoint)
			fmt.Fprintf(out, "API resource: %s\n", disc.Resource)
			fmt.Fprintf(out, "Login URL:    %s\n", disc.Issuer)
			fmt.Fprintf(out, "Source:       %s\n", disc.Source)
			if len(disc.Scopes) > 0 {
				fmt.Fprintf(out, "Scopes:       %s\n", strings.Join(disc.Scopes, ", "))
			}

			return nil
		},
	}.CobraCommand(s)
}
