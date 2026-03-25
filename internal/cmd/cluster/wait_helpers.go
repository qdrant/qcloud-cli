package cluster

import (
	"context"
	"io"
	"time"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/qcloudapi"
)

func waitForHealthyWithInterval(
	ctx context.Context,
	svc clusterv1.ClusterServiceClient,
	out io.Writer,
	accountID, clusterID string,
	timeout, pollInterval time.Duration,
) (*clusterv1.Cluster, error) {
	return qcloudapi.WaitForClusterHealthy(ctx, svc, out, accountID, clusterID, timeout, pollInterval)
}
