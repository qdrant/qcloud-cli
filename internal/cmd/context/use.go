package context

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newUseCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Example: `# Switch to another context
qcloud context use production`,
		BaseCobraCommand: func() *cobra.Command {
			return &cobra.Command{
				Use:   "use <name>",
				Short: "Set the active context",
				Args:  util.ExactArgs(1, "a context name"),
			}
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			name := args[0]

			if _, ok := s.Config.GetContext(name); !ok {
				return fmt.Errorf("context %q not found", name)
			}

			s.Config.SetCurrentContext(name)
			if err := s.Config.WriteToFile(); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Switched to context %q.\n", name)
			return nil
		},
		ValidArgsFunction: contextNameCompletion(s),
	}.CobraCommand(s)
}
