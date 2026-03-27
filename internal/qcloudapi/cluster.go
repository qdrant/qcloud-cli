package qcloudapi

import (
	"context"
	"fmt"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
)

const hybridCloudProviderID = "hybrid"

// ClusterClient wraps the generated ClusterServiceClient with convenience methods.
type ClusterClient struct {
	clusterv1.ClusterServiceClient
}

// ListCloudClusters returns all non-hybrid clusters, auto-paginating.
func (c *ClusterClient) ListCloudClusters(ctx context.Context, accountID string) ([]*clusterv1.Cluster, error) {
	return c.listFiltered(ctx, accountID, false)
}

// ListHybridClusters returns all hybrid clusters, auto-paginating.
func (c *ClusterClient) ListHybridClusters(ctx context.Context, accountID string) ([]*clusterv1.Cluster, error) {
	return c.listFiltered(ctx, accountID, true)
}

func (c *ClusterClient) listFiltered(ctx context.Context, accountID string, hybrid bool) ([]*clusterv1.Cluster, error) {
	req := &clusterv1.ListClustersRequest{AccountId: accountID}
	var all []*clusterv1.Cluster
	var nextToken *string
	for {
		if nextToken != nil {
			req.PageToken = nextToken
		}
		resp, err := c.ListClusters(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to list clusters: %w", err)
		}
		for _, cluster := range resp.Items {
			isHybrid := cluster.GetCloudProviderId() == hybridCloudProviderID
			if isHybrid == hybrid {
				all = append(all, cluster)
			}
		}
		if resp.NextPageToken == nil || *resp.NextPageToken == "" {
			break
		}
		nextToken = resp.NextPageToken
	}
	return all, nil
}
