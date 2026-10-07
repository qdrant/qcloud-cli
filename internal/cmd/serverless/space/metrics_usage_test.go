package space_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	serverlessmonitoringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/monitoring/v1"

	"github.com/qdrant/qcloud-cli/internal/resource"
	"github.com/qdrant/qcloud-cli/internal/testutil"
)

var usageStart = time.Date(2026, 10, 7, 10, 0, 0, 0, time.UTC)

func series(values ...float64) []*serverlessmonitoringv1.Metric {
	out := make([]*serverlessmonitoringv1.Metric, 0, len(values))
	for i, v := range values {
		out = append(out, &serverlessmonitoringv1.Metric{
			Timestamp: timestamppb.New(usageStart.Add(time.Duration(i) * time.Minute)),
			Value:     v,
		})
	}

	return out
}

func sampleUsage() *serverlessmonitoringv1.GetSpaceUsageMetricsResponse {
	return &serverlessmonitoringv1.GetSpaceUsageMetricsResponse{
		Quota: sampleQuotaSnapshot(),
		Items: []*serverlessmonitoringv1.SpaceCollectionUsageMetrics{
			{
				CollectionName:         "products",
				PointCount:             series(100, 200, 300),
				UsedStorageBytes:       series(0, float64(resource.GiB), float64(2*resource.GiB)),
				CurrentSearcherWorkers: series(1, 3, 2),
				SearchRequests:         series(1, 2, 6),
				WriteRequests:          series(0.5, 0.5, 2),
				SearchLatency:          series(10, 40, 25),
			},
			{
				CollectionName: "empty",
			},
		},
	}
}

func TestSpaceMetricsUsage_TextOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceUsageMetricsCalls.Returns(sampleUsage(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "usage", "space-abc")
	require.NoError(t, err)

	assert.Contains(t, stdout, "Collections:             2 / 10")
	assert.Contains(t, stdout, "Period: 2026-10-07 10:00:00 UTC to 2026-10-07 10:02:00 UTC")
	for _, h := range []string{
		"COLLECTION", "POINTS", "STORAGE", "WORKERS",
		"SEARCHES AVG", "SEARCHES MAX", "WRITES AVG", "WRITES MAX", "LATENCY AVG", "LATENCY MAX",
	} {
		assert.Contains(t, stdout, h)
	}

	// Gauges show the latest value, rates and latency show avg and max.
	assert.Regexp(t, `products\s+300\s+2\.0 GiB\s+2\s+3\.00/s\s+6\.00/s\s+1\.00/s\s+2\.00/s\s+25\.0ms\s+40\.0ms`, stdout)
	assert.Regexp(t, `empty(\s+-){9}`, stdout)

	req, ok := env.ServerlessMonitoringServer.GetSpaceUsageMetricsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "space-abc", req.GetSpaceId())
	assert.Nil(t, req.Since)
	assert.Nil(t, req.Until)
	assert.Nil(t, req.Aggregator)
	assert.Nil(t, req.GetCollectionFilter())
}

