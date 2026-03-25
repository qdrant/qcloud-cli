package cluster

import (
	"context"
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
)

var failurePhases = map[clusterv1.ClusterPhase]bool{
	clusterv1.ClusterPhase_CLUSTER_PHASE_FAILED_TO_CREATE: true,
	clusterv1.ClusterPhase_CLUSTER_PHASE_FAILED_TO_SYNC:   true,
	clusterv1.ClusterPhase_CLUSTER_PHASE_NOT_FOUND:        true,
}

func waitForHealthyWithInterval(
	ctx context.Context,
	svc clusterv1.ClusterServiceClient,
	out io.Writer,
	accountID, clusterID string,
	timeout, pollInterval time.Duration,
) (*clusterv1.Cluster, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	start := time.Now()

	poll := func() (*clusterv1.Cluster, error) {
		resp, err := svc.GetCluster(ctx, &clusterv1.GetClusterRequest{
			AccountId: accountID,
			ClusterId: clusterID,
		})
		if err != nil {
			if s, ok := status.FromError(err); ok && s.Code() == codes.DeadlineExceeded {
				return nil, fmt.Errorf("timed out waiting for cluster to become healthy: %w", err)
			}
			return nil, fmt.Errorf("failed to get cluster status: %w", err)
		}

		cluster := resp.GetCluster()
		phase := cluster.GetState().GetPhase()
		elapsed := time.Since(start).Round(time.Second)
		fmt.Fprintf(out, "phase=%s (%s)\n", phaseString(phase), elapsed)

		if phase == clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY {
			return cluster, nil
		}
		if failurePhases[phase] {
			reason := cluster.GetState().GetReason()
			return nil, fmt.Errorf("failed waiting for cluster to become healthy: phase=%s, reason=%s", phaseString(phase), reason)
		}
		return nil, nil //nolint:nilnil // nil cluster means keep polling
	}

	// Poll immediately, then on each tick.
	for first := true; ; first = false {
		if !first {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("timed out waiting for cluster to become healthy: %w", ctx.Err())
			case <-ticker.C:
			}
		}

		cluster, err := poll()
		if err != nil {
			return nil, err
		}

		if cluster != nil {
			return cluster, nil
		}
	}
}
