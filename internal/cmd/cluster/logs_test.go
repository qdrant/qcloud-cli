package cluster_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	monitoringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/monitoring/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func logEntry(ts time.Time, msg string) *monitoringv1.LogEntry {
	return &monitoringv1.LogEntry{
		Timestamp: timestamppb.New(ts),
		Message:   msg,
	}
}

func TestClusterLogs_TextOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.MonitoringServer.GetClusterLogsCalls.Returns(&monitoringv1.GetClusterLogsResponse{
		Items: []*monitoringv1.LogEntry{
			logEntry(time.Now(), "Starting Qdrant server"),
			logEntry(time.Now(), "Loaded 3 collections"),
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "logs", "my-cluster")
	require.NoError(t, err)

	assert.Contains(t, stdout, "Starting Qdrant server")
	assert.Contains(t, stdout, "Loaded 3 collections")
}

func TestClusterLogs_NoTimestampByDefault(t *testing.T) {
	env := testutil.NewTestEnv(t)

	ts := time.Date(2024, 1, 15, 10, 23, 45, 0, time.UTC)
	env.MonitoringServer.GetClusterLogsCalls.Returns(&monitoringv1.GetClusterLogsResponse{
		Items: []*monitoringv1.LogEntry{
			logEntry(ts, "some message"),
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "logs", "my-cluster")
	require.NoError(t, err)

	assert.Contains(t, stdout, "some message")
	assert.NotContains(t, stdout, "2024-01-15")
}

func TestClusterLogs_TimestampsFlag(t *testing.T) {
	env := testutil.NewTestEnv(t)

	ts := time.Date(2024, 1, 15, 10, 23, 45, 0, time.UTC)
	env.MonitoringServer.GetClusterLogsCalls.Returns(&monitoringv1.GetClusterLogsResponse{
		Items: []*monitoringv1.LogEntry{
			logEntry(ts, "some message"),
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "logs", "my-cluster", "--timestamps")
	require.NoError(t, err)

	assert.Contains(t, stdout, "2024-01-15 10:23:45 UTC")
	assert.Contains(t, stdout, "some message")
}

func TestClusterLogs_TimestampsShorthand(t *testing.T) {
	env := testutil.NewTestEnv(t)

	ts := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	env.MonitoringServer.GetClusterLogsCalls.Returns(&monitoringv1.GetClusterLogsResponse{
		Items: []*monitoringv1.LogEntry{
			logEntry(ts, "shorthand check"),
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "logs", "my-cluster", "-t")
	require.NoError(t, err)

	assert.Contains(t, stdout, "2024-06-01")
	assert.Contains(t, stdout, "shorthand check")
}

func TestClusterLogs_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.MonitoringServer.GetClusterLogsCalls.Returns(&monitoringv1.GetClusterLogsResponse{
		Items: []*monitoringv1.LogEntry{
			logEntry(time.Now(), "json log line"),
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "logs", "my-cluster", "--json")
	require.NoError(t, err)

	assert.Contains(t, stdout, `"items"`)
	assert.Contains(t, stdout, `"message"`)
	assert.Contains(t, stdout, "json log line")
}

func TestClusterLogs_EmptyResponse(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.MonitoringServer.GetClusterLogsCalls.Returns(&monitoringv1.GetClusterLogsResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "logs", "my-cluster")
	require.NoError(t, err)

	assert.Empty(t, stdout)
}

func TestClusterLogs_AccountIDPassedToServer(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.MonitoringServer.GetClusterLogsCalls.Returns(&monitoringv1.GetClusterLogsResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "logs", "my-cluster")
	require.NoError(t, err)

	req, ok := env.MonitoringServer.GetClusterLogsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
}

func TestClusterLogs_ClusterIDPassedToServer(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.MonitoringServer.GetClusterLogsCalls.Returns(&monitoringv1.GetClusterLogsResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "logs", "target-cluster")
	require.NoError(t, err)

	req, ok := env.MonitoringServer.GetClusterLogsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "target-cluster", req.GetClusterId())
}

func TestClusterLogs_SinceFlag_RFC3339(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.MonitoringServer.GetClusterLogsCalls.Returns(&monitoringv1.GetClusterLogsResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "logs", "my-cluster", "--since", "2024-01-01T00:00:00Z")
	require.NoError(t, err)

	req, ok := env.MonitoringServer.GetClusterLogsCalls.Last()
	require.True(t, ok)
	require.NotNil(t, req.GetSince())
	assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), req.GetSince().AsTime())
}

func TestClusterLogs_SinceFlag_DateOnly(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.MonitoringServer.GetClusterLogsCalls.Returns(&monitoringv1.GetClusterLogsResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "logs", "my-cluster", "-s", "2024-03-15")
	require.NoError(t, err)

	req, ok := env.MonitoringServer.GetClusterLogsCalls.Last()
	require.True(t, ok)
	require.NotNil(t, req.GetSince())
	assert.Equal(t, time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), req.GetSince().AsTime())
}

func TestClusterLogs_UntilFlag(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.MonitoringServer.GetClusterLogsCalls.Returns(&monitoringv1.GetClusterLogsResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "logs", "my-cluster", "-u", "2024-02-01T12:00:00Z")
	require.NoError(t, err)

	req, ok := env.MonitoringServer.GetClusterLogsCalls.Last()
	require.True(t, ok)
	require.NotNil(t, req.GetUntil())
	assert.Equal(t, time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC), req.GetUntil().AsTime())
}

func TestClusterLogs_NoSinceByDefault(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.MonitoringServer.GetClusterLogsCalls.Returns(&monitoringv1.GetClusterLogsResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "logs", "my-cluster")
	require.NoError(t, err)

	req, ok := env.MonitoringServer.GetClusterLogsCalls.Last()
	require.True(t, ok)
	assert.Nil(t, req.GetSince())
	assert.Nil(t, req.GetUntil())
}

func TestClusterLogs_InvalidSince(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "cluster", "logs", "my-cluster", "--since", "not-a-date")
	require.Error(t, err)
}

func TestClusterLogs_InvalidUntil(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "cluster", "logs", "my-cluster", "--until", "not-a-date")
	require.Error(t, err)
}

func TestClusterLogs_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "cluster", "logs")
	require.Error(t, err)
}
