package clusterutil

import (
	"context"
	"fmt"
	"io"
	"time"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
)

var clusterFailurePhases = map[clusterv1.ClusterPhase]bool{
	clusterv1.ClusterPhase_CLUSTER_PHASE_FAILED_TO_CREATE: true,
	clusterv1.ClusterPhase_CLUSTER_PHASE_FAILED_TO_SYNC:   true,
	clusterv1.ClusterPhase_CLUSTER_PHASE_NOT_FOUND:        true,
}

// WaitForClusterHealthy polls cluster status until it becomes healthy, times out,
// or enters a failure phase. Progress lines are written to out.
func WaitForClusterHealthy(
	ctx context.Context,
	svc clusterv1.ClusterServiceClient,
	out io.Writer,
	accountID, clusterID string,
	timeout, pollInterval time.Duration,
) (*clusterv1.Cluster, error) {
	start := time.Now()
	return util.PollUntilDone(ctx, timeout, pollInterval, "cluster to become healthy",
		func(ctx context.Context) (*clusterv1.Cluster, bool, error) {
			resp, err := svc.GetCluster(ctx, &clusterv1.GetClusterRequest{
				AccountId: accountID,
				ClusterId: clusterID,
			})
			if err != nil {
				return nil, false, fmt.Errorf("failed to get cluster status: %w", err)
			}

			cluster := resp.GetCluster()
			phase := cluster.GetState().GetPhase()
			phaseStr := output.ClusterPhase(phase)
			fmt.Fprintf(out, "phase=%s (%s)\n", phaseStr, time.Since(start).Round(time.Second))

			if phase == clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY {
				return cluster, true, nil
			}

			if clusterFailurePhases[phase] {
				return nil, false, fmt.Errorf("failed waiting for cluster to become healthy: phase=%s, reason=%s",
					phaseStr, cluster.GetState().GetReason())
			}

			return nil, false, nil
		})
}
