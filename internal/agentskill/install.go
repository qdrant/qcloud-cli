package agentskill

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Agent is a coding agent that discovers skills from a known directory.
type Agent struct {
	Name string
	// SkillsDir is relative to the project root, or to the home directory for
	// a global install.
	SkillsDir string
	// Marker is a path whose presence means the agent is in use. Agents
	// without one are always installed to.
	Marker string
}

// Agents lists the supported install targets. The shared .agents/skills
// directory is read by Codex, Cursor, Gemini CLI, GitHub Copilot and OpenCode.
var Agents = []Agent{
	{Name: "agents", SkillsDir: filepath.Join(".agents", "skills")},
	{Name: "claude", SkillsDir: filepath.Join(".claude", "skills"), Marker: ".claude"},
}

// AgentNames returns the names of all supported agents.
func AgentNames() []string {
	names := make([]string, 0, len(Agents))
	for _, a := range Agents {
		names = append(names, a.Name)
	}

	return names
}

// SkillsDirs returns the skills directories for the named agents under base.
// With no names, it returns the directories of agents detected under base.
func SkillsDirs(base string, names []string) ([]string, error) {
	for _, n := range names {
		if !slices.Contains(AgentNames(), n) {
			return nil, fmt.Errorf("unknown agent %q (supported: %s)", n, strings.Join(AgentNames(), ", "))
		}
	}

	var dirs []string
	for _, a := range Agents {
		if len(names) > 0 {
			if !slices.Contains(names, a.Name) {
				continue
			}
		} else if a.Marker != "" {
			if _, err := os.Stat(filepath.Join(base, a.Marker)); err != nil {
				continue
			}
		}

		dirs = append(dirs, filepath.Join(base, a.SkillsDir))
	}

	return dirs, nil
}

// Write installs the skill into <skillsDir>/qcloud, replacing a previous
// install so files from removed command groups don't linger. It returns the
// skill directory.
func Write(skillsDir string, files map[string][]byte) (string, error) {
	dir := filepath.Join(skillsDir, Name)
	if err := remove(dir); err != nil {
		return "", err
	}

	for rel, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return "", err
		}

		if err := os.WriteFile(path, content, 0o644); err != nil {
			return "", err
		}
	}

	return dir, nil
}

// Remove deletes <skillsDir>/qcloud if it exists.
func Remove(skillsDir string) error {
	return remove(filepath.Join(skillsDir, Name))
}

// remove deletes dir only if it holds a skill named qcloud, so a directory the
// user created for something else is never wiped.
func remove(dir string) error {
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if !isQcloudSkill(filepath.Join(dir, "SKILL.md")) {
		return fmt.Errorf("refusing to replace %s: it does not contain a qcloud SKILL.md", dir)
	}

	return os.RemoveAll(dir)
}

func isQcloudSkill(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() || scanner.Text() != "---" {
		return false
	}

	for scanner.Scan() {
		switch scanner.Text() {
		case "---":
			return false
		case "name: " + Name:
			return true
		}
	}

	return false
}
