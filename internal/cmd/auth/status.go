package auth

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/oauth"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newStatusCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Long: `Show whether an OAuth session exists for the current endpoint.

Prints the inferred login URL, token expiry, and scope. Does not print the
access token.`,
		Example: `qcloud auth status`,
		BaseCobraCommand: func() *cobra.Command {
			return &cobra.Command{
				Use:   "status",
				Short: "Show OAuth login status for the current endpoint",
				Args:  cobra.NoArgs,
			}
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			disc, err := s.OAuthDiscoverer().Discover(cmd.Context(), s.Config.Endpoint())
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Endpoint:  %s\n", disc.Hosts.GRPCEndpoint)
			fmt.Fprintf(out, "Login URL: %s (%s)\n", disc.Issuer, disc.Source)

			path := oauth.CredentialsPath(s.Config.ConfigFilePath())
			tok, ok, err := oauth.LoadToken(path, s.Config.Endpoint())
			if err != nil {
				return err
			}

			if !ok {
				fmt.Fprintln(out, "Status:    logged out")
				return nil
			}

			status := "logged in"
			if tok.Expired(time.Now()) {
				status = "expired (will refresh on next API call)"
			}

			fmt.Fprintf(out, "Status:    %s\n", status)
			fmt.Fprintf(out, "Scope:     %s\n", tok.Scope)
			if !tok.Expiry.IsZero() {
				fmt.Fprintf(out, "Expiry:    %s\n", tok.Expiry.UTC().Format(time.RFC3339))
			}

			return nil
		},
	}.CobraCommand(s)
}
