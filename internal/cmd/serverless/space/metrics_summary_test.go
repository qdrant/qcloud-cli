package space_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"

	serverlessmonitoringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/monitoring/v1"

	"github.com/qdrant/qcloud-cli/internal/resource"
	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func overview(values ...float64) *serverlessmonitoringv1.SpaceMetricOverview {
	intervals := []time.Duration{2 * time.Minute, time.Hour, 3 * time.Hour, 24 * time.Hour}
	o := &serverlessmonitoringv1.SpaceMetricOverview{}
	for i, v := range values {
		o.Avg = append(o.Avg, &serverlessmonitoringv1.IntervalAverage{Interval: durationpb.New(intervals[i]), Value: v})
	}

	return o
}

func sampleQuotaSnapshot() *serverlessmonitoringv1.SpaceQuotaSnapshot {
	return &serverlessmonitoringv1.SpaceQuotaSnapshot{
		MaxCollections:                     10,
		CurrentCollections:                 2,
		EffectiveMaxSizePerCollectionBytes: uint64(5 * resource.GiB),
		SearcherEffectiveMaxWorkers:        4,
	}
}

func sampleSummary() *serverlessmonitoringv1.GetSpaceSummaryMetricsResponse {
	return &serverlessmonitoringv1.GetSpaceSummaryMetricsResponse{
		Quota: sampleQuotaSnapshot(),
		Items: []*serverlessmonitoringv1.SpaceCollectionMetrics{
			{
				CollectionName:         "products",
				PointCount:             12345,
				UsedStorageBytes:       uint64(2 * resource.GiB),
				CurrentSearcherWorkers: 2,
				// Intervals deliberately out of order to check the shortest one is picked.
				SearchRequests: &serverlessmonitoringv1.SpaceMetricOverview{Avg: []*serverlessmonitoringv1.IntervalAverage{
					{Interval: durationpb.New(time.Hour), Value: 9.99},
					{Interval: durationpb.New(2 * time.Minute), Value: 1.5},
				}},
				WriteRequests: overview(0.25, 0.5),
				SearchLatency: overview(12.34, 20),
			},
			{
				CollectionName: "empty",
			},
		},
	}
}

func TestSpaceMetricsSummary_TextOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Returns(sampleSummary(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "summary", "space-abc")
	require.NoError(t, err)

	assert.Contains(t, stdout, "Collections:             2 / 10")
	assert.Contains(t, stdout, "Max Size per Collection: 5GiB")
	assert.Contains(t, stdout, "Max Search Workers:      4")

	assert.Contains(t, stdout, "Rates and latency averaged over the last 2m.")
	for _, h := range []string{"COLLECTION", "POINTS", "STORAGE", "WORKERS", "SEARCHES", "WRITES", "LATENCY"} {
		assert.Contains(t, stdout, h)
	}

	assert.Contains(t, stdout, "products")
	assert.Contains(t, stdout, "12345")
	assert.Contains(t, stdout, "2.0 GiB")
	assert.Contains(t, stdout, "0 B")
	assert.Contains(t, stdout, "1.50/s")
	assert.NotContains(t, stdout, "9.99/s")
	assert.Contains(t, stdout, "0.25/s")
	assert.Contains(t, stdout, "12.3ms")
	assert.Contains(t, stdout, "empty")
	assert.NotContains(t, stdout, "Next page token")

	req, ok := env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "space-abc", req.GetSpaceId())
	assert.Nil(t, req.GetCollectionFilter())
	assert.Nil(t, req.PageSize)
	assert.Nil(t, req.PageToken)
}

