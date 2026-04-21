package selfupgrade

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qdrant/qcloud-cli/internal/cmdexec"
)

func TestIsHomebrewInstall(t *testing.T) {
	tests := []struct {
		name    string
		result  *cmdexec.CommandResult
		runErr  error
		exePath string
		want    bool
	}{
		{
			name:    "Apple Silicon Homebrew cask",
			result:  &cmdexec.CommandResult{Stdout: []byte("/opt/homebrew\n")},
			exePath: "/opt/homebrew/Caskroom/qcloud/0.23.0/qcloud",
			want:    true,
		},
		{
			name:    "Intel Mac Homebrew cask",
			result:  &cmdexec.CommandResult{Stdout: []byte("/usr/local\n")},
			exePath: "/usr/local/Caskroom/qcloud/0.23.0/qcloud",
			want:    true,
		},
		{
			name:    "custom Homebrew prefix",
			result:  &cmdexec.CommandResult{Stdout: []byte("/Users/me/.homebrew\n")},
			exePath: "/Users/me/.homebrew/Caskroom/qcloud/1.0.0/qcloud",
			want:    true,
		},
		{
			name:    "direct download in /usr/local/bin",
			result:  &cmdexec.CommandResult{Stdout: []byte("/opt/homebrew\n")},
			exePath: "/usr/local/bin/qcloud",
			want:    false,
		},
		{
			name:    "go install path",
			result:  &cmdexec.CommandResult{Stdout: []byte("/opt/homebrew\n")},
			exePath: "/Users/me/go/bin/qcloud",
			want:    false,
		},
		{
			name:    "brew not installed",
			runErr:  fmt.Errorf("exec: \"brew\": executable file not found in $PATH"),
			exePath: "/usr/local/bin/qcloud",
			want:    false,
		},
		{
			name:    "brew exits non-zero",
			result:  &cmdexec.CommandResult{Stderr: []byte("error\n"), ExitCode: 1},
			exePath: "/usr/local/bin/qcloud",
			want:    false,
		},
		{
			name:    "empty exe path",
			result:  &cmdexec.CommandResult{Stdout: []byte("/opt/homebrew\n")},
			exePath: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := cmdexec.NewMockRunner()
			if tt.runErr != nil {
				runner.Respond([]string{"brew", "--prefix"}, nil, tt.runErr)
			} else {
				runner.Respond([]string{"brew", "--prefix"}, tt.result, nil)
			}

			got := IsHomebrewInstall(runner, tt.exePath)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsHomebrewInstall_CallsBrewPrefix(t *testing.T) {
	runner := cmdexec.NewMockRunner().
		Respond([]string{"brew"}, &cmdexec.CommandResult{Stdout: []byte("/opt/homebrew\n")}, nil)

	IsHomebrewInstall(runner, "/some/path/qcloud")

	require.Equal(t, 1, runner.CallCount())
	assert.Equal(t, []string{"brew", "--prefix"}, runner.Call(0))
}
