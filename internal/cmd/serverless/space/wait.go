package space

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newWaitCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Long: `Wait for a serverless space to become ready.

Polls the space until its phase is READY and prints the endpoint. The command
fails as soon as the space is DISABLED or DELETING, printing the reason reported
by the server, or when the timeout expires.`,
		Example: `# Wait for a space to become ready
qcloud serverless space wait 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Wait with a custom timeout
qcloud serverless space wait 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --timeout 20m`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "wait <space-id>",
				Short: "Wait for a space to become ready",
				Args:  util.ExactArgs(1, "a space ID"),
			}
			cmd.Flags().Duration("timeout", 10*time.Minute, "Maximum time to wait for the space to become ready")
			cmd.Flags().Duration("poll-interval", 5*time.Second, "How often to poll the space status")
			_ = cmd.Flags().MarkHidden("poll-interval")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return err
			}

			timeout, _ := cmd.Flags().GetDuration("timeout")
			pollInterval, _ := cmd.Flags().GetDuration("poll-interval")
			space, err := waitForSpaceReady(ctx, client.ServerlessSpace(), cmd.ErrOrStderr(), accountID, args[0], timeout, pollInterval)
			if err != nil {
				return err
			}

			fmt.Fprint(cmd.OutOrStdout(), spaceResultMessage(space, "is ready"))
			return nil
		},
		ValidArgsFunction: completion.SpaceIDCompletion(s),
	}.CobraCommand(s)
}
