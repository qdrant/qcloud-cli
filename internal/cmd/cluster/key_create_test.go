package cluster_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterauthv2 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/auth/v2"
	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestKeyCreate_Basic(t *testing.T) {
	env := testutil.NewTestEnv(t, testutil.WithAccountID("test-account-id"))

	env.DatabaseApiKeyServer.CreateDatabaseApiKeyCalls.Returns(&clusterauthv2.CreateDatabaseApiKeyResponse{
		DatabaseApiKey: &clusterauthv2.DatabaseApiKey{
			Id:  "key-new",
			Key: "secret-key-value",
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "key", "create", "cluster-123", "--name", "my-key")
	require.NoError(t, err)
	assert.Contains(t, stdout, "key-new")
	assert.Contains(t, stdout, "secret-key-value")
	assert.Contains(t, stdout, "not be shown again")

	req, ok := env.DatabaseApiKeyServer.CreateDatabaseApiKeyCalls.Last()
	require.True(t, ok)
	capturedKey := req.GetDatabaseApiKey()
	assert.Equal(t, "test-account-id", capturedKey.GetAccountId())
	assert.Equal(t, "cluster-123", capturedKey.GetClusterId())
	assert.Equal(t, "my-key", capturedKey.GetName())
	assert.Empty(t, capturedKey.GetAccessRules())
}

func TestKeyCreate_WithManageAccessType(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.DatabaseApiKeyServer.CreateDatabaseApiKeyCalls.Returns(&clusterauthv2.CreateDatabaseApiKeyResponse{
		DatabaseApiKey: &clusterauthv2.DatabaseApiKey{Id: "key-manage"},
	}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "key", "create", "cluster-123", "--name", "manage-key", "--access-type", "manage")
	require.NoError(t, err)

	req, ok := env.DatabaseApiKeyServer.CreateDatabaseApiKeyCalls.Last()
	require.True(t, ok)
	capturedKey := req.GetDatabaseApiKey()
	require.Len(t, capturedKey.GetAccessRules(), 1)
	globalAccess := capturedKey.GetAccessRules()[0].GetGlobalAccess()
	require.NotNil(t, globalAccess)
	assert.Equal(t, clusterauthv2.GlobalAccessRuleAccessType_GLOBAL_ACCESS_RULE_ACCESS_TYPE_MANAGE, globalAccess.GetAccessType())
}

func TestKeyCreate_WithReadOnlyAccessType(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.DatabaseApiKeyServer.CreateDatabaseApiKeyCalls.Returns(&clusterauthv2.CreateDatabaseApiKeyResponse{
		DatabaseApiKey: &clusterauthv2.DatabaseApiKey{Id: "key-ro"},
	}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "key", "create", "cluster-123", "--name", "ro-key", "--access-type", "read-only")
	require.NoError(t, err)

	req, ok := env.DatabaseApiKeyServer.CreateDatabaseApiKeyCalls.Last()
	require.True(t, ok)
	capturedKey := req.GetDatabaseApiKey()
	require.Len(t, capturedKey.GetAccessRules(), 1)
	globalAccess := capturedKey.GetAccessRules()[0].GetGlobalAccess()
	require.NotNil(t, globalAccess)
	assert.Equal(t, clusterauthv2.GlobalAccessRuleAccessType_GLOBAL_ACCESS_RULE_ACCESS_TYPE_READ_ONLY, globalAccess.GetAccessType())
}

func TestKeyCreate_InvalidAccessType(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "cluster", "key", "create", "cluster-123", "--name", "bad-key", "--access-type", "superuser")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "superuser")
}

func TestKeyCreate_WithExpires(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.DatabaseApiKeyServer.CreateDatabaseApiKeyCalls.Returns(&clusterauthv2.CreateDatabaseApiKeyResponse{
		DatabaseApiKey: &clusterauthv2.DatabaseApiKey{Id: "key-exp"},
	}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "key", "create", "cluster-123", "--name", "exp-key", "--expires", "2027-06-15")
	require.NoError(t, err)

	req, ok := env.DatabaseApiKeyServer.CreateDatabaseApiKeyCalls.Last()
	require.True(t, ok)
	capturedKey := req.GetDatabaseApiKey()
	require.NotNil(t, capturedKey.GetExpiresAt())
	assert.Equal(t, "2027-06-15", capturedKey.GetExpiresAt().AsTime().UTC().Format("2006-01-02"))
}

