package skills

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/agentskill"
)

func addTargetFlags(cmd *cobra.Command) {
	cmd.Flags().StringArray("agent", nil, "Agent to target ("+strings.Join(agentskill.AgentNames(), ", ")+"); can be specified multiple times")
	cmd.Flags().Bool("global", false, "Target the user-level skills directories in your home directory instead of the current project")
	cmd.Flags().String("dir", "", "Skills directory to target; the skill lives in <dir>/qcloud (ignores --agent and --global)")
	cmd.MarkFlagsMutuallyExclusive("dir", "agent")
	cmd.MarkFlagsMutuallyExclusive("dir", "global")
	_ = cmd.MarkFlagDirname("dir")
	_ = cmd.RegisterFlagCompletionFunc("agent", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return agentskill.AgentNames(), cobra.ShellCompDirectiveNoFileComp
	})
}

// skillsDirs resolves the target flags to skills directories. With no --agent,
// detect controls whether only detected agents are returned or all of them.
func skillsDirs(cmd *cobra.Command, detect bool) ([]string, error) {
	if dir, _ := cmd.Flags().GetString("dir"); dir != "" {
		return []string{dir}, nil
	}

	agents, _ := cmd.Flags().GetStringArray("agent")
	if len(agents) == 0 && !detect {
		agents = agentskill.AgentNames()
	}

	global, _ := cmd.Flags().GetBool("global")
	base, err := os.Getwd()
	if global {
		base, err = os.UserHomeDir()
	}

	if err != nil {
		return nil, err
	}

	dirs, err := agentskill.SkillsDirs(base, agents)
	if err != nil {
		return nil, err
	}

	for i, d := range dirs {
		if rel, err := filepath.Rel(base, d); err == nil && !global {
			dirs[i] = rel
		}
	}

	return dirs, nil
}
