package selfupgrade

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/selfupgrade"
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
			cmd.Flags().String("version", "", "Upgrade to a specific version (e.g. 0.5.0)")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			check, _ := cmd.Flags().GetBool("check")
			force, _ := cmd.Flags().GetBool("force")
			targetVersion, _ := cmd.Flags().GetString("version")

			currentVersion := strings.TrimPrefix(s.Version, "v")
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

			var release *selfupgrade.ReleaseInfo
			var found bool

			if targetVersion != "" {
				targetVersion = strings.TrimPrefix(targetVersion, "v")
				release, found, err = updater.DetectVersion(ctx, targetVersion)
			} else {
				release, found, err = updater.DetectLatest(ctx)
			}
			if err != nil {
				return fmt.Errorf("failed to check for updates: %w", err)
			}

			if !found {
				if targetVersion != "" {
					return fmt.Errorf("version %s not found", targetVersion)
				}
				return fmt.Errorf("no releases found")
			}

			releaseVersion := release.Version
			if targetVersion == "" && releaseVersion == currentVersion && !isDev {
				fmt.Fprintf(out, "qcloud v%s is already up to date.\n", currentVersion)
				return nil
			}

			if check {
				fmt.Fprintf(out, "New version available: v%s (current: v%s)\n", releaseVersion, currentVersion)
				return nil
			}

			if isDev && !force {
				fmt.Fprintf(out, "Warning: you are running a dev build (v%s-dev). Use --force to upgrade.\n", currentVersion)
				return nil
			}

			action := "Upgrade"
			currentSemver, _ := semver.NewVersion(currentVersion)
			releaseSemver, _ := semver.NewVersion(releaseVersion)
			if currentSemver != nil && releaseSemver != nil && releaseSemver.LessThan(currentSemver) {
				action = "Downgrade"
			}

			prompt := fmt.Sprintf("%s qcloud from v%s to v%s?", action, currentVersion, releaseVersion)
			if !util.ConfirmAction(force, prompt) {
				fmt.Fprintln(out, "Aborted.")
				return nil
			}

			execPath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("failed to find executable path: %w", err)
			}

			fmt.Fprintf(out, "Downloading v%s...\n", releaseVersion)

			if err := updater.UpdateTo(ctx, releaseVersion, execPath); err != nil {
				if errors.Is(err, os.ErrPermission) {
					hint := "sudo qcloud self-upgrade"
					if runtime.GOOS == "windows" {
						hint = "running as Administrator"
					}
					return fmt.Errorf("permission denied: try %s", hint)
				}
				return fmt.Errorf("failed to update: %w", err)
			}

			fmt.Fprintf(out, "Successfully upgraded qcloud to v%s\n", releaseVersion)
			return nil
		},
	}.CobraCommand(s)
}