func TestSpaceMetricsSummary_Unlimited(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Returns(&serverlessmonitoringv1.GetSpaceSummaryMetricsResponse{
		Quota: &serverlessmonitoringv1.SpaceQuotaSnapshot{CurrentCollections: 1},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "summary", "space-abc")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Collections:             1 / unlimited")
	assert.Contains(t, stdout, "Max Size per Collection: unlimited")
	assert.Contains(t, stdout, "Max Search Workers:      unlimited")
	assert.Contains(t, stdout, "SEARCHES")
	assert.NotContains(t, stdout, "averaged over")
}

func TestSpaceMetricsSummary_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Returns(sampleSummary(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "summary", "space-abc", "--json")
	require.NoError(t, err)

	var result struct {
		Items []struct {
			CollectionName string `json:"collectionName"`
			PointCount     string `json:"pointCount"`
			SearchRequests struct {
				Avg []struct {
					Interval string  `json:"interval"`
					Value    float64 `json:"value"`
				} `json:"avg"`
			} `json:"searchRequests"`
		} `json:"items"`
		Quota struct {
			MaxCollections string `json:"maxCollections"`
		} `json:"quota"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Items, 2)
	assert.Equal(t, "products", result.Items[0].CollectionName)
	assert.Equal(t, "12345", result.Items[0].PointCount)
	require.Len(t, result.Items[0].SearchRequests.Avg, 2)
	assert.Equal(t, "3600s", result.Items[0].SearchRequests.Avg[0].Interval)
	assert.Equal(t, "10", result.Quota.MaxCollections)
}

func TestSpaceMetricsSummary_CollectionFilters(t *testing.T) {
	t.Run("exact name", func(t *testing.T) {
		env := testutil.NewTestEnv(t)
		env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Returns(sampleSummary(), nil)

		_, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "summary", "space-abc", "--collection", "products")
		require.NoError(t, err)

		req, ok := env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Last()
		require.True(t, ok)
		assert.Equal(t, "products", req.GetCollectionName())
	})

	t.Run("contains", func(t *testing.T) {
		env := testutil.NewTestEnv(t)
		env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Returns(sampleSummary(), nil)

		_, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "summary", "space-abc", "--collection-contains", "prod")
		require.NoError(t, err)

		req, ok := env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Last()
		require.True(t, ok)
		assert.Equal(t, "prod", req.GetCollectionNameContains())
	})

	t.Run("mutually exclusive", func(t *testing.T) {
		env := testutil.NewTestEnv(t)

		_, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "summary", "space-abc",
			"--collection", "products", "--collection-contains", "prod")
		require.Error(t, err)
		assert.Equal(t, 0, env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Count())
	})
}

func TestSpaceMetricsSummary_AutoPagination(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.
		OnCall(0, func(_ context.Context, _ *serverlessmonitoringv1.GetSpaceSummaryMetricsRequest) (*serverlessmonitoringv1.GetSpaceSummaryMetricsResponse, error) {
			return &serverlessmonitoringv1.GetSpaceSummaryMetricsResponse{
				Quota:         sampleQuotaSnapshot(),
				Items:         []*serverlessmonitoringv1.SpaceCollectionMetrics{{CollectionName: "first"}},
				NextPageToken: new("page-2"),
			}, nil
		}).
		OnCall(1, func(_ context.Context, _ *serverlessmonitoringv1.GetSpaceSummaryMetricsRequest) (*serverlessmonitoringv1.GetSpaceSummaryMetricsResponse, error) {
			return &serverlessmonitoringv1.GetSpaceSummaryMetricsResponse{
				Quota: &serverlessmonitoringv1.SpaceQuotaSnapshot{MaxCollections: 99},
				Items: []*serverlessmonitoringv1.SpaceCollectionMetrics{{CollectionName: "second"}},
			}, nil
		})

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "summary", "space-abc", "--json")
	require.NoError(t, err)

	var result struct {
		Items []struct {
			CollectionName string `json:"collectionName"`
		} `json:"items"`
		Quota struct {
			MaxCollections string `json:"maxCollections"`
		} `json:"quota"`
		NextPageToken *string `json:"nextPageToken"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Items, 2)
	assert.Equal(t, "first", result.Items[0].CollectionName)
	assert.Equal(t, "second", result.Items[1].CollectionName)
	assert.Equal(t, "10", result.Quota.MaxCollections)
	assert.Nil(t, result.NextPageToken)

	assert.Equal(t, 2, env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Count())
	req, ok := env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "page-2", req.GetPageToken())
}

func TestSpaceMetricsSummary_ManualPagination(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Returns(&serverlessmonitoringv1.GetSpaceSummaryMetricsResponse{
		Items:         []*serverlessmonitoringv1.SpaceCollectionMetrics{{CollectionName: "first"}},
		NextPageToken: new("page-2"),
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "summary", "space-abc", "--page-size", "1")
	require.NoError(t, err)
	assert.NotContains(t, stdout, "page-2")
	assert.Equal(t, 1, env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Count())

	req, ok := env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, int32(1), req.GetPageSize())
}

func TestSpaceMetricsSummary_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "summary", "space-abc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get space summary metrics")
}

func TestSpaceMetricsSummary_NilIntervalIgnored(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Returns(&serverlessmonitoringv1.GetSpaceSummaryMetricsResponse{
		Items: []*serverlessmonitoringv1.SpaceCollectionMetrics{{
			CollectionName: "products",
			SearchRequests: &serverlessmonitoringv1.SpaceMetricOverview{Avg: []*serverlessmonitoringv1.IntervalAverage{
				{Value: 99},
				{Interval: durationpb.New(time.Hour), Value: 1.5},
			}},
		}},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "summary", "space-abc")
	require.NoError(t, err)
	assert.Contains(t, stdout, "averaged over the last 1h.")
	assert.Regexp(t, `products\s+0\s+0 B\s+0\s+1\.50/s`, stdout)
}

func TestSpaceMetricsSummary_ConflictingFilters(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "summary", "space-abc",
		"--collection", "a", "--collection-contains", "b")
	require.Error(t, err)
	assert.Equal(t, 0, env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Count())
}

