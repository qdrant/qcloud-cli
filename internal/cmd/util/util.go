package util

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

// ConfirmAction prompts the user for confirmation. Returns true if they confirm.
// If force is true, skips the prompt and returns true.
// The prompt is written to w.
func ConfirmAction(force bool, w io.Writer, prompt string) bool {
	if force {
		return true
	}
	fmt.Fprintf(w, "%s [y/N]: ", prompt)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}

// AnyFlagChanged reports whether any flag in the given list was set by the user.
func AnyFlagChanged(cmd *cobra.Command, flags []string) bool {
	return slices.ContainsFunc(flags, func(f string) bool {
		return cmd.Flags().Changed(f)
	})
}

// ExactArgs returns a PositionalArgs that requires exactly n args with a descriptive error.
func ExactArgs(n int, usage string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n {
			return fmt.Errorf("requires %s\n\nUsage: %s", usage, cmd.UseLine())
		}
		return nil
	}
}
