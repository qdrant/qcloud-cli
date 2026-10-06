package space_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"
	spacev1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/v1"

	"github.com/qdrant/qcloud-cli/internal/resource"
	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestSpaceDescribe_TextOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.GetSpaceCalls.Returns(&spacev1.GetSpaceResponse{
		Space: &spacev1.Space{
			Id:                  "space-abc",
			Name:                "my-space",
			CloudRegionId:       "aws-eu-central-1",
			CreatedAt:           timestamppb.Now(),
			CostAllocationLabel: new("team-search"),
			Labels:              []*commonv1.KeyValue{{Key: "env", Value: "prod"}},
			Configuration: &spacev1.SpaceConfiguration{
				AllowedIpSourceRanges:  []string{"10.0.0.0/8"},
				AllowedOrigins:         []string{"https://app.example.com"},
				MaxCollectionsPerSpace: 50,
				CollectionSettings: &spacev1.CollectionSettings{
					PlatformMaxSize: uint64(100 * resource.GiB),
					MaxSize:         new(uint64(10 * resource.GiB)),
				},
				SearcherSettings: &spacev1.SearcherSettings{
					IdleTimeout:        durationpb.New(5 * time.Minute),
					PlatformMaxWorkers: 8,
					MaxWorkers:         new(uint64(2)),
				},
			},
			State: &spacev1.SpaceState{
				Phase: spacev1.SpaceStatePhase_SPACE_STATE_PHASE_READY,
				Endpoint: &spacev1.SpaceEndpoint{
					Url:      "https://space-abc.serverless.qdrant.io",
					RestPort: 443,
					GrpcPort: 6334,
				},
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "describe", "space-abc")
	require.NoError(t, err)
	assert.Contains(t, stdout, "space-abc")
	assert.Contains(t, stdout, "my-space")
	assert.Contains(t, stdout, "READY")
	assert.Contains(t, stdout, "aws-eu-central-1")
	assert.Contains(t, stdout, "https://space-abc.serverless.qdrant.io")
	assert.Contains(t, stdout, "6334")
	assert.Contains(t, stdout, "team-search")
	assert.Contains(t, stdout, "env=prod")
	assert.Contains(t, stdout, "10.0.0.0/8")
	assert.Contains(t, stdout, "https://app.example.com")
	assert.Contains(t, stdout, "10GiB")
	assert.Contains(t, stdout, "100GiB")
	assert.Contains(t, stdout, "5m0s")
	assert.NotContains(t, stdout, "Cloud:")

	req, ok := env.ServerlessSpaceServer.GetSpaceCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "space-abc", req.GetSpaceId())
}

func TestSpaceDescribe_UnsetLimits(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.GetSpaceCalls.Returns(&spacev1.GetSpaceResponse{
		Space: &spacev1.Space{
			Id:            "space-abc",
			Configuration: &spacev1.SpaceConfiguration{},
			State: &spacev1.SpaceState{
				Phase:  spacev1.SpaceStatePhase_SPACE_STATE_PHASE_DISABLED,
				Reason: "quota exceeded",
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "describe", "space-abc")
	require.NoError(t, err)
	assert.Contains(t, stdout, "DISABLED")
	assert.Contains(t, stdout, "quota exceeded")
	assert.Contains(t, stdout, "unlimited")
	assert.Contains(t, stdout, "(not set)")
	assert.NotContains(t, stdout, "Endpoint:")
}

func TestSpaceDescribe_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.GetSpaceCalls.Returns(&spacev1.GetSpaceResponse{
		Space: &spacev1.Space{Id: "space-json", Name: "json-space", CloudRegionId: "aws-eu-central-1"},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "describe", "space-json", "--json")
	require.NoError(t, err)

	var result struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		CloudRegionID string `json:"cloudRegionId"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "space-json", result.ID)
	assert.Equal(t, "json-space", result.Name)
	assert.Equal(t, "aws-eu-central-1", result.CloudRegionID)
}

func TestSpaceDescribe_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.GetSpaceCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "describe", "space-abc")
	require.Error(t, err)
}

func TestSpaceDescribe_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "describe")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "space ID")
}
