package context

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newDeleteCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Example: `# Delete a context
qcloud context delete staging

# Delete without confirmation
qcloud context delete staging --force`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "delete <name>",
				Short: "Delete a context",
				Args:  util.ExactArgs(1, "a context name"),
			}
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			name := args[0]

			ctx, ok := s.Config.GetContext(name)
			if !ok {
				return fmt.Errorf("context %q not found", name)
			}

			// OAuth contexts follow the login: "qcloud auth login" creates and updates them, "qcloud auth
			// logout" removes them.
			if ctx.IsOAuth() {
				return fmt.Errorf("context %q belongs to your OAuth login; run \"qcloud auth logout --endpoint %s\" to remove it", name, ctx.Endpoint)
			}

			force, _ := cmd.Flags().GetBool("force")
			if !util.ConfirmAction(force, cmd.ErrOrStderr(), fmt.Sprintf("Are you sure you want to delete context %q?", name)) {
				fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
				return nil
			}

			s.Config.DeleteContext(name)
			if err := s.Config.WriteToFile(); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Context %q deleted.\n", name)
			return nil
		},
		ValidArgsFunction: contextNameCompletion(s),
	}.CobraCommand(s)
}
