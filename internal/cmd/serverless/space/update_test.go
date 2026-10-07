package space_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"

	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"
	spacev1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/v1"

	"github.com/qdrant/qcloud-cli/internal/resource"
	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func existingSpace() *spacev1.Space {
	return &spacev1.Space{
		Id:                  "space-abc",
		AccountId:           "test-account-id",
		Name:                "my-space",
		CloudRegionId:       "eu-central-1",
		CostAllocationLabel: new("cc-1"),
		Labels: []*commonv1.KeyValue{
			{Key: "env", Value: "staging"},
			{Key: "team", Value: "search"},
		},
		Configuration: &spacev1.SpaceConfiguration{
			AllowedIpSourceRanges: []string{"10.0.0.0/8", "192.168.0.0/16"},
			AllowedOrigins:        []string{"https://old.example.com"},
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
	}
}

func setupUpdate(env *testutil.TestEnv) {
	env.ServerlessSpaceServer.GetSpaceCalls.Returns(&spacev1.GetSpaceResponse{Space: existingSpace()}, nil)
	env.ServerlessSpaceServer.UpdateSpaceCalls.Always(func(_ context.Context, req *spacev1.UpdateSpaceRequest) (*spacev1.UpdateSpaceResponse, error) {
		return &spacev1.UpdateSpaceResponse{Space: req.GetSpace()}, nil
	})
}

func TestSpaceUpdate_OnlyChangedFields(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdate(env)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "update", "space-abc",
		"--searcher-max-workers", "4",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "space-abc")
	assert.Contains(t, stdout, "updated")

	getReq, ok := env.ServerlessSpaceServer.GetSpaceCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "space-abc", getReq.GetSpaceId())

	req, ok := env.ServerlessSpaceServer.UpdateSpaceCalls.Last()
	require.True(t, ok)
	assert.Equal(t, []string{"configuration.searcher_settings.max_workers"}, req.GetUpdateMask().GetPaths())

	sp := req.GetSpace()
	want := existingSpace()
	assert.Equal(t, want.GetName(), sp.GetName())
	assert.Equal(t, want.GetCloudRegionId(), sp.GetCloudRegionId())
	assert.Equal(t, want.GetCostAllocationLabel(), sp.GetCostAllocationLabel())
	assert.Len(t, sp.GetLabels(), 2)
	assert.Equal(t, want.GetConfiguration().GetAllowedIpSourceRanges(), sp.GetConfiguration().GetAllowedIpSourceRanges())
	assert.Equal(t, want.GetConfiguration().GetAllowedOrigins(), sp.GetConfiguration().GetAllowedOrigins())
	assert.Equal(t, uint64(10*resource.GiB), sp.GetConfiguration().GetCollectionSettings().GetMaxSize())
	assert.Equal(t, 5*time.Minute, sp.GetConfiguration().GetSearcherSettings().GetIdleTimeout().AsDuration())
	assert.Equal(t, uint64(4), sp.GetConfiguration().GetSearcherSettings().GetMaxWorkers())
}

func TestSpaceUpdate_MergeAndRemove(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdate(env)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "update", "space-abc",
		"--name", "renamed-space",
		"--label", "env=prod",
		"--label", "team-",
		"--allowed-ip", "10.0.0.0/8-",
		"--allowed-ip", "172.16.0.0/12",
		"--allowed-origin", "https://old.example.com-",
		"--allowed-origin", "https://new.example.com",
		"--max-collection-size", "20GiB",
		"--searcher-idle-timeout", "15m",
	)
	require.NoError(t, err)

	req, ok := env.ServerlessSpaceServer.UpdateSpaceCalls.Last()
	require.True(t, ok)
	assert.Equal(t, []string{
		"name",
		"labels",
		"configuration.allowed_ip_source_ranges",
		"configuration.allowed_origins",
		"configuration.collection_settings.max_size",
		"configuration.searcher_settings.idle_timeout",
	}, req.GetUpdateMask().GetPaths())

	sp := req.GetSpace()
	assert.Equal(t, "renamed-space", sp.GetName())
	require.Len(t, sp.GetLabels(), 1)
	assert.Equal(t, "env", sp.GetLabels()[0].GetKey())
	assert.Equal(t, "prod", sp.GetLabels()[0].GetValue())

	cfg := sp.GetConfiguration()
	assert.Equal(t, []string{"172.16.0.0/12", "192.168.0.0/16"}, cfg.GetAllowedIpSourceRanges())
	assert.Equal(t, []string{"https://new.example.com"}, cfg.GetAllowedOrigins())
	assert.Equal(t, uint64(20*resource.GiB), cfg.GetCollectionSettings().GetMaxSize())
	assert.Equal(t, uint64(100*resource.GiB), cfg.GetCollectionSettings().GetPlatformMaxSize())
	assert.Equal(t, 15*time.Minute, cfg.GetSearcherSettings().GetIdleTimeout().AsDuration())
	assert.Equal(t, uint64(2), cfg.GetSearcherSettings().GetMaxWorkers())
}

func TestSpaceUpdate_FetchError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.GetSpaceCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "update", "space-abc", "--name", "x-space")
	require.Error(t, err)
	assert.Equal(t, 0, env.ServerlessSpaceServer.UpdateSpaceCalls.Count())
}

func TestSpaceUpdate_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.GetSpaceCalls.Returns(&spacev1.GetSpaceResponse{Space: existingSpace()}, nil)
	env.ServerlessSpaceServer.UpdateSpaceCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "update", "space-abc", "--name", "x-space")
	require.Error(t, err)
}

func TestSpaceUpdate_InvalidLabel(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdate(env)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "update", "space-abc", "--label", "=bad")
	require.Error(t, err)
	assert.Equal(t, 0, env.ServerlessSpaceServer.UpdateSpaceCalls.Count())
}

func TestSpaceUpdate_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "update")
	require.Error(t, err)
	assert.Equal(t, 0, env.ServerlessSpaceServer.GetSpaceCalls.Count())
}
