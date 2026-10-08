package agentskill_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/qdrant/qcloud-cli/internal/agentskill"
	"github.com/qdrant/qcloud-cli/internal/cli"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func generate(t *testing.T) map[string][]byte {
	t.Helper()
	files, err := agentskill.Generate(cli.NewRootCommand(state.New("1.2.3")), "1.2.3")
	require.NoError(t, err)
	return files
}

func TestGenerate_Frontmatter(t *testing.T) {
	skill := string(generate(t)["SKILL.md"])

	parts := strings.SplitN(skill, "---\n", 3)
	require.Len(t, parts, 3, "SKILL.md must start with YAML frontmatter")
	require.Empty(t, parts[0])

	var fm struct {
		Name        string            `yaml:"name"`
		Description string            `yaml:"description"`
		Metadata    map[string]string `yaml:"metadata"`
	}
	require.NoError(t, yaml.Unmarshal([]byte(parts[1]), &fm))

	assert.Equal(t, agentskill.Name, fm.Name)
	assert.NotEmpty(t, fm.Description)
	assert.LessOrEqual(t, len(fm.Description), 1024)
	assert.Equal(t, "1.2.3", fm.Metadata["qcloud-version"])
	assert.Less(t, strings.Count(skill, "\n"), 500)
}

func TestGenerate_References(t *testing.T) {
	files := generate(t)
	skill := string(files["SKILL.md"])

	for _, group := range []string{"cluster", "backup", "hybrid", "serverless", "iam", "context"} {
		path := "references/" + group + ".md"
		assert.Contains(t, files, path)
		assert.Contains(t, skill, "("+path+")", "SKILL.md should link %s", path)
	}

	assert.Contains(t, files, "references/global-flags.md")

	for _, excluded := range []string{"skills", "self-upgrade", "version", "help", "completion"} {
		assert.NotContains(t, files, "references/"+excluded+".md")
	}
}

func TestGenerate_GroupContent(t *testing.T) {
	cluster := string(generate(t)["references/cluster.md"])

	assert.Contains(t, cluster, "## qcloud cluster create\n")
	assert.Contains(t, cluster, "## qcloud cluster key create\n")
	assert.Contains(t, cluster, "qcloud cluster key create <cluster-id> [flags]")
	assert.Contains(t, cluster, "--cloud-provider string")
	assert.Contains(t, cluster, "# Create a free-tier cluster")
	assert.NotContains(t, cluster, "--poll-interval", "hidden flags must be omitted")
	assert.NotContains(t, cluster, "--help")
	assert.NotContains(t, cluster, "--api-key string", "global flags belong in global-flags.md")
}

func TestGenerate_GlobalFlags(t *testing.T) {
	global := string(generate(t)["references/global-flags.md"])

	assert.Contains(t, global, "--api-key string")
	assert.Contains(t, global, "--json")
	assert.Contains(t, global, "QDRANT_CLOUD_ACCOUNT_ID")
}

func TestWrite_ReplacesPreviousInstall(t *testing.T) {
	dir := t.TempDir()
	files := generate(t)

	path, err := agentskill.Write(dir, files)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "qcloud"), path)

	stale := filepath.Join(path, "references", "removed-group.md")
	require.NoError(t, os.WriteFile(stale, []byte("stale"), 0o644))

	_, err = agentskill.Write(dir, files)
	require.NoError(t, err)
	assert.NoFileExists(t, stale)
	assert.FileExists(t, filepath.Join(path, "SKILL.md"))
	assert.FileExists(t, filepath.Join(path, "references", "cluster.md"))
}

func TestWrite_RefusesForeignDirectory(t *testing.T) {
	dir := t.TempDir()
	foreign := filepath.Join(dir, "qcloud", "notes.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(foreign), 0o755))
	require.NoError(t, os.WriteFile(foreign, []byte("mine"), 0o644))

	_, err := agentskill.Write(dir, generate(t))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "refusing")
	assert.FileExists(t, foreign)

	require.Error(t, agentskill.Remove(dir))
	assert.FileExists(t, foreign)
}

func TestSkillsDirs(t *testing.T) {
	base := t.TempDir()

	dirs, err := agentskill.SkillsDirs(base, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(base, ".agents", "skills")}, dirs)

	require.NoError(t, os.Mkdir(filepath.Join(base, ".claude"), 0o755))
	dirs, err = agentskill.SkillsDirs(base, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{
		filepath.Join(base, ".agents", "skills"),
		filepath.Join(base, ".claude", "skills"),
	}, dirs)

	dirs, err = agentskill.SkillsDirs(base, []string{"claude"})
	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(base, ".claude", "skills")}, dirs)

	_, err = agentskill.SkillsDirs(base, []string{"bogus"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bogus")
}
