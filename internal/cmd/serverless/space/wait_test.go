package space_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spacev1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestSpaceWait_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.GetSpaceCalls.
		OnCall(0, func(_ context.Context, _ *spacev1.GetSpaceRequest) (*spacev1.GetSpaceResponse, error) {
			return &spacev1.GetSpaceResponse{Space: &spacev1.Space{
				Id:    "space-abc",
				State: &spacev1.SpaceState{Phase: spacev1.SpaceStatePhase_SPACE_STATE_PHASE_PROCESSING},
			}}, nil
		}).
		Always(func(_ context.Context, _ *spacev1.GetSpaceRequest) (*spacev1.GetSpaceResponse, error) {
			return &spacev1.GetSpaceResponse{Space: &spacev1.Space{
				Id:   "space-abc",
				Name: "my-space",
				State: &spacev1.SpaceState{
					Phase:    spacev1.SpaceStatePhase_SPACE_STATE_PHASE_READY,
					Endpoint: &spacev1.SpaceEndpoint{Url: "https://space-abc.serverless.qdrant.io"},
				},
			}}, nil
		})

	stdout, stderr, err := testutil.Exec(t, env, "serverless", "space", "wait", "space-abc",
		"--timeout", "30s", "--poll-interval", "10ms",
	)
	require.NoError(t, err)
	assert.Contains(t, stderr, "phase=PROCESSING")
	assert.Contains(t, stdout, "space-abc")
	assert.Contains(t, stdout, "https://space-abc.serverless.qdrant.io")

	req, ok := env.ServerlessSpaceServer.GetSpaceCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "space-abc", req.GetSpaceId())
}

func TestSpaceWait_ReadyWithoutEndpoint(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.GetSpaceCalls.Returns(&spacev1.GetSpaceResponse{Space: &spacev1.Space{
		Id:    "space-abc",
		Name:  "my-space",
		State: &spacev1.SpaceState{Phase: spacev1.SpaceStatePhase_SPACE_STATE_PHASE_READY},
	}}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "wait", "space-abc", "--poll-interval", "10ms")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Space space-abc (my-space) is ready.")
}

func TestSpaceWait_FailurePhases(t *testing.T) {
	for _, phase := range []spacev1.SpaceStatePhase{
		spacev1.SpaceStatePhase_SPACE_STATE_PHASE_DISABLED,
		spacev1.SpaceStatePhase_SPACE_STATE_PHASE_DELETING,
	} {
		t.Run(phase.String(), func(t *testing.T) {
			env := testutil.NewTestEnv(t)

			env.ServerlessSpaceServer.GetSpaceCalls.Returns(&spacev1.GetSpaceResponse{Space: &spacev1.Space{
				Id:    "space-fail",
				State: &spacev1.SpaceState{Phase: phase, Reason: "quota exceeded"},
			}}, nil)

			_, _, err := testutil.Exec(t, env, "serverless", "space", "wait", "space-fail",
				"--timeout", "30s", "--poll-interval", "10ms",
			)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "quota exceeded")
			assert.Equal(t, 1, env.ServerlessSpaceServer.GetSpaceCalls.Count())
		})
	}
}

func TestSpaceWait_Timeout(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.GetSpaceCalls.Returns(&spacev1.GetSpaceResponse{Space: &spacev1.Space{
		Id:    "space-slow",
		State: &spacev1.SpaceState{Phase: spacev1.SpaceStatePhase_SPACE_STATE_PHASE_PROCESSING},
	}}, nil)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "wait", "space-slow",
		"--timeout", "200ms", "--poll-interval", "10ms",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
}

func TestSpaceWait_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.ServerlessSpaceServer.GetSpaceCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "wait", "space-abc", "--poll-interval", "10ms")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get space status")
}
