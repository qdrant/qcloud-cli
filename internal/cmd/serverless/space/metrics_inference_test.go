package space_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	serverlessmonitoringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/monitoring/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

var inferenceStart = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

func dailySeries(values ...float64) []*serverlessmonitoringv1.Metric {
	out := make([]*serverlessmonitoringv1.Metric, 0, len(values))
	for i, v := range values {
		out = append(out, &serverlessmonitoringv1.Metric{
			Timestamp: timestamppb.New(inferenceStart.AddDate(0, 0, i)),
			Value:     v,
		})
	}

	return out
}

func sampleInference() *serverlessmonitoringv1.GetSpaceInferenceMetricsResponse {
	return &serverlessmonitoringv1.GetSpaceInferenceMetricsResponse{
		Items: []*serverlessmonitoringv1.SpaceCollectionInferenceMetrics{
			{
				CollectionName: "products",
				Models: []*serverlessmonitoringv1.SpaceInferenceModelMetrics{
					{InferenceModelId: "model-a", Values: dailySeries(1000, 5000, 2500)},
					{InferenceModelId: "model-b", Values: dailySeries(10, 20)},
				},
			},
			{
				CollectionName: "docs",
				Models: []*serverlessmonitoringv1.SpaceInferenceModelMetrics{
					{InferenceModelId: "model-a"},
				},
			},
			{CollectionName: "empty"},
		},
		TotalSize: new(int32(3)),
	}
}

func TestSpaceMetricsInference_TextOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceInferenceMetricsCalls.Returns(sampleInference(), nil)

	before := time.Now()
	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "inference", "space-abc")
	require.NoError(t, err)

	assert.Contains(t, stdout, "Buckets: per day, 2026-10-01 00:00:00 UTC to 2026-10-03 00:00:00 UTC")
	assert.Contains(t, stdout, "Total Tokens: 8530")
	for _, h := range []string{"COLLECTION", "MODEL", "TOKENS", "PEAK PER DAY"} {
		assert.Contains(t, stdout, h)
	}

	assert.Regexp(t, `products\s+model-a\s+8500\s+5000`, stdout)
	assert.Regexp(t, `products\s+model-b\s+30\s+20`, stdout)
	assert.Regexp(t, `docs\s+model-a\s+-\s+-`, stdout)
	assert.Regexp(t, `empty\s+-\s+-\s+-`, stdout)

	req, ok := env.ServerlessMonitoringServer.GetSpaceInferenceMetricsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "space-abc", req.GetSpaceId())
	assert.WithinRange(t, req.GetUntil().AsTime(), before, time.Now())
	assert.Equal(t, 7*24*time.Hour, req.GetUntil().AsTime().Sub(req.GetSince().AsTime()))
	assert.Nil(t, req.Interval)
	assert.Nil(t, req.InferenceModelId)
	assert.Nil(t, req.GetCollectionFilter())
	assert.Nil(t, req.PageSize)
}

