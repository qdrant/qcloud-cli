package skills

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/agentskill"
	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newPrintCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Long: `Print the qcloud Agent Skill to stdout without installing it.

Without flags, prints SKILL.md, the entry point an agent reads first. Use
--reference to print the generated reference for one command group, or
"global-flags" for the flags shared by every command. This is useful for
agents that do not support skill directories, or to inspect what
"qcloud skills install" would write.`,
		Example: `# Print SKILL.md
qcloud skills print

# Print the reference for the cluster command group
qcloud skills print --reference cluster

# Print the global flags reference
qcloud skills print --reference global-flags`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "print",
				Short: "Print the qcloud Agent Skill to stdout",
				Args:  cobra.NoArgs,
			}
			cmd.Flags().String("reference", "", "Print the reference for a command group instead of SKILL.md")
			_ = cmd.RegisterFlagCompletionFunc("reference", func(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
				names := []string{"global-flags"}
				for _, g := range agentskill.ReferenceGroups(cmd.Root()) {
					names = append(names, g.Name())
				}

				return names, cobra.ShellCompDirectiveNoFileComp
			})
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			files, err := agentskill.Generate(cmd.Root(), s.Version)
			if err != nil {
				return err
			}

			path := "SKILL.md"
			ref, _ := cmd.Flags().GetString("reference")
			if ref != "" {
				path = "references/" + ref + ".md"
			}

			content, ok := files[path]
			if !ok {
				return fmt.Errorf("unknown reference %q", ref)
			}

			_, err = cmd.OutOrStdout().Write(content)
			return err
		},
	}.CobraCommand(s)
}
