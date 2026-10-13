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

func sampleAlerts() *serverlessmonitoringv1.ListSpaceAlertsResponse {
	return &serverlessmonitoringv1.ListSpaceAlertsResponse{
		Items: []*serverlessmonitoringv1.SpaceAlert{
			{
				Id:             "alert-1",
				Type:           serverlessmonitoringv1.SpaceAlertType_SPACE_ALERT_TYPE_COLLECTION_STORAGE_OVERUTILIZED,
				Severity:       serverlessmonitoringv1.SpaceAlertSeverity_SPACE_ALERT_SEVERITY_WARNING,
				Title:          "Collection storage almost full",
				Description:    "Collection products uses 95% of its storage limit.",
				LastFiringAt:   timestamppb.New(time.Now().Add(-2 * time.Hour)),
				State:          serverlessmonitoringv1.SpaceAlertState_SPACE_ALERT_STATE_FIRING,
				CollectionName: new("products"),
			},
			{
				Id:       "alert-2",
				Type:     serverlessmonitoringv1.SpaceAlertType_SPACE_ALERT_TYPE_SPACE_UNHEALTHY,
				Severity: serverlessmonitoringv1.SpaceAlertSeverity_SPACE_ALERT_SEVERITY_CRITICAL,
				Title:    "Space unhealthy",
				State:    serverlessmonitoringv1.SpaceAlertState_SPACE_ALERT_STATE_RESOLVED,
			},
		},
	}
}

func TestSpaceAlertList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.ListSpaceAlertsCalls.Returns(sampleAlerts(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "alert", "list", "space-abc")
	require.NoError(t, err)

	for _, h := range []string{"ID", "SEVERITY", "TYPE", "STATE", "COLLECTION", "TITLE", "LAST FIRING"} {
		assert.Contains(t, stdout, h)
	}

	assert.Regexp(t, `alert-1\s+WARNING\s+COLLECTION_STORAGE_OVERUTILIZED\s+FIRING\s+products\s+Collection storage almost full\s+2 hours ago`, stdout)
	assert.Regexp(t, `alert-2\s+CRITICAL\s+SPACE_UNHEALTHY\s+RESOLVED\s+\(space\)\s+Space unhealthy\s+never`, stdout)

	req, ok := env.ServerlessMonitoringServer.ListSpaceAlertsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "space-abc", req.GetSpaceId())
	assert.Nil(t, req.State)
	assert.Nil(t, req.GetCollectionFilter())
}

