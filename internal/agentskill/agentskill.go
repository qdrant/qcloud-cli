// Package agentskill generates the qcloud Agent Skill (https://agentskills.io)
// from the command tree and installs it into agent skill directories.
package agentskill

import (
	"bytes"
	_ "embed"
	"fmt"
	"slices"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Name is the skill name and the name of its directory.
const Name = "qcloud"

//go:embed template/SKILL.md.tmpl
var skillTemplate string

var excludedCommands = []string{"help", "completion", "skills", "self-upgrade", "version"}

// Reference describes one generated references/<name>.md file.
type Reference struct {
	Name  string
	Short string
}

// Generate renders the skill from the given root command. The result maps
// paths relative to the skill directory to file contents.
func Generate(root *cobra.Command, version string) (map[string][]byte, error) {
	files := map[string][]byte{
		"references/global-flags.md": renderGlobalFlags(root),
	}

	var refs []Reference
	for _, group := range ReferenceGroups(root) {
		refs = append(refs, Reference{Name: group.Name(), Short: group.Short})
		files["references/"+group.Name()+".md"] = renderGroup(group)
	}

	tmpl, err := template.New("SKILL.md").Parse(skillTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse skill template: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {
		Version    string
		References []Reference
	}{Version: version, References: refs})
	if err != nil {
		return nil, fmt.Errorf("render skill template: %w", err)
	}

	files["SKILL.md"] = buf.Bytes()

	return files, nil
}

// ReferenceGroups returns the top-level commands that get a reference file.
func ReferenceGroups(root *cobra.Command) []*cobra.Command {
	var groups []*cobra.Command
	for _, c := range root.Commands() {
		if !c.IsAvailableCommand() || slices.Contains(excludedCommands, c.Name()) {
			continue
		}

		groups = append(groups, c)
	}

	return groups
}

func renderGlobalFlags(root *cobra.Command) []byte {
	var b strings.Builder
	b.WriteString("# Global flags\n\n")
	b.WriteString("These flags are accepted by every qcloud command. Flags override environment variables, which override the active context in the config file.\n\n")
	writeFlags(&b, root.PersistentFlags())
	return []byte(b.String())
}

func renderGroup(group *cobra.Command) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", group.CommandPath())
	writeDescription(&b, group)

	var commands []*cobra.Command
	walk(group, func(c *cobra.Command) {
		if c.Runnable() {
			commands = append(commands, c)
		}
	})

	if len(commands) > 1 {
		b.WriteString("Commands:\n\n")
		for _, c := range commands {
			fmt.Fprintf(&b, "- `%s`: %s\n", c.CommandPath(), c.Short)
		}

		b.WriteString("\n")
	}

	for _, c := range commands {
		fmt.Fprintf(&b, "## %s\n\n", c.CommandPath())
		if c != group {
			writeDescription(&b, c)
		}

		fmt.Fprintf(&b, "```\n%s\n```\n\n", c.UseLine())

		var flags strings.Builder
		writeFlags(&flags, c.NonInheritedFlags())
		if flags.Len() > 0 {
			b.WriteString("Flags:\n\n")
			b.WriteString(flags.String())
		}

		if c.Example != "" {
			fmt.Fprintf(&b, "Examples:\n\n```sh\n%s\n```\n\n", strings.TrimSpace(c.Example))
		}
	}

	return []byte(strings.TrimRight(b.String(), "\n") + "\n")
}

func walk(c *cobra.Command, fn func(*cobra.Command)) {
	fn(c)
	for _, sub := range c.Commands() {
		if sub.IsAvailableCommand() {
			walk(sub, fn)
		}
	}
}

func writeDescription(b *strings.Builder, c *cobra.Command) {
	desc := strings.TrimSpace(c.Long)
	if desc == "" {
		desc = c.Short
	}

	if desc != "" {
		fmt.Fprintf(b, "%s\n\n", desc)
	}
}

func writeFlags(b *strings.Builder, fs *pflag.FlagSet) {
	visible := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.VisitAll(func(f *pflag.Flag) {
		if f.Name != "help" && !f.Hidden {
			visible.AddFlag(f)
		}
	})
	if usages := visible.FlagUsages(); usages != "" {
		fmt.Fprintf(b, "```\n%s```\n\n", usages)
	}
}