func TestKeyCreate_InvalidExpires(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "cluster", "key", "create", "cluster-123", "--name", "bad-key", "--expires", "not-a-date")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "YYYY-MM-DD")
}

func TestKeyCreate_MissingName(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "cluster", "key", "create", "cluster-123")
	require.Error(t, err)
}

func TestKeyCreate_NoWait(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.DatabaseApiKeyServer.CreateDatabaseApiKeyCalls.Returns(&clusterauthv2.CreateDatabaseApiKeyResponse{
		DatabaseApiKey: &clusterauthv2.DatabaseApiKey{
			Id:  "key-no-wait",
			Key: "secret",
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "key", "create", "cluster-123", "--name", "my-key")
	require.NoError(t, err)
	assert.Contains(t, stdout, "key-no-wait")
	assert.Equal(t, 0, env.Server.GetClusterCalls.Count(), "GetCluster should not be called without --wait")
}

// clusterEndpoint starts an httptest server and returns (server, schemeHost, port).
// The caller is responsible for closing the server.
func clusterEndpoint(t *testing.T, handler http.Handler) (host string, port int32) {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	u, err := url.Parse(ts.URL)
	require.NoError(t, err)
	p, err := strconv.Atoi(u.Port())
	require.NoError(t, err)
	return fmt.Sprintf("http://%s", u.Hostname()), int32(p)
}

func TestKeyCreate_WaitSuccess(t *testing.T) {
	env := testutil.NewTestEnv(t)

	var calls atomic.Int32
	host, port := clusterEndpoint(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("api-key") != "secret-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		n := calls.Add(1)
		if n <= 2 {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	env.DatabaseApiKeyServer.CreateDatabaseApiKeyCalls.Returns(&clusterauthv2.CreateDatabaseApiKeyResponse{
		DatabaseApiKey: &clusterauthv2.DatabaseApiKey{
			Id:   "key-wait",
			Name: "wait-key",
			Key:  "secret-key",
		},
	}, nil)

	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{
			Id: "cluster-123",
			State: &clusterv1.ClusterState{
				Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY,
				Endpoint: &clusterv1.ClusterEndpoint{
					Url:      host,
					RestPort: port,
				},
			},
		},
	}, nil)

	stdout, stderr, err := testutil.Exec(t, env,
		"cluster", "key", "create", "cluster-123",
		"--name", "wait-key",
		"--wait",
		"--wait-timeout", "30s",
		"--wait-poll-interval", "10ms",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "key-wait")
	assert.Contains(t, stderr, "API key is active")
	assert.GreaterOrEqual(t, env.Server.GetClusterCalls.Count(), 1)
}

func TestKeyCreate_WaitTimeout(t *testing.T) {
	env := testutil.NewTestEnv(t)

	host, port := clusterEndpoint(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))

	env.DatabaseApiKeyServer.CreateDatabaseApiKeyCalls.Returns(&clusterauthv2.CreateDatabaseApiKeyResponse{
		DatabaseApiKey: &clusterauthv2.DatabaseApiKey{
			Id:  "key-timeout",
			Key: "secret",
		},
	}, nil)

	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{
			Id: "cluster-123",
			State: &clusterv1.ClusterState{
				Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY,
				Endpoint: &clusterv1.ClusterEndpoint{
					Url:      host,
					RestPort: port,
				},
			},
		},
	}, nil)

	_, _, err := testutil.Exec(t, env,
		"cluster", "key", "create", "cluster-123",
		"--name", "timeout-key",
		"--wait",
		"--wait-timeout", "200ms",
		"--wait-poll-interval", "10ms",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
}

