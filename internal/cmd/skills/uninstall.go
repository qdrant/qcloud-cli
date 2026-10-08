package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/agentskill"
	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newUninstallCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Long: `Remove the qcloud Agent Skill from agent skills directories.

By default the skill is removed from every supported agent directory in the
current project. Use --agent, --global and --dir to choose targets the same
way as "qcloud skills install". Only directories containing the qcloud skill
are removed.`,
		Example: `# Remove the skill from the current project
qcloud skills uninstall

# Remove the user-level skill without confirmation
qcloud skills uninstall --global --force`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "uninstall",
				Short: "Remove the qcloud Agent Skill",
				Args:  cobra.NoArgs,
			}
			addTargetFlags(cmd)
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			dirs, err := skillsDirs(cmd, false)
			if err != nil {
				return err
			}

			var installed []string
			for _, d := range dirs {
				if _, err := os.Stat(filepath.Join(d, agentskill.Name)); err == nil {
					installed = append(installed, d)
				}
			}

			if len(installed) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No qcloud skill installed.")
				return nil
			}

			paths := make([]string, len(installed))
			for i, d := range installed {
				paths[i] = filepath.Join(d, agentskill.Name)
			}

			force, _ := cmd.Flags().GetBool("force")
			if !util.ConfirmAction(force, cmd.ErrOrStderr(), fmt.Sprintf("Remove the qcloud skill from %s?", strings.Join(paths, ", "))) {
				fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
				return nil
			}

			for i, d := range installed {
				if err := agentskill.Remove(d); err != nil {
					return err
				}

				fmt.Fprintf(cmd.OutOrStdout(), "Removed %s\n", paths[i])
			}

			return nil
		},
	}.CobraCommand(s)
}