func TestSpaceMetricsInference_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceInferenceMetricsCalls.Returns(sampleInference(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "inference", "space-abc", "--json")
	require.NoError(t, err)

	var result struct {
		Items []struct {
			CollectionName string `json:"collectionName"`
			Models         []struct {
				InferenceModelID string `json:"inferenceModelId"`
				Values           []struct {
					Timestamp string  `json:"timestamp"`
					Value     float64 `json:"value"`
				} `json:"values"`
			} `json:"models"`
		} `json:"items"`
		TotalSize int32 `json:"totalSize"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Items, 3)
	assert.Equal(t, "products", result.Items[0].CollectionName)
	require.Len(t, result.Items[0].Models, 2)
	assert.Equal(t, "model-a", result.Items[0].Models[0].InferenceModelID)
	require.Len(t, result.Items[0].Models[0].Values, 3)
	assert.InDelta(t, 5000.0, result.Items[0].Models[0].Values[1].Value, 1e-9)
	assert.Equal(t, "2026-10-02T00:00:00Z", result.Items[0].Models[0].Values[1].Timestamp)
	assert.Zero(t, result.TotalSize, "the total size is only kept with manual pagination")
}

func TestSpaceMetricsInference_RequestFields(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceInferenceMetricsCalls.Returns(sampleInference(), nil)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "inference", "space-abc",
		"--since", "2026-10-01",
		"--until", "2026-10-31",
		"--interval", "week",
		"--model", "model-a",
		"--collection", "products",
	)
	require.NoError(t, err)

	req, ok := env.ServerlessMonitoringServer.GetSpaceInferenceMetricsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), req.GetSince().AsTime())
	assert.Equal(t, time.Date(2026, 10, 31, 23, 59, 59, 0, time.UTC), req.GetUntil().AsTime())
	assert.Equal(t, serverlessmonitoringv1.InferenceMetricsInterval_INFERENCE_METRICS_INTERVAL_WEEK, req.GetInterval())
	assert.Equal(t, "model-a", req.GetInferenceModelId())
	assert.Equal(t, "products", req.GetCollectionName())
}

func TestSpaceMetricsInference_Intervals(t *testing.T) {
	cases := map[string]serverlessmonitoringv1.InferenceMetricsInterval{
		"day":   serverlessmonitoringv1.InferenceMetricsInterval_INFERENCE_METRICS_INTERVAL_DAY,
		"week":  serverlessmonitoringv1.InferenceMetricsInterval_INFERENCE_METRICS_INTERVAL_WEEK,
		"month": serverlessmonitoringv1.InferenceMetricsInterval_INFERENCE_METRICS_INTERVAL_MONTH,
	}
	for flag, want := range cases {
		t.Run(flag, func(t *testing.T) {
			env := testutil.NewTestEnv(t)
			env.ServerlessMonitoringServer.GetSpaceInferenceMetricsCalls.Returns(sampleInference(), nil)

			stdout, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "inference", "space-abc", "--interval", flag)
			require.NoError(t, err)
			assert.Contains(t, stdout, "per "+flag)

			req, ok := env.ServerlessMonitoringServer.GetSpaceInferenceMetricsCalls.Last()
			require.True(t, ok)
			assert.Equal(t, want, req.GetInterval())
		})
	}
}

func TestSpaceMetricsInference_CollectionContains(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceInferenceMetricsCalls.Returns(sampleInference(), nil)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "inference", "space-abc", "--collection-contains", "prod")
	require.NoError(t, err)

	req, ok := env.ServerlessMonitoringServer.GetSpaceInferenceMetricsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "prod", req.GetCollectionNameContains())
}

func TestSpaceMetricsInference_OnlyUntilCoversSevenDaysBefore(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceInferenceMetricsCalls.Returns(sampleInference(), nil)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "inference", "space-abc",
		"--until", "2026-10-08T00:00:00Z",
	)
	require.NoError(t, err)

	req, ok := env.ServerlessMonitoringServer.GetSpaceInferenceMetricsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), req.GetSince().AsTime())
	assert.Equal(t, time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC), req.GetUntil().AsTime())
}

func TestSpaceMetricsInference_FetchesAllPages(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceInferenceMetricsCalls.
		OnCall(0, func(_ context.Context, _ *serverlessmonitoringv1.GetSpaceInferenceMetricsRequest) (*serverlessmonitoringv1.GetSpaceInferenceMetricsResponse, error) {
			return &serverlessmonitoringv1.GetSpaceInferenceMetricsResponse{
				Items:         []*serverlessmonitoringv1.SpaceCollectionInferenceMetrics{{CollectionName: "first"}},
				TotalSize:     new(int32(2)),
				NextPageToken: new("page-2"),
			}, nil
		}).
		OnCall(1, func(_ context.Context, _ *serverlessmonitoringv1.GetSpaceInferenceMetricsRequest) (*serverlessmonitoringv1.GetSpaceInferenceMetricsResponse, error) {
			return &serverlessmonitoringv1.GetSpaceInferenceMetricsResponse{
				Items:     []*serverlessmonitoringv1.SpaceCollectionInferenceMetrics{{CollectionName: "second"}},
				TotalSize: new(int32(2)),
			}, nil
		})

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "inference", "space-abc")
	require.NoError(t, err)
	assert.Contains(t, stdout, "first")
	assert.Contains(t, stdout, "second")
	assert.Contains(t, stdout, "Total Tokens: 0")
	assert.NotContains(t, stdout, "Buckets:")

	reqs := env.ServerlessMonitoringServer.GetSpaceInferenceMetricsCalls.All()
	require.Len(t, reqs, 2)
	assert.Nil(t, reqs[0].PageToken)
	assert.Equal(t, "page-2", reqs[1].GetPageToken())
	assert.Equal(t, reqs[0].GetSince().AsTime(), reqs[1].GetSince().AsTime(), "all pages use the same period")
}

func TestSpaceMetricsInference_ManualPagination(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceInferenceMetricsCalls.Returns(&serverlessmonitoringv1.GetSpaceInferenceMetricsResponse{
		Items:         []*serverlessmonitoringv1.SpaceCollectionInferenceMetrics{{CollectionName: "first"}},
		TotalSize:     new(int32(5)),
		NextPageToken: new("page-2"),
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "inference", "space-abc",
		"--page-size", "1", "--json")
	require.NoError(t, err)

	var result struct {
		NextPageToken string `json:"nextPageToken"`
		TotalSize     int32  `json:"totalSize"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "page-2", result.NextPageToken)
	assert.Equal(t, int32(5), result.TotalSize)
	assert.Equal(t, 1, env.ServerlessMonitoringServer.GetSpaceInferenceMetricsCalls.Count())

	req, ok := env.ServerlessMonitoringServer.GetSpaceInferenceMetricsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, int32(1), req.GetPageSize())
}

func TestSpaceMetricsInference_InputErrors(t *testing.T) {
	cases := map[string][]string{
		"missing space id":    {},
		"invalid since":       {"space-abc", "--since", "yesterday"},
		"since after until":   {"space-abc", "--since", "1h", "--until", "2h"},
		"since in the future": {"space-abc", "--since", "2099-01-01"},
		"unknown interval":    {"space-abc", "--interval", "hour"},
		"empty model":         {"space-abc", "--model", ""},
		"conflicting filters": {"space-abc", "--collection", "a", "--collection-contains", "b"},
	}
	for name, extra := range cases {
		t.Run(name, func(t *testing.T) {
			env := testutil.NewTestEnv(t)

			args := append([]string{"serverless", "space", "metrics", "inference"}, extra...)
			_, _, err := testutil.Exec(t, env, args...)
			require.Error(t, err)
			assert.Equal(t, 0, env.ServerlessMonitoringServer.GetSpaceInferenceMetricsCalls.Count())
		})
	}
}

func TestSpaceMetricsInference_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.GetSpaceInferenceMetricsCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "metrics", "inference", "space-abc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get space inference metrics")
}