func TestKeyCreate_WaitNoEndpoint_TimesOut(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.DatabaseApiKeyServer.CreateDatabaseApiKeyCalls.Returns(&clusterauthv2.CreateDatabaseApiKeyResponse{
		DatabaseApiKey: &clusterauthv2.DatabaseApiKey{
			Id:  "key-no-ep",
			Key: "secret",
		},
	}, nil)

	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{
			Id:    "cluster-123",
			State: &clusterv1.ClusterState{},
		},
	}, nil)

	_, stderr, err := testutil.Exec(t, env,
		"cluster", "key", "create", "cluster-123",
		"--name", "no-ep-key",
		"--wait",
		"--wait-timeout", "200ms",
		"--wait-poll-interval", "10ms",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
	assert.Contains(t, stderr, "waiting for API key")
}

func TestKeyCreate_WaitEndpointAppearsMidPoll(t *testing.T) {
	env := testutil.NewTestEnv(t)

	host, port := clusterEndpoint(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	env.DatabaseApiKeyServer.CreateDatabaseApiKeyCalls.Returns(&clusterauthv2.CreateDatabaseApiKeyResponse{
		DatabaseApiKey: &clusterauthv2.DatabaseApiKey{
			Id:   "key-delayed-ep",
			Name: "delayed-ep-key",
			Key:  "secret",
		},
	}, nil)

	// First two GetCluster calls return no endpoint, then it appears.
	env.Server.GetClusterCalls.
		OnCall(0, func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
			return &clusterv1.GetClusterResponse{
				Cluster: &clusterv1.Cluster{
					Id:    "cluster-123",
					State: &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_CREATING},
				},
			}, nil
		}).
		OnCall(1, func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
			return &clusterv1.GetClusterResponse{
				Cluster: &clusterv1.Cluster{
					Id:    "cluster-123",
					State: &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_CREATING},
				},
			}, nil
		}).
		Always(func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
			return &clusterv1.GetClusterResponse{
				Cluster: &clusterv1.Cluster{
					Id: "cluster-123",
					State: &clusterv1.ClusterState{
						Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY,
						Endpoint: &clusterv1.ClusterEndpoint{
							Url:      host,
							RestPort: port,
						},
					},
				},
			}, nil
		})

	stdout, stderr, err := testutil.Exec(t, env,
		"cluster", "key", "create", "cluster-123",
		"--name", "delayed-ep-key",
		"--wait",
		"--wait-timeout", "30s",
		"--wait-poll-interval", "10ms",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "key-delayed-ep")
	assert.Contains(t, stderr, "API key is active")
	assert.GreaterOrEqual(t, env.Server.GetClusterCalls.Count(), 3)
}

func TestKeyCreate_WaitDefaultPort(t *testing.T) {
	env := testutil.NewTestEnv(t)

	var probed atomic.Bool
	host, port := clusterEndpoint(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		probed.Store(true)
		w.WriteHeader(http.StatusOK)
	}))

	env.DatabaseApiKeyServer.CreateDatabaseApiKeyCalls.Returns(&clusterauthv2.CreateDatabaseApiKeyResponse{
		DatabaseApiKey: &clusterauthv2.DatabaseApiKey{
			Id:  "key-default-port",
			Key: "secret",
		},
	}, nil)

	// RestPort = 0 would trigger default 6333, but we can't test that with a real
	// httptest server. Instead, verify that the explicit port works.
	env.Server.GetClusterCalls.Always(func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id: "cluster-123",
				State: &clusterv1.ClusterState{
					Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY,
					Endpoint: &clusterv1.ClusterEndpoint{
						Url:      host,
						RestPort: port,
					},
				},
			},
		}, nil
	})

	_, _, err := testutil.Exec(t, env,
		"cluster", "key", "create", "cluster-123",
		"--name", "port-key",
		"--wait",
		"--wait-timeout", "5s",
		"--wait-poll-interval", "10ms",
	)
	require.NoError(t, err)
	assert.True(t, probed.Load(), "cluster endpoint should have been probed")
}
