package selfupgrade

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/qdrant/qcloud-cli/internal/cmdexec"
)

// IsHomebrewInstall reports whether the running binary was installed via
// Homebrew by checking if exePath lives under <brew --prefix>/Caskroom/.
// The runner is used to execute "brew --prefix".
func IsHomebrewInstall(runner cmdexec.CommandRunner, exePath string) bool {
	if exePath == "" {
		return false
	}

	result, err := runner.Run("brew", "--prefix")
	if err != nil || result.ExitCode != 0 {
		return false
	}

	prefix := strings.TrimSpace(string(result.Stdout))
	if prefix == "" {
		return false
	}

	caskroomDir := filepath.Join(prefix, "Caskroom") + string(filepath.Separator)
	return strings.HasPrefix(exePath, caskroomDir)
}

// ResolveExecutablePath returns the real path of the running binary,
// following symlinks. Returns an empty string on any error.
func ResolveExecutablePath() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return ""
	}
	return resolved
}