func TestSpaceMetricsSummary_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "summary")
	require.Error(t, err)
	assert.Equal(t, 0, env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Count())
}

func TestSpaceMetrics_Help(t *testing.T) {
	env := testutil.NewTestEnv(t)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "--help")
	require.NoError(t, err)
	assert.Contains(t, stdout, "metrics")

	stdout, _, err = testutil.Exec(t, env, "serverless", "space", "metrics", "--help")
	require.NoError(t, err)
	assert.Contains(t, stdout, "summary")
	assert.Contains(t, stdout, "usage")
}

func sampleSummaryTotals() *serverlessmonitoringv1.SpaceSummaryTotals {
	return &serverlessmonitoringv1.SpaceSummaryTotals{
		PointCount:             54321,
		UsedStorageBytes:       uint64(3 * resource.GiB),
		CurrentSearcherWorkers: 5,
		SearchRequests:         overview(4.5, 9),
		WriteRequests:          overview(0.75, 1),
		SearchLatency:          overview(8.5, 9),
	}
}

func TestSpaceMetricsSummary_Totals(t *testing.T) {
	env := testutil.NewTestEnv(t)
	resp := sampleSummary()
	resp.Totals = sampleSummaryTotals()
	resp.TotalSize = new(int32(2))
	env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Returns(resp, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "summary", "space-abc")
	require.NoError(t, err)

	assert.Contains(t, stdout, "Totals:")
	assert.Contains(t, stdout, "  Points:         54321")
	assert.Contains(t, stdout, "  Storage:        3.0 GiB")
	assert.Contains(t, stdout, "  Search Workers: 5")
	assert.Contains(t, stdout, "  Searches:       4.50/s")
	assert.Contains(t, stdout, "  Writes:         0.75/s")
	assert.Contains(t, stdout, "  Latency:        8.5ms")
	assert.Contains(t, stdout, "COLLECTION")
	assert.Contains(t, stdout, "products")

	req, ok := env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Last()
	require.True(t, ok)
	assert.False(t, req.GetTotalsOnly())
}

