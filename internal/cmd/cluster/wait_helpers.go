package cluster

import (
	"context"
	"fmt"
	"io"
	"time"
)

// waitForKeyReady polls probe until it returns nil (key accepted) or the
// timeout expires. Any non-nil error from probe is treated as "not yet ready"
// and the loop keeps going — only a context deadline causes a hard failure.
func waitForKeyReady(
	ctx context.Context,
	out io.Writer,
	probe func(ctx context.Context) error,
	timeout, pollInterval time.Duration,
) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	start := time.Now()

	for first := true; ; first = false {
		if !first {
			select {
			case <-ctx.Done():
				return fmt.Errorf("timed out waiting for API key to become active: %w", ctx.Err())
			case <-ticker.C:
			}
		}

		elapsed := time.Since(start).Round(time.Second)
		err := probe(ctx)
		if err != nil {
			fmt.Fprintf(out, "waiting for API key... %v (%s)\n", err, elapsed)
			continue
		}
		fmt.Fprintf(out, "API key is active (%s)\n", elapsed)
		return nil
	}
}
