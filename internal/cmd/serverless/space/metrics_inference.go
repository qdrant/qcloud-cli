package space

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"

	serverlessmonitoringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/monitoring/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

// defaultInferenceMetricsRange matches the server default of 7 days.
const defaultInferenceMetricsRange = 7 * 24 * time.Hour

var inferenceIntervals = []string{"day", "week", "month"}

func newMetricsInferenceCommand(s *state.State) *cobra.Command {
	cmd := base.DescribeCmd[*serverlessmonitoringv1.GetSpaceInferenceMetricsResponse]{
		Use:   "inference <space-id>",
		Short: "Show inference token usage of a space's collections",
		Long: `Show the inference token usage of the collections in a serverless space over a period.

The server reports, per collection and inference model, the number of tokens
processed by inference, aggregated into buckets of one day, week or month. The
table shows one row per collection and model with the total number of tokens
over the period and the largest bucket. The total over all rows is shown above
the table. Use --json to get the individual buckets.

By default, the period covers the last 7 days and buckets are aggregated per
day. When only --until is given, the period covers the 7 days before it.

By default, all collections are fetched automatically across multiple pages.
Use --page-size and --page-token for manual pagination.`,
		Example: `# Show inference token usage over the last 7 days
qcloud serverless space metrics inference 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Show weekly token usage over the last 90 days
qcloud serverless space metrics inference 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --since 90d --interval week

# Show token usage of a single model in a single collection during October
qcloud serverless space metrics inference 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --collection products --model sentence-transformers/all-minilm-l6-v2 \
  --since 2026-10-01 --until 2026-10-31

# Export the daily buckets as JSON
qcloud serverless space metrics inference 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --json`,
		Args: util.ExactArgs(1, "a space ID"),
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*serverlessmonitoringv1.GetSpaceInferenceMetricsResponse, error) {
			since, until, err := util.ReadPastTimeRange(cmd, time.Now(), defaultInferenceMetricsRange)
			if err != nil {
				return nil, err
			}

			filter, err := readCollectionFilter(cmd)
			if err != nil {
				return nil, err
			}

			var interval *serverlessmonitoringv1.InferenceMetricsInterval
			if cmd.Flags().Changed("interval") {
				v, _ := cmd.Flags().GetString("interval")
				i, err := parseInferenceInterval(v)
				if err != nil {
					return nil, err
				}

				interval = &i
			}

			var model *string
			if cmd.Flags().Changed("model") {
				v, _ := cmd.Flags().GetString("model")
				if v == "" {
					return nil, errors.New("--model must not be empty")
				}

				model = &v
			}

			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			items, next, total, err := util.FetchPages(cmd, func(pageSize *int32, pageToken *string) ([]*serverlessmonitoringv1.SpaceCollectionInferenceMetrics, string, *int32, error) {
				req := &serverlessmonitoringv1.GetSpaceInferenceMetricsRequest{
					AccountId:        accountID,
					SpaceId:          args[0],
					Since:            timestamppb.New(since),
					Until:            timestamppb.New(until),
					Interval:         interval,
					InferenceModelId: model,
					PageSize:         pageSize,
					PageToken:        pageToken,
				}
				switch {
				case filter.name != nil:
					req.CollectionFilter = &serverlessmonitoringv1.GetSpaceInferenceMetricsRequest_CollectionName{CollectionName: *filter.name}
				case filter.contains != nil:
					req.CollectionFilter = &serverlessmonitoringv1.GetSpaceInferenceMetricsRequest_CollectionNameContains{CollectionNameContains: *filter.contains}
				}

				resp, err := client.ServerlessMonitoring().GetSpaceInferenceMetrics(ctx, req)
				if err != nil {
					return nil, "", nil, fmt.Errorf("failed to get space inference metrics: %w", err)
				}

				return resp.GetItems(), resp.GetNextPageToken(), resp.TotalSize, nil
			})
			if err != nil {
				return nil, err
			}

			return &serverlessmonitoringv1.GetSpaceInferenceMetricsResponse{
				Items:         items,
				TotalSize:     total,
				NextPageToken: next,
			}, nil
		},
		PrintText: func(cmd *cobra.Command, w io.Writer, resp *serverlessmonitoringv1.GetSpaceInferenceMetricsResponse) error {
			intervalName := "day"
			if cmd.Flags().Changed("interval") {
				intervalName, _ = cmd.Flags().GetString("interval")
			}

			type row struct {
				collection string
				model      string
				total      float64
				peak       float64
				ok         bool
			}

			var rows []row
			var grandTotal float64
			var first, last time.Time
			var hasPeriod bool
			for _, item := range resp.GetItems() {
				if len(item.GetModels()) == 0 {
					rows = append(rows, row{collection: item.GetCollectionName(), model: "-"})
					continue
				}

				for _, m := range item.GetModels() {
					r := row{collection: item.GetCollectionName(), model: m.GetInferenceModelId()}
					for _, v := range m.GetValues() {
						r.ok = true
						r.total += v.GetValue()
						r.peak = max(r.peak, v.GetValue())
						if v.GetTimestamp() == nil {
							continue
						}

						ts := v.GetTimestamp().AsTime()
						if !hasPeriod || ts.Before(first) {
							first = ts
						}

						if !hasPeriod || ts.After(last) {
							last = ts
						}

						hasPeriod = true
					}

					grandTotal += r.total
					rows = append(rows, r)
				}
			}

			if hasPeriod {
				fmt.Fprintf(w, "Buckets: per %s, %s to %s\n", intervalName, output.FullDateTime(first), output.FullDateTime(last))
			}

			fmt.Fprintf(w, "Total Tokens: %s\n\n", tokens(grandTotal))

			value := func(r row, v float64) string {
				if !r.ok {
					return "-"
				}

				return tokens(v)
			}

			t := output.NewTable[row](w)
			t.AddField("COLLECTION", func(r row) string { return r.collection })
			t.AddField("MODEL", func(r row) string { return r.model })
			t.AddField("TOKENS", func(r row) string { return value(r, r.total) })
			t.AddField("PEAK PER "+strings.ToUpper(intervalName), func(r row) string { return value(r, r.peak) })
			t.SetItems(rows)
			t.Render()

			return nil
		},
		ValidArgsFunction: completion.SpaceIDCompletion(s),
	}.CobraCommand(s)

	util.AddTimeRangeFlags(cmd, defaultInferenceMetricsRange)
	cmd.Flags().String("interval", "", "Size of the aggregation buckets (day, week, month; default day)")
	_ = cmd.RegisterFlagCompletionFunc("interval", cobra.FixedCompletions(inferenceIntervals, cobra.ShellCompDirectiveNoFileComp))
	cmd.Flags().String("model", "", "Only include the inference model with this ID")
	addCollectionFilterFlags(cmd, false)
	util.AddPaginationFlags(cmd, "collections")

	return cmd
}

// parseInferenceInterval maps the --interval flag value to the proto enum.
func parseInferenceInterval(s string) (serverlessmonitoringv1.InferenceMetricsInterval, error) {
	switch s {
	case "day":
		return serverlessmonitoringv1.InferenceMetricsInterval_INFERENCE_METRICS_INTERVAL_DAY, nil
	case "week":
		return serverlessmonitoringv1.InferenceMetricsInterval_INFERENCE_METRICS_INTERVAL_WEEK, nil
	case "month":
		return serverlessmonitoringv1.InferenceMetricsInterval_INFERENCE_METRICS_INTERVAL_MONTH, nil
	default:
		return 0, fmt.Errorf("invalid --interval %q: must be one of %s", s, strings.Join(inferenceIntervals, ", "))
	}
}

// tokens formats a token count as a whole number.
func tokens(v float64) string {
	return strconv.FormatUint(uint64(v+0.5), 10)
}