func TestSpaceAlertList_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.ListSpaceAlertsCalls.Returns(sampleAlerts(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "alert", "list", "space-abc", "--json")
	require.NoError(t, err)

	var result struct {
		Items []struct {
			ID             string  `json:"id"`
			Type           string  `json:"type"`
			Severity       string  `json:"severity"`
			Description    string  `json:"description"`
			CollectionName *string `json:"collectionName"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Items, 2)
	assert.Equal(t, "alert-1", result.Items[0].ID)
	assert.Equal(t, "SPACE_ALERT_TYPE_COLLECTION_STORAGE_OVERUTILIZED", result.Items[0].Type)
	assert.Equal(t, "SPACE_ALERT_SEVERITY_WARNING", result.Items[0].Severity)
	assert.Contains(t, result.Items[0].Description, "95%")
	require.NotNil(t, result.Items[0].CollectionName)
	assert.Equal(t, "products", *result.Items[0].CollectionName)
	assert.Nil(t, result.Items[1].CollectionName)
}

func TestSpaceAlertList_Empty(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.ListSpaceAlertsCalls.Returns(&serverlessmonitoringv1.ListSpaceAlertsResponse{}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "alert", "list", "space-abc")
	require.NoError(t, err)
	assert.Contains(t, stdout, "SEVERITY")
	assert.NotContains(t, stdout, "alert-")
}

func TestSpaceAlertList_StateFlag(t *testing.T) {
	cases := map[string]serverlessmonitoringv1.SpaceAlertState{
		"firing":   serverlessmonitoringv1.SpaceAlertState_SPACE_ALERT_STATE_FIRING,
		"resolved": serverlessmonitoringv1.SpaceAlertState_SPACE_ALERT_STATE_RESOLVED,
	}
	for flag, want := range cases {
		t.Run(flag, func(t *testing.T) {
			env := testutil.NewTestEnv(t)
			env.ServerlessMonitoringServer.ListSpaceAlertsCalls.Returns(sampleAlerts(), nil)

			_, _, err := testutil.Exec(t, env, "serverless", "space", "alert", "list", "space-abc", "--state", flag)
			require.NoError(t, err)

			req, ok := env.ServerlessMonitoringServer.ListSpaceAlertsCalls.Last()
			require.True(t, ok)
			assert.Equal(t, want, req.GetState())
		})
	}
}

func TestSpaceAlertList_CollectionFilters(t *testing.T) {
	cases := map[string]struct {
		args  []string
		check func(t *testing.T, req *serverlessmonitoringv1.ListSpaceAlertsRequest)
	}{
		"exact name": {
			args: []string{"--collection", "products"},
			check: func(t *testing.T, req *serverlessmonitoringv1.ListSpaceAlertsRequest) {
				assert.Equal(t, "products", req.GetCollectionName())
			},
		},
		"contains": {
			args: []string{"--collection-contains", "prod"},
			check: func(t *testing.T, req *serverlessmonitoringv1.ListSpaceAlertsRequest) {
				assert.Equal(t, "prod", req.GetCollectionNameContains())
			},
		},
		"space only": {
			args: []string{"--space-only"},
			check: func(t *testing.T, req *serverlessmonitoringv1.ListSpaceAlertsRequest) {
				assert.True(t, req.GetSpaceGlobalOnly())
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			env := testutil.NewTestEnv(t)
			env.ServerlessMonitoringServer.ListSpaceAlertsCalls.Returns(sampleAlerts(), nil)

			args := append([]string{"serverless", "space", "alert", "list", "space-abc"}, tc.args...)
			_, _, err := testutil.Exec(t, env, args...)
			require.NoError(t, err)

			req, ok := env.ServerlessMonitoringServer.ListSpaceAlertsCalls.Last()
			require.True(t, ok)
			tc.check(t, req)
		})
	}
}

func TestSpaceAlertList_AutoPagination(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.ListSpaceAlertsCalls.
		OnCall(0, func(_ context.Context, _ *serverlessmonitoringv1.ListSpaceAlertsRequest) (*serverlessmonitoringv1.ListSpaceAlertsResponse, error) {
			return &serverlessmonitoringv1.ListSpaceAlertsResponse{
				Items:         []*serverlessmonitoringv1.SpaceAlert{{Id: "alert-1"}},
				NextPageToken: new("page-2"),
			}, nil
		}).
		OnCall(1, func(_ context.Context, _ *serverlessmonitoringv1.ListSpaceAlertsRequest) (*serverlessmonitoringv1.ListSpaceAlertsResponse, error) {
			return &serverlessmonitoringv1.ListSpaceAlertsResponse{
				Items: []*serverlessmonitoringv1.SpaceAlert{{Id: "alert-2"}},
			}, nil
		})

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "alert", "list", "space-abc")
	require.NoError(t, err)
	assert.Contains(t, stdout, "alert-1")
	assert.Contains(t, stdout, "alert-2")
	assert.Equal(t, 2, env.ServerlessMonitoringServer.ListSpaceAlertsCalls.Count())

	req, ok := env.ServerlessMonitoringServer.ListSpaceAlertsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "page-2", req.GetPageToken())
}

func TestSpaceAlertList_ManualPagination(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.ListSpaceAlertsCalls.Returns(&serverlessmonitoringv1.ListSpaceAlertsResponse{
		Items:         []*serverlessmonitoringv1.SpaceAlert{{Id: "alert-1"}},
		NextPageToken: new("page-2"),
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "alert", "list", "space-abc", "--page-size", "1", "--json")
	require.NoError(t, err)

	var result struct {
		NextPageToken string `json:"nextPageToken"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "page-2", result.NextPageToken)
	assert.Equal(t, 1, env.ServerlessMonitoringServer.ListSpaceAlertsCalls.Count())

	req, ok := env.ServerlessMonitoringServer.ListSpaceAlertsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, int32(1), req.GetPageSize())
}

func TestSpaceAlertList_InputErrors(t *testing.T) {
	cases := map[string][]string{
		"missing space id":     {},
		"unknown state":        {"space-abc", "--state", "pending"},
		"collection and space": {"space-abc", "--collection", "a", "--space-only"},
		"contains and space":   {"space-abc", "--collection-contains", "a", "--space-only"},
	}
	for name, extra := range cases {
		t.Run(name, func(t *testing.T) {
			env := testutil.NewTestEnv(t)

			args := append([]string{"serverless", "space", "alert", "list"}, extra...)
			_, _, err := testutil.Exec(t, env, args...)
			require.Error(t, err)
			assert.Equal(t, 0, env.ServerlessMonitoringServer.ListSpaceAlertsCalls.Count())
		})
	}
}

func TestSpaceAlertList_SpaceOnlyFalseWithCollection(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.ListSpaceAlertsCalls.Returns(sampleAlerts(), nil)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "alert", "list", "space-abc", "--space-only=false", "--collection", "products")
	require.NoError(t, err)

	req, ok := env.ServerlessMonitoringServer.ListSpaceAlertsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "products", req.GetCollectionName())
}

func TestSpaceAlertList_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessMonitoringServer.ListSpaceAlertsCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "alert", "list", "space-abc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list space alerts")
}

func TestSpaceAlert_Help(t *testing.T) {
	env := testutil.NewTestEnv(t)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "--help")
	require.NoError(t, err)
	assert.Contains(t, stdout, "alert")
}
