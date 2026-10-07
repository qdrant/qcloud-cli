package util

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AddWaitFlags registers the --wait, --wait-timeout and hidden --wait-poll-interval
// flags. what describes the awaited condition (e.g. "the cluster to become healthy").
func AddWaitFlags(cmd *cobra.Command, what string, timeout, pollInterval time.Duration) {
	cmd.Flags().Bool("wait", false, "Wait for "+what)
	cmd.Flags().Duration("wait-timeout", timeout, "Maximum time to wait for "+what)
	cmd.Flags().Duration("wait-poll-interval", pollInterval, "How often to poll while waiting for "+what)
	_ = cmd.Flags().MarkHidden("wait-poll-interval")
}

// PollUntilDone calls poll immediately and then on every tick until it returns
// done=true, returns an error, or the timeout expires. what describes the awaited
// condition in timeout errors (e.g. "space to become ready").
func PollUntilDone[T any](
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
