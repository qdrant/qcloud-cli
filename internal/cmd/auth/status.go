package auth

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/oauth"
	"github.com/qdrant/qcloud-cli/internal/state"
)

// statusOutput is the login of the current endpoint, and the account and credential commands use.
type statusOutput struct {
	Context  string `json:"context,omitempty"`
	Endpoint string `json:"endpoint"`
	LoginURL string `json:"login_url"`
	LoggedIn bool   `json:"logged_in"`
	// Expired means the access token has expired; the next command refreshes it.
	Expired      bool       `json:"expired"`
	Email        string     `json:"email,omitempty"`
	GrantedScope string     `json:"granted_scope,omitempty"`
	Expiry       *time.Time `json:"expiry,omitempty"`
	AccountID    string     `json:"account_id,omitempty"`
	// AccountSource is where AccountID comes from: --account-id, QDRANT_CLOUD_ACCOUNT_ID, a context, or
	// the config file.
	AccountSource string `json:"account_source,omitempty"`
	// APIKeySet means an API key is configured, which commands use instead of the login.
	APIKeySet bool `json:"api_key_set"`
}

func newStatusCommand(s *state.State) *cobra.Command {
	return base.DescribeCmd[*statusOutput]{
		Use:   "status",
		Short: "Show OAuth login status for the current endpoint",
		Long: `Show whether you are logged in to the current endpoint, as whom, and with what access.

It also shows the account commands use and where it comes from, and the access
token's expiry. An expired access token is refreshed by the next command. The
token itself is never printed; use "qcloud auth token" for that.`,
		Example: `# Show the login of the current endpoint
qcloud auth status

# Check from a script whether you are logged in
qcloud auth status --json | jq .logged_in`,
		Args: cobra.NoArgs,
		Fetch: func(s *state.State, cmd *cobra.Command, _ []string) (*statusOutput, error) {
			disc, err := s.OAuth().Discover(cmd.Context(), s.Config.Endpoint())
			if err != nil {
				return nil, err
			}

			out := &statusOutput{
				Context:   s.Config.ActiveContext(),
				Endpoint:  disc.Endpoint,
				LoginURL:  disc.Issuer,
				APIKeySet: s.Config.APIKey() != "",
			}
			out.AccountID, out.AccountSource = s.Config.AccountIDSource()

			tok, ok, err := s.OAuth().Session(s.Config.Endpoint())
			if err != nil {
				return nil, err
			}

			if ok {
				out.LoggedIn = true
				out.Expired = tok.Expired(time.Now())
				out.Email = tok.Email
				out.GrantedScope = tok.GrantedScope
				if !tok.Expiry.IsZero() {
					out.Expiry = &tok.Expiry
				}
			}

			return out, nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, st *statusOutput) error {
			fmt.Fprintf(w, "Endpoint:  %s\n", st.Endpoint)
			if st.Context != "" {
				fmt.Fprintf(w, "Context:   %s\n", st.Context)
			}

			fmt.Fprintf(w, "Login URL: %s\n", st.LoginURL)
			if st.APIKeySet {
				fmt.Fprintln(w, "Note:      an API key is set; commands use it instead of this login")
			}

			if !st.LoggedIn {
				fmt.Fprintln(w, "Status:    logged out")
				return nil
			}

			status := "logged in"
			if st.Email != "" {
				status += " as " + st.Email
			}

			if st.Expired {
				status += " (access token expired; the next command refreshes it)"
			}

			fmt.Fprintf(w, "Status:    %s\n", status)
			if st.GrantedScope != "" {
				fmt.Fprintf(w, "Access:    %s\n", oauth.ScopeLabel(st.GrantedScope))
			}

			if st.AccountID != "" {
				fmt.Fprintf(w, "Account:   %s (from %s)\n", st.AccountID, st.AccountSource)
			} else {
				fmt.Fprintln(w, "Account:   none; select a context with \"qcloud context use <name>\" or pass --account-id")
			}

			if st.Expiry != nil {
				fmt.Fprintf(w, "Expiry:    %s  (%s)\n", output.HumanTime(*st.Expiry), output.FullDateTime(*st.Expiry))
			}

			return nil
		},
	}.CobraCommand(s)
}
