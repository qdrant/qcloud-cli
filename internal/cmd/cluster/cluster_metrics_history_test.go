package cluster_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	monitoringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/monitoring/v1"

	"github.com/qdrant/qcloud-cli/internal/qcloudapi"
	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func setupClusterHistory(env *testutil.TestEnv, cloudProvider string, resp *monitoringv1.GetClusterUsageMetricsResponse) {
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{Cluster: historyCluster(cloudProvider)}, nil)
	env.MonitoringServer.GetClusterUsageMetricsCalls.Returns(resp, nil)
}

func requestMetrics() *monitoringv1.GetClusterUsageMetricsResponse {
	return &monitoringv1.GetClusterUsageMetricsResponse{
		Rps:     series(10, 35, 20),
		Latency: series(12, 18, 15),
		Nodes:   []*monitoringv1.ClusterNodeUsageMetrics{usageNode("0")},
	}
}

func TestClusterMetricsHistory_TextOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupClusterHistory(env, "aws", requestMetrics())

	stdout, _, err := testutil.Exec(t, env, "cluster", "cluster-metrics-history", metricsClusterID,
		"--since", "2026-10-20T00:00:00Z", "--until", "2026-10-20T02:00:00Z")
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(stdout,
		fmt.Sprintf("Cluster %s  2026-10-20 00:00:00 UTC - 2026-10-20 02:00:00 UTC\n", metricsClusterID)), stdout)
	assert.Contains(t, stdout, "\nRequests\nrequests per second\n")
	assert.Contains(t, stdout, "\nLatency\nlatency\n")
	assert.Less(t, strings.Index(stdout, "Requests"), strings.Index(stdout, "Latency"))
	// 35/s and 18ms with 5% headroom round up to 4 steps of 10/s and 5ms.
	assert.Contains(t, stdout, "40/s│")
	assert.Contains(t, stdout, "20ms│")
	assert.NotContains(t, stdout, "capacity")
	assert.NotContains(t, stdout, "\x1b[")
}

