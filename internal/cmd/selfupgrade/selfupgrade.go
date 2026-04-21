package selfupgrade

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	upgrade "github.com/qdrant/qcloud-cli/internal/selfupgrade"
	"github.com/qdrant/qcloud-cli/internal/state"
)

// NewCommand creates the "self-upgrade" command.
func NewCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "self-upgrade",
				Short: "Upgrade qcloud to the latest version",
				Args:  cobra.NoArgs,
			}
			cmd.Flags().Bool("check", false, "Only check for a new version without installing")
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			check, _ := cmd.Flags().GetBool("check")
			force, _ := cmd.Flags().GetBool("force")

			isHomebrew := upgrade.IsHomebrewInstall(s.CmdRunner, upgrade.ResolveExecutablePath())

			if isHomebrew && !check {
				return fmt.Errorf("this installation is managed by Homebrew; use \"brew upgrade qcloud\" instead")
			}

			currentVersion := s.Version
			isDev := strings.Contains(currentVersion, "-dev")
			if isDev {
				currentVersion = strings.SplitN(currentVersion, "-", 2)[0]
			}

			updater, err := s.Updater()
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Checking for updates...")

			release, found, err := updater.DetectLatest(ctx)
			if err != nil {
				return fmt.Errorf("failed to check for updates: %w", err)
			}

			if !found {
				return fmt.Errorf("no releases found")
			}

			if release.Equal(currentVersion) && !isDev {
				fmt.Fprintf(out, "qcloud %s is already up to date.\n", currentVersion)
				return nil
			}

			if check {
				fmt.Fprintf(out, "New version available: %s (current: %s)\n", release.Version(), currentVersion)
				if isHomebrew {
					fmt.Fprintln(out, "This installation is managed by Homebrew; use \"brew upgrade qcloud\" instead.")
				}
				return nil
			}

			if isDev && !force {
				fmt.Fprintf(out, "Warning: you are running a dev build (%s-dev). Use --force to upgrade.\n", currentVersion)
				return nil
			}

			prompt := fmt.Sprintf("Upgrade qcloud from %s to %s?", currentVersion, release.Version())
			if !util.ConfirmAction(force, cmd.ErrOrStderr(), prompt) {
				fmt.Fprintln(out, "Aborted.")
				return nil
			}

			fmt.Fprintf(out, "Downloading %s...\n", release.Version())

			rel, err := updater.UpdateSelf(ctx, currentVersion)
			if err != nil {
				if errors.Is(err, os.ErrPermission) {
					hint := "sudo qcloud self-upgrade"
					if runtime.GOOS == "windows" {
						hint = "running as Administrator"
					}
					return fmt.Errorf("permission denied: try %s", hint)
				}
				return fmt.Errorf("failed to update: %w", err)
			}

			fmt.Fprintf(out, "Successfully upgraded qcloud to %s\n", rel.Version())
			return nil
		},
	}.CobraCommand(s)
}
