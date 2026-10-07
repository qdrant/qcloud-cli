package cluster

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/qdrant/qcloud-cli/internal/cmd/util"
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
	start := time.Now()
	_, err := util.PollUntilDone(ctx, timeout, pollInterval, "API key to become active",
		func(ctx context.Context) (struct{}, bool, error) {
			elapsed := time.Since(start).Round(time.Second)
			if err := probe(ctx); err != nil {
				fmt.Fprintf(out, "waiting for API key... %v (%s)\n", err, elapsed)
				return struct{}{}, false, nil
			}

			fmt.Fprintf(out, "API key is active (%s)\n", elapsed)
			return struct{}{}, true, nil
		})
	return err
}
