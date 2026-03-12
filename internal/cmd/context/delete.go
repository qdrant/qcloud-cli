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

			if _, ok := s.Config.GetContext(name); !ok {
				return fmt.Errorf("context %q not found", name)
			}

			force, _ := cmd.Flags().GetBool("force")
			if !util.ConfirmAction(force, fmt.Sprintf("Are you sure you want to delete context %q?", name)) {
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
