package auth

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newLogoutCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Long: `Log out of the current endpoint: remove its stored session and the contexts that use it.

The refresh token is revoked at the login server first. Logging out still
succeeds when that is not possible (the keychain entry is gone, or the login
server cannot be reached); qcloud then warns and removes the session anyway.
The access you granted stays listed under "Connected apps" in the console, where
you revoke it. Contexts with an API key are not changed.`,
		Example: `# Log out of the current endpoint
qcloud auth logout

# Log out of another endpoint
qcloud auth logout --endpoint grpc.cloud.qdrant.io:443`,
		BaseCobraCommand: func() *cobra.Command {
			return &cobra.Command{
				Use:   "logout",
				Short: "Log out and drop stored OAuth tokens",
				Args:  cobra.NoArgs,
			}
		},
		Run: func(s *state.State, cmd *cobra.Command, _ []string) error {
			endpoint := s.Config.Endpoint()
			warnings, err := s.OAuth().Logout(cmd.Context(), endpoint)
			for _, w := range warnings {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %v\n", w)
			}

			if err != nil {
				return err
			}

			// The contexts the login created use this login; without it they cannot work.
			removed, err := removeOAuthContexts(s, endpoint)
			if err != nil {
				return err
			}

			if len(removed) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Logged out; removed contexts %s.\n", strings.Join(removed, ", "))
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Logged out.")
			return nil
		},
	}.CobraCommand(s)
}
