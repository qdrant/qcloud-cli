package util_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/qdrant/qcloud-cli/internal/cmd/util"
)

func TestIsUUID(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"00000000-0000-0000-0000-000000000000", true},
		{"not-a-uuid", false},
		{"550e8400e29b41d4a716446655440000", true}, // no dashes, still valid
		{"", false},
		{"my-package-name", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, util.IsUUID(tt.input))
		})
	}
}

func TestAnyFlagChanged(t *testing.T) {
	t.Run("returns true when one flag is changed", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("alpha", "", "")
		cmd.Flags().String("beta", "", "")
		_ = cmd.Flags().Set("alpha", "value")

		assert.True(t, util.AnyFlagChanged(cmd, []string{"alpha", "beta"}))
	})

	t.Run("returns false when no flags changed", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("alpha", "", "")
		cmd.Flags().String("beta", "", "")

		assert.False(t, util.AnyFlagChanged(cmd, []string{"alpha", "beta"}))
	})

	t.Run("returns false for empty flag list", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("alpha", "", "")
		_ = cmd.Flags().Set("alpha", "value")

		assert.False(t, util.AnyFlagChanged(cmd, nil))
	})
}
