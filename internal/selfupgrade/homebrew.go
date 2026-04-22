package selfupgrade

import (
	"os"
	"path/filepath"
	"strings"
)

// homebrewPrefixes lists well-known paths where Homebrew installs packages.
// /opt/homebrew/ is exclusive to Homebrew on Apple Silicon.
// /usr/local/Cellar/ and /usr/local/Caskroom/ are Homebrew-specific
// subdirectories on Intel Macs (the /usr/local/ prefix itself is shared).
var homebrewPrefixes = []string{
	"/opt/homebrew/",
	"/usr/local/Cellar/",
	"/usr/local/Caskroom/",
}

// IsHomebrewInstall reports whether the running binary was installed via
// Homebrew by checking if exePath lives under a known Homebrew prefix.
func IsHomebrewInstall(exePath string) bool {
	if exePath == "" {
		return false
	}

	for _, prefix := range homebrewPrefixes {
		if strings.HasPrefix(exePath, prefix) {
			return true
		}
	}
	return false
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
