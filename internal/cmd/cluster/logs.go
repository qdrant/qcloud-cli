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
		Example: `# Get logs for a cluster
qcloud cluster logs abc-123

# Get logs since a specific date
qcloud cluster logs abc-123 --since 2024-01-01

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
				t, err := parseLogTime(sinceStr)
				if err != nil {
					return nil, fmt.Errorf("invalid --since %q: must be RFC3339 or YYYY-MM-DD", sinceStr)
				}
				req.Since = timestamppb.New(t)
			}

			if cmd.Flags().Changed("until") {
				untilStr, _ := cmd.Flags().GetString("until")
				t, err := parseLogTime(untilStr)
				if err != nil {
					return nil, fmt.Errorf("invalid --until %q: must be RFC3339 or YYYY-MM-DD", untilStr)
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

	cmd.Flags().StringP("since", "s", "", "Start time for logs (RFC3339 or YYYY-MM-DD, default: 3 days ago)")
	cmd.Flags().StringP("until", "u", "", "End time for logs (RFC3339 or YYYY-MM-DD, default: now)")
	cmd.Flags().BoolP("timestamps", "t", false, "Prepend each log line with its timestamp")

	return cmd
}

func parseLogTime(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", s)
}