func TestClusterMetricsHistory_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupClusterHistory(env, "aws", requestMetrics())

	stdout, _, err := testutil.Exec(t, env, "cluster", "cluster-metrics-history", metricsClusterID, "--json")
	require.NoError(t, err)

	var result struct {
		ClusterID         string    `json:"clusterId"`
		Since             time.Time `json:"since"`
		Until             time.Time `json:"until"`
		RequestsPerSecond []struct {
			Timestamp time.Time `json:"timestamp"`
			Value     float64   `json:"value"`
		} `json:"requestsPerSecond"`
		LatencyMs []struct {
			Timestamp time.Time `json:"timestamp"`
			Value     float64   `json:"value"`
		} `json:"latencyMs"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))

	assert.Equal(t, metricsClusterID, result.ClusterID)
	assert.Equal(t, 7*24*time.Hour, result.Until.Sub(result.Since))
	require.Len(t, result.RequestsPerSecond, 3)
	assert.True(t, result.RequestsPerSecond[1].Timestamp.Equal(historyStart.Add(time.Hour)))
	assert.InDelta(t, 35, result.RequestsPerSecond[1].Value, 1e-9)
	require.Len(t, result.LatencyMs, 3)
	assert.InDelta(t, 18, result.LatencyMs[1].Value, 1e-9)
	assert.NotContains(t, stdout, "nodes")
}

func TestClusterMetricsHistory_Requests(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupClusterHistory(env, "aws", requestMetrics())

	before := time.Now().Truncate(time.Second)
	_, _, err := testutil.Exec(t, env, "cluster", "cluster-metrics-history", metricsClusterID)
	require.NoError(t, err)
	after := time.Now()

	clusterReq, ok := env.Server.GetClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", clusterReq.GetAccountId())
	assert.Equal(t, metricsClusterID, clusterReq.GetClusterId())

	req, ok := env.MonitoringServer.GetClusterUsageMetricsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, metricsClusterID, req.GetClusterId())
	until := req.GetUntil().AsTime()
	assert.False(t, until.Before(before) || until.After(after), "until defaults to now")
	assert.Equal(t, 7*24*time.Hour, until.Sub(req.GetSince().AsTime()))
	assert.Nil(t, req.Aggregator)
}

func TestClusterMetricsHistory_TimeRange(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupClusterHistory(env, "aws", requestMetrics())

	_, _, err := testutil.Exec(t, env, "cluster", "cluster-metrics-history", metricsClusterID, "--until", "2026-10-10T00:00:00Z")
	require.NoError(t, err)

	req, ok := env.MonitoringServer.GetClusterUsageMetricsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC), req.GetSince().AsTime())
	assert.Equal(t, time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC), req.GetUntil().AsTime())

	_, _, err = testutil.Exec(t, env, "cluster", "cluster-metrics-history", metricsClusterID, "--since", "2026-10-20", "--until", "2026-10-01")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--since must be before --until")
}

func TestClusterMetricsHistory_NoData(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupClusterHistory(env, "aws", &monitoringv1.GetClusterUsageMetricsResponse{Rps: series(1, 2)})

	stdout, _, err := testutil.Exec(t, env, "cluster", "cluster-metrics-history", metricsClusterID)
	require.NoError(t, err)

	assert.Contains(t, stdout, "\nRequests\nrequests per second\n")
	assert.Contains(t, stdout, "\nLatency\nno data\n")
}

func TestClusterMetricsHistory_HybridCluster(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupClusterHistory(env, qcloudapi.HybridCloudProviderID, requestMetrics())

	_, _, err := testutil.Exec(t, env, "cluster", "cluster-metrics-history", metricsClusterID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "request metrics are not available for Hybrid Cloud clusters")
	assert.Equal(t, 0, env.MonitoringServer.GetClusterUsageMetricsCalls.Count())
}

func TestClusterMetricsHistory_MetricsError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{Cluster: historyCluster("aws")}, nil)
	env.MonitoringServer.GetClusterUsageMetricsCalls.Returns(nil, fmt.Errorf("metrics unavailable"))

	_, _, err := testutil.Exec(t, env, "cluster", "cluster-metrics-history", metricsClusterID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get cluster metrics")
}

func TestClusterMetricsHistory_ClusterError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.Server.GetClusterCalls.Returns(nil, fmt.Errorf("not found"))

	_, _, err := testutil.Exec(t, env, "cluster", "cluster-metrics-history", metricsClusterID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get cluster")
	assert.Equal(t, 0, env.MonitoringServer.GetClusterUsageMetricsCalls.Count())
}

func TestClusterMetricsHistory_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "cluster", "cluster-metrics-history")
	require.Error(t, err)
}

func TestClusterMetricsHistory_LowRateLabels(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupClusterHistory(env, "aws", &monitoringv1.GetClusterUsageMetricsResponse{
		Rps:     series(0, 0.0016, 0),
		Latency: series(0, 0.04, 0),
	})

	stdout, _, err := testutil.Exec(t, env, "cluster", "cluster-metrics-history", metricsClusterID)
	require.NoError(t, err)

	assert.Contains(t, stdout, "0.002/s│")
	assert.Contains(t, stdout, "0.001/s│")
	assert.Contains(t, stdout, "0.06ms│")
	assert.Contains(t, stdout, "0.015ms│")
	assert.NotContains(t, stdout, "0.00/s")
}

func TestClusterMetricsHistory_NoRequests(t *testing.T) {
	env := testutil.NewTestEnv(t)
	// The API fills the request series with zeros where there are none.
	setupClusterHistory(env, "aws", &monitoringv1.GetClusterUsageMetricsResponse{
		Rps:     series(0, 0, 0),
		Latency: series(0, 0, 0),
	})

	stdout, _, err := testutil.Exec(t, env, "cluster", "cluster-metrics-history", metricsClusterID)
	require.NoError(t, err)

	assert.Contains(t, stdout, "\nRequests\nno requests in this time range\n")
	assert.Contains(t, stdout, "\nLatency\nno requests in this time range\n")
}

func TestClusterMetricsHistory_FutureRange(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupClusterHistory(env, "aws", requestMetrics())

	before := time.Now()
	_, _, err := testutil.Exec(t, env, "cluster", "cluster-metrics-history", metricsClusterID,
		"--since", "24h", "--until", before.Add(48*time.Hour).UTC().Format(time.RFC3339))
	require.NoError(t, err)

	req, ok := env.MonitoringServer.GetClusterUsageMetricsCalls.Last()
	require.True(t, ok)
	assert.False(t, req.GetUntil().AsTime().After(time.Now()), "--until is cut to now")

	_, _, err = testutil.Exec(t, env, "cluster", "cluster-metrics-history", metricsClusterID,
		"--since", before.Add(time.Hour).UTC().Format(time.RFC3339))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--since must be in the past")
}
