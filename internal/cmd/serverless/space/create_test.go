package space_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spacev1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/v1"

	"github.com/qdrant/qcloud-cli/internal/resource"
	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func echoCreateSpace(env *testutil.TestEnv) {
	env.ServerlessSpaceServer.CreateSpaceCalls.Always(func(_ context.Context, req *spacev1.CreateSpaceRequest) (*spacev1.CreateSpaceResponse, error) {
		sp := req.GetSpace()
		sp.Id = "space-new"
		return &spacev1.CreateSpaceResponse{Space: sp}, nil
	})
}

func TestSpaceCreate_WithName(t *testing.T) {
	env := testutil.NewTestEnv(t)
	echoCreateSpace(env)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "create",
		"--cloud-region", "eu-central-1",
		"--name", "my-space",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "space-new")
	assert.Contains(t, stdout, "my-space")
	assert.Contains(t, stdout, "created")
	assert.Equal(t, 0, env.ServerlessSpaceServer.SuggestSpaceNameCalls.Count())

	req, ok := env.ServerlessSpaceServer.CreateSpaceCalls.Last()
	require.True(t, ok)
	sp := req.GetSpace()
	assert.Equal(t, "test-account-id", sp.GetAccountId())
	assert.Equal(t, "my-space", sp.GetName())
	assert.Equal(t, "eu-central-1", sp.GetCloudRegionId())
	assert.Empty(t, sp.GetLabels())
	assert.Nil(t, sp.CostAllocationLabel)
	assert.Nil(t, sp.GetConfiguration().GetCollectionSettings())
	assert.Nil(t, sp.GetConfiguration().GetSearcherSettings())
}

func TestSpaceCreate_SuggestsNameWhenOmitted(t *testing.T) {
	env := testutil.NewTestEnv(t)
	echoCreateSpace(env)

	env.ServerlessSpaceServer.SuggestSpaceNameCalls.Returns(&spacev1.SuggestSpaceNameResponse{Name: "brave-falcon"}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "create", "--cloud-region", "eu-central-1")
	require.NoError(t, err)
	assert.Contains(t, stdout, "brave-falcon")

	suggestReq, ok := env.ServerlessSpaceServer.SuggestSpaceNameCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", suggestReq.GetAccountId())

	req, ok := env.ServerlessSpaceServer.CreateSpaceCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "brave-falcon", req.GetSpace().GetName())
}

func TestSpaceCreate_SuggestNameError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.SuggestSpaceNameCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "create", "--cloud-region", "eu-central-1")
	require.Error(t, err)
	assert.Equal(t, 0, env.ServerlessSpaceServer.CreateSpaceCalls.Count())
}

func TestSpaceCreate_AllFlags(t *testing.T) {
	env := testutil.NewTestEnv(t)
	echoCreateSpace(env)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "create",
		"--cloud-region", "eu-central-1",
		"--name", "my-space",
		"--label", "env=prod",
		"--label", "team=search",
		"--cost-allocation-label", "cc-42",
		"--allowed-ip", "10.0.0.0/8",
		"--allowed-ip", "192.168.0.0/16",
		"--allowed-origin", "https://app.example.com",
		"--max-collection-size", "10GiB",
		"--searcher-idle-timeout", "10m",
		"--searcher-max-workers", "4",
	)
	require.NoError(t, err)

	req, ok := env.ServerlessSpaceServer.CreateSpaceCalls.Last()
	require.True(t, ok)
	sp := req.GetSpace()
	require.Len(t, sp.GetLabels(), 2)
	assert.Equal(t, "env", sp.GetLabels()[0].GetKey())
	assert.Equal(t, "prod", sp.GetLabels()[0].GetValue())
	assert.Equal(t, "team", sp.GetLabels()[1].GetKey())
	assert.Equal(t, "cc-42", sp.GetCostAllocationLabel())

	cfg := sp.GetConfiguration()
	assert.Equal(t, []string{"10.0.0.0/8", "192.168.0.0/16"}, cfg.GetAllowedIpSourceRanges())
	assert.Equal(t, []string{"https://app.example.com"}, cfg.GetAllowedOrigins())
	assert.Equal(t, uint64(10*resource.GiB), cfg.GetCollectionSettings().GetMaxSize())
	assert.Equal(t, 10*time.Minute, cfg.GetSearcherSettings().GetIdleTimeout().AsDuration())
	assert.Equal(t, uint64(4), cfg.GetSearcherSettings().GetMaxWorkers())
}

func TestSpaceCreate_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	echoCreateSpace(env)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "create",
		"--cloud-region", "eu-central-1", "--name", "my-space", "--json",
	)
	require.NoError(t, err)

	var result struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "space-new", result.ID)
	assert.Equal(t, "my-space", result.Name)
}

func TestSpaceCreate_Wait(t *testing.T) {
	env := testutil.NewTestEnv(t)
	echoCreateSpace(env)

	env.ServerlessSpaceServer.GetSpaceCalls.
		OnCall(0, func(_ context.Context, _ *spacev1.GetSpaceRequest) (*spacev1.GetSpaceResponse, error) {
			return &spacev1.GetSpaceResponse{Space: &spacev1.Space{
				Id:    "space-new",
				State: &spacev1.SpaceState{Phase: spacev1.SpaceStatePhase_SPACE_STATE_PHASE_PROCESSING},
			}}, nil
		}).
		Always(func(_ context.Context, req *spacev1.GetSpaceRequest) (*spacev1.GetSpaceResponse, error) {
			return &spacev1.GetSpaceResponse{Space: &spacev1.Space{
				Id:   req.GetSpaceId(),
				Name: "my-space",
				State: &spacev1.SpaceState{
					Phase:    spacev1.SpaceStatePhase_SPACE_STATE_PHASE_READY,
					Endpoint: &spacev1.SpaceEndpoint{Url: "https://space-new.serverless.qdrant.io"},
				},
			}}, nil
		})

	stdout, stderr, err := testutil.Exec(t, env, "serverless", "space", "create",
		"--cloud-region", "eu-central-1", "--name", "my-space",
		"--wait", "--wait-timeout", "30s", "--wait-poll-interval", "10ms",
	)
	require.NoError(t, err)
	assert.Contains(t, stderr, "phase=PROCESSING")
	assert.Contains(t, stdout, "is ready")
	assert.Contains(t, stdout, "https://space-new.serverless.qdrant.io")

	req, ok := env.ServerlessSpaceServer.GetSpaceCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "space-new", req.GetSpaceId())
}

func TestSpaceCreate_MissingCloudRegion(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "create", "--name", "my-space")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cloud-region")
	assert.Equal(t, 0, env.ServerlessSpaceServer.CreateSpaceCalls.Count())
}

func TestSpaceCreate_InvalidFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "invalid label", args: []string{"--label", "novalue"}},
		{name: "empty allowed ip", args: []string{"--allowed-ip", "-"}},
		{name: "empty allowed origin", args: []string{"--allowed-origin", "-"}},
		{name: "invalid size", args: []string{"--max-collection-size", "lots"}},
		{name: "zero size", args: []string{"--max-collection-size", "0"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := testutil.NewTestEnv(t)

			args := append([]string{"serverless", "space", "create", "--cloud-region", "r", "--name", "my-space"}, tt.args...)
			_, _, err := testutil.Exec(t, env, args...)
			require.Error(t, err)
			assert.Equal(t, 0, env.ServerlessSpaceServer.CreateSpaceCalls.Count())
		})
	}
}

func TestSpaceCreate_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.CreateSpaceCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "create",
		"--cloud-region", "eu-central-1", "--name", "my-space",
	)
	require.Error(t, err)
}
