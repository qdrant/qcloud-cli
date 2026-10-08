package skills

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// NewCommand creates the "skills" parent command and registers all subcommands.
func NewCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skills",
		Short: "Manage the qcloud Agent Skill for AI coding agents",
		Long: `Manage the qcloud Agent Skill, which teaches AI coding agents how to use qcloud.

The skill follows the Agent Skills format (https://agentskills.io) and is read
by Claude Code, Codex, Cursor, Gemini CLI, GitHub Copilot, OpenCode and other
agents. It contains a guide to authentication, conventions and common
workflows, plus a reference for every command group generated from this
binary's help text, so it always matches the installed qcloud version.
Re-run "qcloud skills install" after upgrading qcloud to refresh it.`,
		Example: `# Install the skill into the current project for detected agents
qcloud skills install

# Install the skill for all your projects
qcloud skills install --global

# Print the skill to stdout
qcloud skills print`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(
		newPrintCommand(s),
		newInstallCommand(s),
		newUninstallCommand(s),
	)
	return cmd
}
