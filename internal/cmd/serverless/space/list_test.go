package space_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	spacev1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestSpaceList_TableOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.ListSpacesCalls.Returns(&spacev1.ListSpacesResponse{
		Items: []*spacev1.Space{
			{
				Id:            "space-1",
				Name:          "my-space",
				CloudRegionId: "eu-central-1",
				CreatedAt:     timestamppb.Now(),
				State: &spacev1.SpaceState{
					Phase:    spacev1.SpaceStatePhase_SPACE_STATE_PHASE_READY,
					Endpoint: &spacev1.SpaceEndpoint{Url: "https://space-1.serverless.qdrant.io"},
				},
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "list")
	require.NoError(t, err)
	for _, h := range []string{"ID", "NAME", "PHASE", "REGION", "ENDPOINT", "CREATED"} {
		assert.Contains(t, stdout, h)
	}

	assert.Contains(t, stdout, "space-1")
	assert.Contains(t, stdout, "my-space")
	assert.Contains(t, stdout, "READY")
	assert.Contains(t, stdout, "eu-central-1")
	assert.Contains(t, stdout, "https://space-1.serverless.qdrant.io")

	req, ok := env.ServerlessSpaceServer.ListSpacesCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Nil(t, req.CloudRegionId)
	assert.Nil(t, req.PageSize)
}

func TestSpaceList_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.ListSpacesCalls.Returns(&spacev1.ListSpacesResponse{
		Items: []*spacev1.Space{{Id: "space-json", Name: "json-space"}},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "list", "--json")
	require.NoError(t, err)

	var result struct {
		Items []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Items, 1)
	assert.Equal(t, "space-json", result.Items[0].ID)
	assert.Equal(t, "json-space", result.Items[0].Name)
}

func TestSpaceList_CloudRegionFilter(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.ListSpacesCalls.Returns(&spacev1.ListSpacesResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "list", "--cloud-region", "us-east4")
	require.NoError(t, err)

	req, ok := env.ServerlessSpaceServer.ListSpacesCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "us-east4", req.GetCloudRegionId())
}

func TestSpaceList_AutoPaginates(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.ListSpacesCalls.
		OnCall(0, func(_ context.Context, _ *spacev1.ListSpacesRequest) (*spacev1.ListSpacesResponse, error) {
			return &spacev1.ListSpacesResponse{
				Items:         []*spacev1.Space{{Id: "space-page-1"}},
				NextPageToken: new("token-2"),
			}, nil
		}).
		OnCall(1, func(_ context.Context, req *spacev1.ListSpacesRequest) (*spacev1.ListSpacesResponse, error) {
			assert.Equal(t, "token-2", req.GetPageToken())
			return &spacev1.ListSpacesResponse{
				Items: []*spacev1.Space{{Id: "space-page-2"}},
			}, nil
		})

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "space-page-1")
	assert.Contains(t, stdout, "space-page-2")
	assert.Equal(t, 2, env.ServerlessSpaceServer.ListSpacesCalls.Count())
}

func TestSpaceList_ManualPagination(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.ListSpacesCalls.Returns(&spacev1.ListSpacesResponse{
		Items:         []*spacev1.Space{{Id: "space-1"}},
		NextPageToken: new("token-2"),
	}, nil)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "list", "--page-size", "1")
	require.NoError(t, err)
	assert.Equal(t, 1, env.ServerlessSpaceServer.ListSpacesCalls.Count())

	req, ok := env.ServerlessSpaceServer.ListSpacesCalls.Last()
	require.True(t, ok)
	assert.Equal(t, int32(1), req.GetPageSize())
}

func TestSpaceList_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.ListSpacesCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "list")
	require.Error(t, err)
}

func TestSpaceList_RejectsArgs(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "list", "extra")
	require.Error(t, err)
	assert.Equal(t, 0, env.ServerlessSpaceServer.ListSpacesCalls.Count())
}
