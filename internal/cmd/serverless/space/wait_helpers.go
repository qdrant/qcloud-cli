package space

import (
	"context"
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	spaceauthv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/auth/v1"
	spacev1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
)

var spaceFailurePhases = map[spacev1.SpaceStatePhase]bool{
	spacev1.SpaceStatePhase_SPACE_STATE_PHASE_DISABLED: true,
	spacev1.SpaceStatePhase_SPACE_STATE_PHASE_DELETING: true,
}

// pollUntilDone calls poll immediately and then on every tick until it returns
// done=true, returns an error, or the timeout expires. what describes the awaited
// condition in timeout errors (e.g. "space to become ready").
func pollUntilDone[T any](
	ctx context.Context,
	timeout, pollInterval time.Duration,
	what string,
	poll func(ctx context.Context) (T, bool, error),
) (T, error) {
	var zero T

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for first := true; ; first = false {
		if !first {
			select {
			case <-ctx.Done():
				return zero, fmt.Errorf("timed out waiting for %s: %w", what, ctx.Err())
			case <-ticker.C:
			}
		}

		v, done, err := poll(ctx)
		if err != nil {
			if s, ok := status.FromError(err); ok && s.Code() == codes.DeadlineExceeded {
				return zero, fmt.Errorf("timed out waiting for %s: %w", what, err)
			}

			return zero, err
		}

		if done {
			return v, nil
		}
	}
}

// waitForSpaceReady polls the space until it becomes ready, times out, or enters
// a failure phase. Progress lines are written to out.
func waitForSpaceReady(
	ctx context.Context,
	svc spacev1.SpaceServiceClient,
	out io.Writer,
	accountID, spaceID string,
	timeout, pollInterval time.Duration,
) (*spacev1.Space, error) {
	start := time.Now()
	return pollUntilDone(ctx, timeout, pollInterval, "space to become ready",
		func(ctx context.Context) (*spacev1.Space, bool, error) {
			resp, err := svc.GetSpace(ctx, &spacev1.GetSpaceRequest{
				AccountId: accountID,
				SpaceId:   spaceID,
			})
			if err != nil {
				return nil, false, fmt.Errorf("failed to get space status: %w", err)
			}

			space := resp.GetSpace()
			phase := space.GetState().GetPhase()
			phaseStr := output.SpacePhase(phase)
			fmt.Fprintf(out, "phase=%s (%s)\n", phaseStr, time.Since(start).Round(time.Second))

			if phase == spacev1.SpaceStatePhase_SPACE_STATE_PHASE_READY {
				return space, true, nil
			}

			if spaceFailurePhases[phase] {
				return nil, false, fmt.Errorf("failed waiting for space to become ready: phase=%s, reason=%s",
					phaseStr, space.GetState().GetReason())
			}

			return nil, false, nil
		})
}

var spaceApiKeyFailurePhases = map[spaceauthv1.SpaceApiKeyStatePhase]bool{
	spaceauthv1.SpaceApiKeyStatePhase_SPACE_API_KEY_STATE_PHASE_DISABLED: true,
	spaceauthv1.SpaceApiKeyStatePhase_SPACE_API_KEY_STATE_PHASE_DELETING: true,
}

// waitForSpaceApiKeyReady polls the API keys of the space until the key with the
// given ID becomes ready, times out, or enters a failure phase. Progress lines are
// written to out.
func waitForSpaceApiKeyReady(
	ctx context.Context,
	svc spaceauthv1.SpaceApiKeyServiceClient,
	out io.Writer,
	accountID, spaceID, keyID string,
	timeout, pollInterval time.Duration,
) (*spaceauthv1.SpaceApiKey, error) {
	start := time.Now()
	return pollUntilDone(ctx, timeout, pollInterval, "API key to become ready",
		func(ctx context.Context) (*spaceauthv1.SpaceApiKey, bool, error) {
			resp, err := svc.ListSpaceApiKeys(ctx, &spaceauthv1.ListSpaceApiKeysRequest{
				AccountId: accountID,
				SpaceId:   spaceID,
			})
			if err != nil {
				return nil, false, fmt.Errorf("failed to get API key status: %w", err)
			}

			var key *spaceauthv1.SpaceApiKey
			for _, k := range resp.GetItems() {
				if k.GetId() == keyID {
					key = k
					break
				}
			}

			if key == nil {
				fmt.Fprintf(out, "phase=NOT_FOUND (%s)\n", time.Since(start).Round(time.Second))
				return nil, false, nil
			}

			phase := key.GetState().GetPhase()
			phaseStr := output.SpaceApiKeyPhase(phase)
			fmt.Fprintf(out, "phase=%s (%s)\n", phaseStr, time.Since(start).Round(time.Second))

			if phase == spaceauthv1.SpaceApiKeyStatePhase_SPACE_API_KEY_STATE_PHASE_READY {
				return key, true, nil
			}

			if spaceApiKeyFailurePhases[phase] {
				return nil, false, fmt.Errorf("failed waiting for API key to become ready: phase=%s, reason=%s",
					phaseStr, key.GetState().GetReason())
			}

			return nil, false, nil
		})
}
