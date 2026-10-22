package auth

import (
	"fmt"
	"os"
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
			if name := s.Config.ActiveContext(); name != "" {
				fmt.Fprintf(out, "Context:   %s\n", name)
			}

			fmt.Fprintf(out, "Login URL: %s (%s)\n", disc.Issuer, disc.Source)
			if s.Config.APIKey() != "" {
				fmt.Fprintln(out, "Note:      an API key is set; commands use it instead of this login")
			}

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
			// The access the user granted on the consent page. The requested scope (what qcloud asked for)
			// is in credentials.yaml; showing it here next to this only confuses.
			if tok.GrantedScope != "" {
				fmt.Fprintf(out, "Access:    %s\n", oauth.ScopeLabel(tok.GrantedScope))
			}
			if id, source := accountSource(cmd, s, tok); id != "" {
				fmt.Fprintf(out, "Account:   %s (from %s)\n", id, source)
			} else {
				fmt.Fprintln(out, "Account:   none; pass --account-id")
			}

			if !tok.Expiry.IsZero() {
				fmt.Fprintf(out, "Expiry:    %s\n", tok.Expiry.UTC().Format(time.RFC3339))
			}

			return nil
		},
	}.CobraCommand(s)
}

// accountSource returns the account commands will use, and where it comes from, in the order the
// config resolves it: the --account-id flag, QDRANT_CLOUD_ACCOUNT_ID, the active context, and last the
// account the login was granted.
func accountSource(cmd *cobra.Command, s *state.State, tok oauth.Token) (string, string) {
	if f := cmd.Flag("account-id"); f != nil && f.Changed {
		return s.Config.AccountID(), "--account-id"
	}

	if id := os.Getenv("QDRANT_CLOUD_ACCOUNT_ID"); id != "" {
		return id, "QDRANT_CLOUD_ACCOUNT_ID"
	}

	if name := s.Config.ActiveContext(); name != "" {
		if entry, ok := s.Config.GetContext(name); ok && entry.AccountID != "" {
			return entry.AccountID, "context " + name
		}
	}

	if tok.AccountID != "" && s.Config.APIKey() == "" {
		return tok.AccountID, "the login"
	}

	return "", ""
}
