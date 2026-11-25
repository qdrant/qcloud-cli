package space

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	spacev1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newUpdateCommand(s *state.State) *cobra.Command {
	return base.UpdateCmd[*spacev1.Space]{
		Long: `Update an existing serverless space.

Use this command to rename a space or change its labels, network restrictions,
per-collection size limits, size warning threshold, search-worker settings, and
its storage and monthly spend limits. The cloud region cannot be changed after
creation.

Only the settings passed as flags are changed. Set --storage-limit or
--spend-limit to 'unlimited' to remove a limit. A limit's type or alert
threshold can be changed on its own once the limit is set. Lowering a limit
below current usage, switching a reached hard limit to soft, or removing it
applies to requests within about a minute.

Labels are merged with existing labels. Use 'key=value' to add or overwrite a
label, and 'key-' (with a trailing dash) to remove one. Allowed IPs and allowed
origins are merged the same way: specify a value to add it, or append '-' to
remove it (e.g. '10.0.0.0/8-').`,
		Example: `# Rename a space
qcloud serverless space update 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --name my-renamed-space

# Add a label and remove another one
qcloud serverless space update 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --label env=prod --label team-

# Allow a new origin and remove an old IP range
qcloud serverless space update 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --allowed-origin https://app.example.com --allowed-ip 10.0.0.0/8-

# Tune the collection limits and search workers
qcloud serverless space update 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --max-collection-indexed-size 2GiB --collection-size-warning-percent 90 \
  --searcher-max-workers 4

# Cap monthly spend at $49.99 and alert at 90%
qcloud serverless space update 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --spend-limit 49.99 --spend-limit-alert-percent 90

# Switch the storage limit to a soft limit, keeping its size
qcloud serverless space update 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --storage-limit-type soft

# Remove the monthly spend limit
qcloud serverless space update 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --spend-limit unlimited`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "update <space-id>",
				Short: "Update an existing space",
				Args:  util.ExactArgs(1, "a space ID"),
			}
			cmd.Flags().String("name", "", "New name of the space")
			addSharedSpaceFlags(cmd)
			return cmd
		},
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*spacev1.Space, error) {
			return getSpace(s, cmd, args[0])
		},
		Update: func(s *state.State, cmd *cobra.Command, space *spacev1.Space) (*spacev1.Space, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			updated := proto.CloneOf(space)

			if cmd.Flags().Changed("name") {
				updated.Name, _ = cmd.Flags().GetString("name")
			}

			if err := applySharedSpaceFlags(cmd, updated); err != nil {
				return nil, err
			}

			resp, err := client.ServerlessSpace().UpdateSpace(ctx, &spacev1.UpdateSpaceRequest{
				Space:      updated,
				UpdateMask: updateMask(cmd, spaceUpdatePaths),
			})
			if err != nil {
				return nil, fmt.Errorf("failed to update space: %w", err)
			}

			return resp.GetSpace(), nil
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, updated *spacev1.Space) {
			fmt.Fprintf(out, "Space %s (%s) updated successfully.\n", updated.GetId(), updated.GetName())
		},
		ValidArgsFunction: completion.SpaceIDCompletion(s),
	}.CobraCommand(s)
}
