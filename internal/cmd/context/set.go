package context

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newSetCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "set <name>",
				Short: "Create or update a context",
				Args:  util.ExactArgs(1, "a context name"),
			}
			cmd.Flags().String("endpoint", "", "API endpoint for this context")
			cmd.Flags().String("api-key", "", "API key for this context")
			cmd.Flags().String("account-id", "", "Account ID for this context")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			name := args[0]

			// Load existing context values or start with an empty entry.
			ctx, _ := s.Config.GetContext(name)
			ctx.Name = name

			// Apply only flags that were explicitly provided.
			if ep, ok := changedString(cmd, "endpoint"); ok {
				ctx.Endpoint = ep
			}
			if key, ok := changedString(cmd, "api-key"); ok {
				ctx.APIKey = key
			}
			if id, ok := changedString(cmd, "account-id"); ok {
				ctx.AccountID = id
			}

			s.Config.UpsertContext(ctx)
			activated := false
			if s.Config.CurrentContext() == "" {
				s.Config.SetCurrentContext(name)
				activated = true
			}
			if err := s.Config.WriteToFile(); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Context %q saved.\n", name)
			if activated {
				fmt.Fprintf(cmd.OutOrStdout(), "Switched to context %q.\n", name)
			}
			return nil
		},
	}.CobraCommand(s)
}

func changedString(cmd *cobra.Command, name string) (string, bool) {
	if !cmd.Flags().Changed(name) {
		return "", false
	}
	v, _ := cmd.Flags().GetString(name)
	return v, true
}
