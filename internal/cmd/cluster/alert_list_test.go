package cluster_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	monitoringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/monitoring/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func sampleClusterAlerts() *monitoringv1.ListClusterAlertsResponse {
	return &monitoringv1.ListClusterAlertsResponse{
		Alerts: []*monitoringv1.ClusterAlert{
			{
				Id:           "alert-1",
				Type:         monitoringv1.ClusterAlertType_CLUSTER_ALERT_TYPE_DISK_OVERUTILIZED,
				Severity:     monitoringv1.ClusterAlertSeverity_CLUSTER_ALERT_SEVERITY_WARNING,
				Title:        "Disk almost full",
				Description:  "Disk usage on node-0 has been above 92% for the last 15 minutes. If the disk fills up, the node goes into recovery mode.",
				LastFiringAt: timestamppb.New(time.Now().Add(-2 * time.Hour)),
				State:        monitoringv1.ClusterAlertState_CLUSTER_ALERT_STATE_FIRING,
			},
			{
				Id:       "alert-2",
				Type:     monitoringv1.ClusterAlertType_CLUSTER_ALERT_TYPE_QDRANT_CLUSTER_UNHEALTHY,
				Severity: monitoringv1.ClusterAlertSeverity_CLUSTER_ALERT_SEVERITY_CRITICAL,
				Title:    "Cluster unhealthy",
				State:    monitoringv1.ClusterAlertState_CLUSTER_ALERT_STATE_RESOLVED,
			},
		},
	}
}

func TestClusterAlertList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.MonitoringServer.ListClusterAlertsCalls.Returns(sampleClusterAlerts(), nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "alert", "list", "cluster-abc")
	require.NoError(t, err)

	for _, h := range []string{"SEVERITY", "TYPE", "STATE", "TITLE", "LAST FIRING", "DESCRIPTION"} {
		assert.Contains(t, stdout, h)
	}

	assert.Regexp(t, `WARNING\s+DISK_OVERUTILIZED\s+FIRING\s+Disk almost full\s+2 hours ago\s+Disk usage on node-0 has been above 92% for the last 15\s*\n\s+minutes\. If the disk fills up, the node goes into recovery\s*\n\s+mode\.\s*\n`, stdout)
	assert.Regexp(t, `CRITICAL\s+QDRANT_CLUSTER_UNHEALTHY\s+RESOLVED\s+Cluster unhealthy\s+never\s*\n`, stdout)
	assert.NotContains(t, stdout, "alert-1")
	assert.NotContains(t, stdout, "alert-2")

	req, ok := env.MonitoringServer.ListClusterAlertsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "cluster-abc", req.GetClusterId())
	require.NotNil(t, req.State)
	assert.Equal(t, monitoringv1.ClusterAlertState_CLUSTER_ALERT_STATE_FIRING, req.GetState())
}

func TestClusterAlertList_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.MonitoringServer.ListClusterAlertsCalls.Returns(sampleClusterAlerts(), nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "alert", "list", "cluster-abc", "--json")
	require.NoError(t, err)

	var result struct {
		Alerts []struct {
			ID          string `json:"id"`
			Type        string `json:"type"`
			Severity    string `json:"severity"`
			State       string `json:"state"`
			Description string `json:"description"`
		} `json:"alerts"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Alerts, 2)
	assert.Equal(t, "alert-1", result.Alerts[0].ID)
	assert.Equal(t, "CLUSTER_ALERT_TYPE_DISK_OVERUTILIZED", result.Alerts[0].Type)
	assert.Equal(t, "CLUSTER_ALERT_SEVERITY_WARNING", result.Alerts[0].Severity)
	assert.Equal(t, "CLUSTER_ALERT_STATE_FIRING", result.Alerts[0].State)
	assert.Contains(t, result.Alerts[0].Description, "92%")
	assert.Equal(t, "alert-2", result.Alerts[1].ID)
}

func TestClusterAlertList_Empty(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.MonitoringServer.ListClusterAlertsCalls.Returns(&monitoringv1.ListClusterAlertsResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "alert", "list", "cluster-abc")
	require.NoError(t, err)
	assert.Contains(t, stdout, "SEVERITY")
	assert.NotContains(t, stdout, "WARNING")
	assert.NotContains(t, stdout, "CRITICAL")
}

func TestClusterAlertList_NoHeaders(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.MonitoringServer.ListClusterAlertsCalls.Returns(sampleClusterAlerts(), nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "alert", "list", "cluster-abc", "--no-headers")
	require.NoError(t, err)
	assert.NotContains(t, stdout, "SEVERITY")
	assert.Regexp(t, `^WARNING\s+DISK_OVERUTILIZED`, stdout)
}

func TestClusterAlertList_ShowResolved(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.MonitoringServer.ListClusterAlertsCalls.Returns(sampleClusterAlerts(), nil)

	_, _, err := testutil.Exec(t, env, "cluster", "alert", "list", "cluster-abc", "--show-resolved")
	require.NoError(t, err)

	req, ok := env.MonitoringServer.ListClusterAlertsCalls.Last()
	require.True(t, ok)
	assert.Nil(t, req.State)
}

func TestClusterAlertList_ShowResolvedFalse(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.MonitoringServer.ListClusterAlertsCalls.Returns(sampleClusterAlerts(), nil)

	_, _, err := testutil.Exec(t, env, "cluster", "alert", "list", "cluster-abc", "--show-resolved=false")
	require.NoError(t, err)

	req, ok := env.MonitoringServer.ListClusterAlertsCalls.Last()
	require.True(t, ok)
	require.NotNil(t, req.State)
	assert.Equal(t, monitoringv1.ClusterAlertState_CLUSTER_ALERT_STATE_FIRING, req.GetState())
}

func TestClusterAlertList_InputErrors(t *testing.T) {
	cases := map[string][]string{
		"missing cluster id": {},
		"too many args":      {"cluster-abc", "cluster-def"},
		"removed state flag": {"cluster-abc", "--state", "firing"},
		"unknown flag":       {"cluster-abc", "--collection", "products"},
	}
	for name, extra := range cases {
		t.Run(name, func(t *testing.T) {
			env := testutil.NewTestEnv(t)

			args := append([]string{"cluster", "alert", "list"}, extra...)
			_, _, err := testutil.Exec(t, env, args...)
			require.Error(t, err)
			assert.Equal(t, 0, env.MonitoringServer.ListClusterAlertsCalls.Count())
		})
	}
}

func TestClusterAlertList_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.MonitoringServer.ListClusterAlertsCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "cluster", "alert", "list", "cluster-abc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list cluster alerts")
}

func TestClusterAlert_Help(t *testing.T) {
	env := testutil.NewTestEnv(t)

	stdout, _, err := testutil.Exec(t, env, "cluster", "--help")
	require.NoError(t, err)
	assert.Contains(t, stdout, "alert")
}
