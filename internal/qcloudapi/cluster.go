package qcloudapi

import (
	"context"
	"fmt"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
)

// HybridCloudProviderID is the cloud provider identifier for hybrid cloud environments.
const HybridCloudProviderID = "hybrid"

// ClusterClient wraps the generated ClusterServiceClient with convenience methods.
type ClusterClient struct {
	clusterv1.ClusterServiceClient
}

// ListAllClusters returns all clusters (cloud and hybrid), auto-paginating.
func (c *ClusterClient) ListAllClusters(ctx context.Context, accountID string) ([]*clusterv1.Cluster, error) {
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

		all = append(all, resp.Items...)
		if resp.NextPageToken == nil || *resp.NextPageToken == "" {
			break
		}

		nextToken = resp.NextPageToken
	}

	return all, nil
}