func TestSpaceMetricsUsage_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceUsageMetricsCalls.Returns(sampleUsage(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "usage", "space-abc", "--json")
	require.NoError(t, err)

	var result struct {
		Items []struct {
			CollectionName string `json:"collectionName"`
			SearchRequests []struct {
				Timestamp string  `json:"timestamp"`
				Value     float64 `json:"value"`
			} `json:"searchRequests"`
		} `json:"items"`
		Quota struct {
			CurrentCollections string `json:"currentCollections"`
		} `json:"quota"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Items, 2)
	assert.Equal(t, "products", result.Items[0].CollectionName)
	require.Len(t, result.Items[0].SearchRequests, 3)
	assert.InDelta(t, 6.0, result.Items[0].SearchRequests[2].Value, 1e-9)
	assert.Equal(t, "2", result.Quota.CurrentCollections)
}

func TestSpaceMetricsUsage_RequestFields(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceUsageMetricsCalls.Returns(sampleUsage(), nil)

	before := time.Now()
	_, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "usage", "space-abc",
		"--since", "24h",
		"--until", "2099-01-01T00:00:00Z",
		"--aggregator", "max",
		"--collection-contains", "prod",
	)
	require.NoError(t, err)

	req, ok := env.ServerlessMonitoringServer.GetSpaceUsageMetricsCalls.Last()
	require.True(t, ok)
	assert.WithinDuration(t, before.Add(-24*time.Hour), req.GetSince().AsTime(), time.Minute)
	assert.Equal(t, time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC), req.GetUntil().AsTime())
	assert.Equal(t, serverlessmonitoringv1.Aggregator_AGGREGATOR_MAX, req.GetAggregator())
	assert.Equal(t, "prod", req.GetCollectionNameContains())
}

func TestSpaceMetricsUsage_SameDateRangeCoversWholeDay(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceUsageMetricsCalls.Returns(sampleUsage(), nil)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "usage", "space-abc",
		"--since", "2026-10-01",
		"--until", "2026-10-01",
	)
	require.NoError(t, err)

	req, ok := env.ServerlessMonitoringServer.GetSpaceUsageMetricsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), req.GetSince().AsTime())
	assert.Equal(t, time.Date(2026, 10, 1, 23, 59, 59, 0, time.UTC), req.GetUntil().AsTime())
}

func TestSpaceMetricsUsage_Aggregators(t *testing.T) {
	cases := map[string]serverlessmonitoringv1.Aggregator{
		"sum": serverlessmonitoringv1.Aggregator_AGGREGATOR_SUM,
		"avg": serverlessmonitoringv1.Aggregator_AGGREGATOR_AVG,
		"max": serverlessmonitoringv1.Aggregator_AGGREGATOR_MAX,
		"min": serverlessmonitoringv1.Aggregator_AGGREGATOR_MIN,
	}
	for flag, want := range cases {
		t.Run(flag, func(t *testing.T) {
			env := testutil.NewTestEnv(t)
			env.ServerlessMonitoringServer.GetSpaceUsageMetricsCalls.Returns(sampleUsage(), nil)

			_, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "usage", "space-abc", "--aggregator", flag)
			require.NoError(t, err)

			req, ok := env.ServerlessMonitoringServer.GetSpaceUsageMetricsCalls.Last()
			require.True(t, ok)
			assert.Equal(t, want, req.GetAggregator())
		})
	}
}

func TestSpaceMetricsUsage_CollectionName(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceUsageMetricsCalls.Returns(sampleUsage(), nil)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "usage", "space-abc", "--collection", "products")
	require.NoError(t, err)

	req, ok := env.ServerlessMonitoringServer.GetSpaceUsageMetricsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "products", req.GetCollectionName())
}

func TestSpaceMetricsUsage_InputErrors(t *testing.T) {
	cases := map[string][]string{
		"missing space id":    {},
		"invalid since":       {"space-abc", "--since", "yesterday"},
		"invalid until":       {"space-abc", "--until", "-5h"},
		"since after until":   {"space-abc", "--since", "1h", "--until", "2h"},
		"unknown aggregator":  {"space-abc", "--aggregator", "median"},
		"conflicting filters": {"space-abc", "--collection", "a", "--collection-contains", "b"},
	}
	for name, extra := range cases {
		t.Run(name, func(t *testing.T) {
			env := testutil.NewTestEnv(t)

			args := append([]string{"serverless", "space", "metrics", "usage"}, extra...)
			_, _, err := testutil.Exec(t, env, args...)
			require.Error(t, err)
			assert.Equal(t, 0, env.ServerlessMonitoringServer.GetSpaceUsageMetricsCalls.Count())
		})
	}
}

func TestSpaceMetricsUsage_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceUsageMetricsCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "usage", "space-abc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get space usage metrics")
}
