package space

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"

	serverlessmonitoringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/monitoring/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/resource"
)

// addCollectionFilterFlags registers the mutually exclusive flags that map to
// the collection_filter oneof of the monitoring requests. When allowSpaceOnly
// is set, --space-only is registered as a third option.
func addCollectionFilterFlags(cmd *cobra.Command, allowSpaceOnly bool) {
	cmd.Flags().String("collection", "", "Only include the collection with this exact name")
	cmd.Flags().String("collection-contains", "", "Only include collections whose name contains this substring")
	names := []string{"collection", "collection-contains"}
	if allowSpaceOnly {
		cmd.Flags().Bool("space-only", false, "Only include entries that are not tied to a collection")
		names = append(names, "space-only")
	}

	cmd.MarkFlagsMutuallyExclusive(names...)
}

// collectionFilter holds the collection filter flags. At most one field is set.
type collectionFilter struct {
	name      *string
	contains  *string
	spaceOnly bool
}

func readCollectionFilter(cmd *cobra.Command) collectionFilter {
	var f collectionFilter
	if cmd.Flags().Changed("collection") {
		v, _ := cmd.Flags().GetString("collection")
		f.name = &v
	}

	if cmd.Flags().Changed("collection-contains") {
		v, _ := cmd.Flags().GetString("collection-contains")
		f.contains = &v
	}

	if cmd.Flags().Lookup("space-only") != nil {
		f.spaceOnly, _ = cmd.Flags().GetBool("space-only")
	}

	return f
}

// addTimeRangeFlags registers --since and --until. defaultSince describes the
// server-side default used when --since is omitted.
func addTimeRangeFlags(cmd *cobra.Command, defaultSince string) {
	cmd.Flags().String("since", "", "Start of the period (RFC3339, YYYY-MM-DD, or a duration ago such as 6h or 7d; default "+defaultSince+")")
	cmd.Flags().String("until", "", "End of the period (RFC3339, YYYY-MM-DD, or a duration ago such as 1h; default now)")
}

// readTimeRange parses --since and --until. Unset flags yield nil so that the
// server defaults apply.
func readTimeRange(cmd *cobra.Command, now time.Time) (since, until *timestamppb.Timestamp, err error) {
	var sinceT, untilT time.Time
	if cmd.Flags().Changed("since") {
		v, _ := cmd.Flags().GetString("since")
		sinceT, err = util.ParseTimeFlag(v, now)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid --since %q: %w", v, err)
		}

		since = timestamppb.New(sinceT)
	}

	if cmd.Flags().Changed("until") {
		v, _ := cmd.Flags().GetString("until")
		untilT, err = util.ParseUntilFlag(v, now)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid --until %q: %w", v, err)
		}

		until = timestamppb.New(untilT)
	}

	if since != nil && until != nil && !sinceT.Before(untilT) {
		return nil, nil, fmt.Errorf("--since must be before --until")
	}

	return since, until, nil
}

// seriesRollup summarizes a metric time series.
type seriesRollup struct {
	latest float64
	avg    float64
	max    float64
	ok     bool
}

// rollup computes the latest, average and maximum value of a series.
func rollup(series []*serverlessmonitoringv1.Metric) seriesRollup {
	if len(series) == 0 {
		return seriesRollup{}
	}

	r := seriesRollup{ok: true, max: series[0].GetValue()}
	var sum float64
	for _, m := range series {
		v := m.GetValue()
		sum += v
		r.max = max(r.max, v)
	}

	r.avg = sum / float64(len(series))
	r.latest = series[len(series)-1].GetValue()
	return r
}

// printQuotaSnapshot prints the space quota returned with monitoring responses.
func printQuotaSnapshot(w io.Writer, q *serverlessmonitoringv1.SpaceQuotaSnapshot) {
	if q == nil {
		return
	}

	maxCollections := output.UnlimitedIfZero(strconv.FormatUint(q.GetMaxCollections(), 10), q.GetMaxCollections())
	fmt.Fprintln(w, "Quota:")
	fmt.Fprintf(w, "  Collections:             %d / %s\n", q.GetCurrentCollections(), maxCollections)
	fmt.Fprintf(w, "  Max Size per Collection: %s\n",
		output.UnlimitedIfZero(resource.ByteQuantity(q.GetEffectiveMaxSizePerCollectionBytes()).String(), q.GetEffectiveMaxSizePerCollectionBytes()))
	fmt.Fprintf(w, "  Max Search Workers:      %s\n",
		output.UnlimitedIfZero(strconv.FormatUint(q.GetSearcherEffectiveMaxWorkers(), 10), q.GetSearcherEffectiveMaxWorkers()))
	fmt.Fprintln(w)
}

// parseAggregator maps the --aggregator flag value to the proto enum.
func parseAggregator(s string) (serverlessmonitoringv1.Aggregator, error) {
	switch s {
	case "sum":
		return serverlessmonitoringv1.Aggregator_AGGREGATOR_SUM, nil
	case "avg":
		return serverlessmonitoringv1.Aggregator_AGGREGATOR_AVG, nil
	case "max":
		return serverlessmonitoringv1.Aggregator_AGGREGATOR_MAX, nil
	case "min":
		return serverlessmonitoringv1.Aggregator_AGGREGATOR_MIN, nil
	default:
		return 0, fmt.Errorf("invalid --aggregator %q: must be one of sum, avg, max, min", s)
	}
}
