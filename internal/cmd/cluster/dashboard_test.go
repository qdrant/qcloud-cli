package cluster_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestDashboard_PrintURL(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("acct-1"))
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{Id: "cluster-123"},
	}, nil)

	opened := 0
	env.State.SetBrowserOpener(func(string) error { opened++; return nil })

	stdout, _, err := testutil.Exec(t, env, "cluster", "dashboard", "cluster-123", "--print-url")
	require.NoError(t, err)
	assert.Contains(t, stdout, "https://cloud.qdrant.io/accounts/acct-1/clusters/cluster-123/dashboard")
	assert.Equal(t, 0, opened, "browser should not be opened with --print-url")

	req, ok := env.Server.GetClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "acct-1", req.GetAccountId())
	assert.Equal(t, "cluster-123", req.GetClusterId())
}

func TestDashboard_OpensBrowser(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("acct-1"))
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{Id: "cluster-123"},
	}, nil)

	var opened string
	calls := 0
	env.State.SetBrowserOpener(func(u string) error { opened = u; calls++; return nil })

	_, stderr, err := testutil.Exec(t, env, "cluster", "dashboard", "cluster-123")
	require.NoError(t, err)
	assert.Equal(t, 1, calls)
	assert.Equal(t, "https://cloud.qdrant.io/accounts/acct-1/clusters/cluster-123/dashboard", opened)
	assert.Contains(t, stderr, "Opening dashboard for cluster cluster-123")
}

func TestDashboard_ClusterNotFound(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.Server.GetClusterCalls.Returns(nil, status.Error(codes.NotFound, "cluster not found"))

	opened := 0
	env.State.SetBrowserOpener(func(string) error { opened++; return nil })

	_, _, err := testutil.Exec(t, env, "cluster", "dashboard", "cluster-123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cluster-123")
	assert.Equal(t, 0, opened, "browser should not be opened when the cluster is unknown")
}

func TestDashboard_MissingArgs(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "cluster", "dashboard")
	require.Error(t, err)
}

func TestDashboard_CustomConsoleURL_Env(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("acct-1"))
	t.Setenv("QDRANT_CLOUD_CONSOLE_URL", "https://cloud.stage.qdrant.io")
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{Id: "cluster-123"},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "dashboard", "cluster-123", "--print-url")
	require.NoError(t, err)
	assert.Contains(t, stdout, "https://cloud.stage.qdrant.io/accounts/acct-1/clusters/cluster-123/dashboard")
}

func TestDashboard_CustomConsoleURL_Flag(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("acct-1"))
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{Id: "cluster-123"},
	}, nil)

	// A trailing slash on the base is trimmed before joining the path.
	stdout, _, err := testutil.Exec(t, env,
		"cluster", "dashboard", "cluster-123", "--print-url",
		"--console-url", "https://console.example.com/",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "https://console.example.com/accounts/acct-1/clusters/cluster-123/dashboard")
}
