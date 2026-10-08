package skills

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/agentskill"
	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newInstallCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Long: `Install the qcloud Agent Skill so AI coding agents can discover it.

By default the skill is installed into the current project: always into
.agents/skills (read by Codex, Cursor, Gemini CLI, GitHub Copilot and OpenCode),
and into .claude/skills when a .claude directory exists. Use --agent to pick
targets explicitly, --global to install into your home directory for every
project, or --dir to write to any skills directory.

Installing replaces any previous qcloud skill in the target, so re-run this
command after upgrading qcloud to keep the skill in sync with the binary.`,
		Example: `# Install into the current project for detected agents
qcloud skills install

# Install for Claude Code only, for all projects
qcloud skills install --agent claude --global

# Install into a custom skills directory
qcloud skills install --dir ./tools/skills`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "install",
				Short: "Install the qcloud Agent Skill for AI coding agents",
				Args:  cobra.NoArgs,
			}
			addTargetFlags(cmd)
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			dirs, err := skillsDirs(cmd, true)
			if err != nil {
				return err
			}

			files, err := agentskill.Generate(cmd.Root(), s.Version)
			if err != nil {
				return err
			}

			for _, d := range dirs {
				path, err := agentskill.Write(d, files)
				if err != nil {
					return err
				}

				fmt.Fprintf(cmd.OutOrStdout(), "Installed qcloud skill to %s\n", path)
			}

			return nil
		},
	}.CobraCommand(s)
}