func TestSpaceMetricsSummary_NoTotalsSection(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Returns(sampleSummary(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "summary", "space-abc")
	require.NoError(t, err)
	assert.NotContains(t, stdout, "Totals:")
}

func TestSpaceMetricsSummary_TotalsKeptAcrossPages(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.
		OnCall(0, func(_ context.Context, _ *serverlessmonitoringv1.GetSpaceSummaryMetricsRequest) (*serverlessmonitoringv1.GetSpaceSummaryMetricsResponse, error) {
			return &serverlessmonitoringv1.GetSpaceSummaryMetricsResponse{
				Quota:         sampleQuotaSnapshot(),
				Totals:        sampleSummaryTotals(),
				TotalSize:     new(int32(2)),
				Items:         []*serverlessmonitoringv1.SpaceCollectionMetrics{{CollectionName: "first"}},
				NextPageToken: new("page-2"),
			}, nil
		}).
		OnCall(1, func(_ context.Context, _ *serverlessmonitoringv1.GetSpaceSummaryMetricsRequest) (*serverlessmonitoringv1.GetSpaceSummaryMetricsResponse, error) {
			return &serverlessmonitoringv1.GetSpaceSummaryMetricsResponse{
				Quota:     sampleQuotaSnapshot(),
				Totals:    sampleSummaryTotals(),
				TotalSize: new(int32(2)),
				Items:     []*serverlessmonitoringv1.SpaceCollectionMetrics{{CollectionName: "second"}},
			}, nil
		})

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "summary", "space-abc", "--json")
	require.NoError(t, err)

	var result struct {
		Items  []json.RawMessage `json:"items"`
		Totals struct {
			PointCount       string `json:"pointCount"`
			UsedStorageBytes string `json:"usedStorageBytes"`
			SearchRequests   struct {
				Avg []struct {
					Value float64 `json:"value"`
				} `json:"avg"`
			} `json:"searchRequests"`
		} `json:"totals"`
		TotalSize *int32 `json:"totalSize"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Len(t, result.Items, 2)
	assert.Equal(t, "54321", result.Totals.PointCount)
	assert.Equal(t, "3221225472", result.Totals.UsedStorageBytes)
	require.Len(t, result.Totals.SearchRequests.Avg, 2)
	assert.InDelta(t, 4.5, result.Totals.SearchRequests.Avg[0].Value, 1e-9)
	// When all pages are fetched, the total size stays unset like in list commands.
	assert.Nil(t, result.TotalSize)
}

func TestSpaceMetricsSummary_TotalsOnly(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Returns(&serverlessmonitoringv1.GetSpaceSummaryMetricsResponse{
		Quota:     sampleQuotaSnapshot(),
		Totals:    sampleSummaryTotals(),
		TotalSize: new(int32(2)),
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "summary", "space-abc",
		"--totals-only", "--collection-contains", "prod")
	require.NoError(t, err)

	assert.Contains(t, stdout, "Collections:             2 / 10")
	assert.Contains(t, stdout, "Rates and latency averaged over the last 2m.")
	assert.Contains(t, stdout, "  Points:         54321")
	assert.Contains(t, stdout, "  Searches:       4.50/s")
	assert.NotContains(t, stdout, "COLLECTION")

	assert.Equal(t, 1, env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Count())
	req, ok := env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Last()
	require.True(t, ok)
	assert.True(t, req.GetTotalsOnly())
	assert.Equal(t, "prod", req.GetCollectionNameContains())
	assert.Nil(t, req.PageSize)
	assert.Nil(t, req.PageToken)
}

func TestSpaceMetricsSummary_TotalsOnlyWithPagination(t *testing.T) {
	for _, flags := range [][]string{{"--page-size", "5"}, {"--page-token", "page-2"}} {
		t.Run(flags[0], func(t *testing.T) {
			env := testutil.NewTestEnv(t)
			args := append([]string{"serverless", "space", "metrics", "summary", "space-abc", "--totals-only"}, flags...)
			_, _, err := testutil.Exec(t, env, args...)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "totals-only")
			assert.Equal(t, 0, env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Count())
		})
	}
}

func TestSpaceMetricsSummary_OrderBy(t *testing.T) {
	tests := []struct {
		value string
		want  serverlessmonitoringv1.SpaceCollectionMetricsOrder
	}{
		{"name", serverlessmonitoringv1.SpaceCollectionMetricsOrder_SPACE_COLLECTION_METRICS_ORDER_UNSPECIFIED},
		{"points", serverlessmonitoringv1.SpaceCollectionMetricsOrder_SPACE_COLLECTION_METRICS_ORDER_POINT_COUNT_DESC},
		{"storage", serverlessmonitoringv1.SpaceCollectionMetricsOrder_SPACE_COLLECTION_METRICS_ORDER_USED_STORAGE_BYTES_DESC},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			env := testutil.NewTestEnv(t)
			env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Returns(sampleSummary(), nil)

			_, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "summary", "space-abc", "--order-by", tt.value)
			require.NoError(t, err)

			req, ok := env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Last()
			require.True(t, ok)
			assert.Equal(t, tt.want, req.GetOrderBy())
		})
	}
}

func TestSpaceMetricsSummary_OrderByDefault(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Returns(sampleSummary(), nil)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "summary", "space-abc")
	require.NoError(t, err)

	req, ok := env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, serverlessmonitoringv1.SpaceCollectionMetricsOrder_SPACE_COLLECTION_METRICS_ORDER_UNSPECIFIED, req.GetOrderBy())
}

func TestSpaceMetricsSummary_OrderByKeptAcrossPages(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.
		OnCall(0, func(_ context.Context, _ *serverlessmonitoringv1.GetSpaceSummaryMetricsRequest) (*serverlessmonitoringv1.GetSpaceSummaryMetricsResponse, error) {
			return &serverlessmonitoringv1.GetSpaceSummaryMetricsResponse{
				Items:         []*serverlessmonitoringv1.SpaceCollectionMetrics{{CollectionName: "big"}},
				NextPageToken: new("page-2"),
			}, nil
		}).
		OnCall(1, func(_ context.Context, _ *serverlessmonitoringv1.GetSpaceSummaryMetricsRequest) (*serverlessmonitoringv1.GetSpaceSummaryMetricsResponse, error) {
			return &serverlessmonitoringv1.GetSpaceSummaryMetricsResponse{
				Items: []*serverlessmonitoringv1.SpaceCollectionMetrics{{CollectionName: "small"}},
			}, nil
		})

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "summary", "space-abc", "--order-by", "points")
	require.NoError(t, err)
	assert.Less(t, strings.Index(stdout, "big"), strings.Index(stdout, "small"), "the server's order is kept")

	reqs := env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.All()
	require.Len(t, reqs, 2)
	for _, req := range reqs {
		assert.Equal(t, serverlessmonitoringv1.SpaceCollectionMetricsOrder_SPACE_COLLECTION_METRICS_ORDER_POINT_COUNT_DESC, req.GetOrderBy())
	}
}

func TestSpaceMetricsSummary_OrderByInvalid(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "summary", "space-abc", "--order-by", "size")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `invalid --order-by "size"`)
	assert.Equal(t, 0, env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Count())
}

func TestSpaceMetricsSummary_OrderByWithTotalsOnly(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "summary", "space-abc", "--order-by", "points", "--totals-only")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "order-by")
	assert.Equal(t, 0, env.ServerlessMonitoringServer.GetSpaceSummaryMetricsCalls.Count())
}
