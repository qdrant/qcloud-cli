package skills_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qdrant/qcloud-cli/internal/state"
	"github.com/qdrant/qcloud-cli/internal/testutil"
)

// newEnv creates a minimal TestEnv without a gRPC server — skills commands
// don't call the API.
func newEnv(t *testing.T) *testutil.TestEnv {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	return &testutil.TestEnv{
		State:   state.New("1.2.3"),
		Cleanup: func() {},
	}
}

func TestPrint_Skill(t *testing.T) {
	env := newEnv(t)

	stdout, _, err := testutil.Exec(t, env, "skills", "print")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(stdout, "---\nname: qcloud\n"))
	assert.Contains(t, stdout, `qcloud-version: "1.2.3"`)
	assert.Contains(t, stdout, "(references/cluster.md)")
}

func TestPrint_Reference(t *testing.T) {
	env := newEnv(t)

	stdout, _, err := testutil.Exec(t, env, "skills", "print", "--reference", "cluster")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(stdout, "# qcloud cluster\n"))
	assert.Contains(t, stdout, "## qcloud cluster create\n")
}

func TestPrint_UnknownReference(t *testing.T) {
	env := newEnv(t)

	_, _, err := testutil.Exec(t, env, "skills", "print", "--reference", "nope")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown reference "nope"`)
}

func TestInstall_Dir(t *testing.T) {
	env := newEnv(t)
	dir := t.TempDir()

	stdout, _, err := testutil.Exec(t, env, "skills", "install", "--dir", dir)
	require.NoError(t, err)
	assert.Contains(t, stdout, filepath.Join(dir, "qcloud"))
	assert.FileExists(t, filepath.Join(dir, "qcloud", "SKILL.md"))
	assert.FileExists(t, filepath.Join(dir, "qcloud", "references", "cluster.md"))
	assert.FileExists(t, filepath.Join(dir, "qcloud", "references", "global-flags.md"))
}

func TestInstall_ProjectDetectsAgents(t *testing.T) {
	env := newEnv(t)
	project := t.TempDir()
	t.Chdir(project)

	_, _, err := testutil.Exec(t, env, "skills", "install")
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(project, ".agents", "skills", "qcloud", "SKILL.md"))
	assert.NoDirExists(t, filepath.Join(project, ".claude"))

	require.NoError(t, os.Mkdir(filepath.Join(project, ".claude"), 0o755))
	stdout, _, err := testutil.Exec(t, env, "skills", "install")
	require.NoError(t, err)
	assert.Contains(t, stdout, filepath.Join(".agents", "skills", "qcloud"))
	assert.Contains(t, stdout, filepath.Join(".claude", "skills", "qcloud"))
	assert.FileExists(t, filepath.Join(project, ".claude", "skills", "qcloud", "SKILL.md"))
}

func TestInstall_GlobalAgent(t *testing.T) {
	env := newEnv(t)
	t.Chdir(t.TempDir())
	home := os.Getenv("HOME")

	_, _, err := testutil.Exec(t, env, "skills", "install", "--global", "--agent", "claude")
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(home, ".claude", "skills", "qcloud", "SKILL.md"))
	assert.NoDirExists(t, filepath.Join(home, ".agents"))
}

func TestInstall_UnknownAgent(t *testing.T) {
	env := newEnv(t)
	t.Chdir(t.TempDir())

	_, _, err := testutil.Exec(t, env, "skills", "install", "--agent", "bogus")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown agent "bogus"`)
}

func TestInstall_DirAndAgentAreExclusive(t *testing.T) {
	env := newEnv(t)

	_, _, err := testutil.Exec(t, env, "skills", "install", "--dir", t.TempDir(), "--agent", "claude")
	require.Error(t, err)
}

func TestUninstall_Dir(t *testing.T) {
	env := newEnv(t)
	dir := t.TempDir()

	_, _, err := testutil.Exec(t, env, "skills", "install", "--dir", dir)
	require.NoError(t, err)

	stdout, _, err := testutil.Exec(t, env, "skills", "uninstall", "--dir", dir, "--force")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Removed "+filepath.Join(dir, "qcloud"))
	assert.NoDirExists(t, filepath.Join(dir, "qcloud"))
}

func TestUninstall_NothingInstalled(t *testing.T) {
	env := newEnv(t)
	t.Chdir(t.TempDir())

	stdout, _, err := testutil.Exec(t, env, "skills", "uninstall", "--force")
	require.NoError(t, err)
	assert.Contains(t, stdout, "No qcloud skill installed.")
}

func TestUninstall_RefusesForeignDirectory(t *testing.T) {
	env := newEnv(t)
	dir := t.TempDir()
	foreign := filepath.Join(dir, "qcloud", "notes.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(foreign), 0o755))
	require.NoError(t, os.WriteFile(foreign, []byte("mine"), 0o644))

	_, _, err := testutil.Exec(t, env, "skills", "uninstall", "--dir", dir, "--force")
	require.Error(t, err)
	assert.FileExists(t, foreign)
}
