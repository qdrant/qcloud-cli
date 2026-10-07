package cluster

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"

	monitoringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/monitoring/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newLogsCommand(s *state.State) *cobra.Command {
	cmd := base.DescribeCmd[*monitoringv1.GetClusterLogsResponse]{
		Use:   "logs <cluster-id>",
		Short: "Retrieve logs for a cluster",
		Args:  util.ExactArgs(1, "a cluster ID"),
		Long: `Retrieve logs for a cluster.

By default, logs from the last 3 days up to now are returned. --since and --until
accept an RFC3339 timestamp, a YYYY-MM-DD date in UTC, or a duration ago such as
6h or 7d. A date passed to --since starts at the beginning of that day, and a
date passed to --until includes the whole day.`,
		Example: `# Get logs for a cluster
qcloud cluster logs abc-123

# Get logs since a specific date
qcloud cluster logs abc-123 --since 2024-01-01

# Get logs from the last 6 hours
qcloud cluster logs abc-123 --since 6h

# Get logs in a specific time range
qcloud cluster logs abc-123 --since 2024-01-01T00:00:00Z --until 2024-01-02T00:00:00Z

# Get logs in JSON format
qcloud cluster logs abc-123 --json`,
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*monitoringv1.GetClusterLogsResponse, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			req := &monitoringv1.GetClusterLogsRequest{
				AccountId: accountID,
				ClusterId: args[0],
			}

			if cmd.Flags().Changed("since") {
				sinceStr, _ := cmd.Flags().GetString("since")
				t, err := util.ParseTimeFlag(sinceStr, time.Now())
				if err != nil {
					return nil, fmt.Errorf("invalid --since %q: %w", sinceStr, err)
				}

				req.Since = timestamppb.New(t)
			}

			if cmd.Flags().Changed("until") {
				untilStr, _ := cmd.Flags().GetString("until")
				t, err := util.ParseUntilFlag(untilStr, time.Now())
				if err != nil {
					return nil, fmt.Errorf("invalid --until %q: %w", untilStr, err)
				}

				req.Until = timestamppb.New(t)
			}

			resp, err := client.Monitoring().GetClusterLogs(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("failed to get cluster logs: %w", err)
			}

			return resp, nil
		},
		PrintText: func(cmd *cobra.Command, w io.Writer, resp *monitoringv1.GetClusterLogsResponse) error {
			timestamps, _ := cmd.Flags().GetBool("timestamps")
			for _, entry := range resp.GetItems() {
				if timestamps {
					fmt.Fprintf(w, "%s  %s\n", output.FullDateTime(entry.GetTimestamp().AsTime()), entry.GetMessage())
				} else {
					fmt.Fprintln(w, entry.GetMessage())
				}
			}

			return nil
		},
		ValidArgsFunction: completion.ClusterIDCompletion(s),
	}.CobraCommand(s)

	cmd.Flags().StringP("since", "s", "", "Start time for logs (RFC3339, YYYY-MM-DD, or a duration ago such as 24h or 7d; default: 3 days ago)")
	cmd.Flags().StringP("until", "u", "", "End time for logs (RFC3339, YYYY-MM-DD, or a duration ago such as 1h; default: now)")
	cmd.Flags().BoolP("timestamps", "t", false, "Prepend each log line with its timestamp")

	return cmd
}
