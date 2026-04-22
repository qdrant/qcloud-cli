package selfupgrade

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsHomebrewInstall(t *testing.T) {
	tests := []struct {
		name    string
		exePath string
		want    bool
	}{
		{
			name:    "Apple Silicon Homebrew cask",
			exePath: "/opt/homebrew/Caskroom/qcloud/0.23.0/qcloud",
			want:    true,
		},
		{
			name:    "Apple Silicon Homebrew formula",
			exePath: "/opt/homebrew/Cellar/qcloud/0.23.0/bin/qcloud",
			want:    true,
		},
		{
			name:    "Intel Mac Homebrew cask",
			exePath: "/usr/local/Caskroom/qcloud/0.23.0/qcloud",
			want:    true,
		},
		{
			name:    "Intel Mac Homebrew formula",
			exePath: "/usr/local/Cellar/qcloud/0.23.0/bin/qcloud",
			want:    true,
		},
		{
			name:    "direct download in /usr/local/bin",
			exePath: "/usr/local/bin/qcloud",
			want:    false,
		},
		{
			name:    "go install path",
			exePath: "/Users/me/go/bin/qcloud",
			want:    false,
		},
		{
			name:    "empty exe path",
			exePath: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsHomebrewInstall(tt.exePath)
			assert.Equal(t, tt.want, got)
		})
	}
}
