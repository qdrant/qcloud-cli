package testutil

import (
	"context"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
)

// FakeClusterService is a test fake that implements ClusterServiceServer.
// Set the function fields to control responses per test.
type FakeClusterService struct {
	clusterv1.UnimplementedClusterServiceServer

	ListClustersFunc       func(context.Context, *clusterv1.ListClustersRequest) (*clusterv1.ListClustersResponse, error)
	GetClusterFunc         func(context.Context, *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error)
	CreateClusterFunc      func(context.Context, *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error)
	UpdateClusterFunc      func(context.Context, *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error)
	DeleteClusterFunc      func(context.Context, *clusterv1.DeleteClusterRequest) (*clusterv1.DeleteClusterResponse, error)
	RestartClusterFunc     func(context.Context, *clusterv1.RestartClusterRequest) (*clusterv1.RestartClusterResponse, error)
	SuspendClusterFunc     func(context.Context, *clusterv1.SuspendClusterRequest) (*clusterv1.SuspendClusterResponse, error)
	UnsuspendClusterFunc   func(context.Context, *clusterv1.UnsuspendClusterRequest) (*clusterv1.UnsuspendClusterResponse, error)
	SuggestClusterNameFunc func(context.Context, *clusterv1.SuggestClusterNameRequest) (*clusterv1.SuggestClusterNameResponse, error)
	ListQdrantReleasesFunc func(context.Context, *clusterv1.ListQdrantReleasesRequest) (*clusterv1.ListQdrantReleasesResponse, error)
}

// ListClusters delegates to ListClustersFunc if set.
func (f *FakeClusterService) ListClusters(ctx context.Context, req *clusterv1.ListClustersRequest) (*clusterv1.ListClustersResponse, error) {
	if f.ListClustersFunc != nil {
		return f.ListClustersFunc(ctx, req)
	}
	return f.UnimplementedClusterServiceServer.ListClusters(ctx, req)
}

// GetCluster delegates to GetClusterFunc if set.
func (f *FakeClusterService) GetCluster(ctx context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
	if f.GetClusterFunc != nil {
		return f.GetClusterFunc(ctx, req)
	}
	return f.UnimplementedClusterServiceServer.GetCluster(ctx, req)
}

// CreateCluster delegates to CreateClusterFunc if set.
func (f *FakeClusterService) CreateCluster(ctx context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error) {
	if f.CreateClusterFunc != nil {
		return f.CreateClusterFunc(ctx, req)
	}
	return f.UnimplementedClusterServiceServer.CreateCluster(ctx, req)
}

// UpdateCluster delegates to UpdateClusterFunc if set.
func (f *FakeClusterService) UpdateCluster(ctx context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
	if f.UpdateClusterFunc != nil {
		return f.UpdateClusterFunc(ctx, req)
	}
	return f.UnimplementedClusterServiceServer.UpdateCluster(ctx, req)
}

// DeleteCluster delegates to DeleteClusterFunc if set.
func (f *FakeClusterService) DeleteCluster(ctx context.Context, req *clusterv1.DeleteClusterRequest) (*clusterv1.DeleteClusterResponse, error) {
	if f.DeleteClusterFunc != nil {
		return f.DeleteClusterFunc(ctx, req)
	}
	return f.UnimplementedClusterServiceServer.DeleteCluster(ctx, req)
}

// RestartCluster delegates to RestartClusterFunc if set.
func (f *FakeClusterService) RestartCluster(ctx context.Context, req *clusterv1.RestartClusterRequest) (*clusterv1.RestartClusterResponse, error) {
	if f.RestartClusterFunc != nil {
		return f.RestartClusterFunc(ctx, req)
	}
	return f.UnimplementedClusterServiceServer.RestartCluster(ctx, req)
}

// SuspendCluster delegates to SuspendClusterFunc if set.
func (f *FakeClusterService) SuspendCluster(ctx context.Context, req *clusterv1.SuspendClusterRequest) (*clusterv1.SuspendClusterResponse, error) {
	if f.SuspendClusterFunc != nil {
		return f.SuspendClusterFunc(ctx, req)
	}
	return f.UnimplementedClusterServiceServer.SuspendCluster(ctx, req)
}

// UnsuspendCluster delegates to UnsuspendClusterFunc if set.
func (f *FakeClusterService) UnsuspendCluster(ctx context.Context, req *clusterv1.UnsuspendClusterRequest) (*clusterv1.UnsuspendClusterResponse, error) {
	if f.UnsuspendClusterFunc != nil {
		return f.UnsuspendClusterFunc(ctx, req)
	}
	return f.UnimplementedClusterServiceServer.UnsuspendCluster(ctx, req)
}

// SuggestClusterName delegates to SuggestClusterNameFunc if set.
func (f *FakeClusterService) SuggestClusterName(ctx context.Context, req *clusterv1.SuggestClusterNameRequest) (*clusterv1.SuggestClusterNameResponse, error) {
	if f.SuggestClusterNameFunc != nil {
		return f.SuggestClusterNameFunc(ctx, req)
	}
	return f.UnimplementedClusterServiceServer.SuggestClusterName(ctx, req)
}

// ListQdrantReleases delegates to ListQdrantReleasesFunc if set.
func (f *FakeClusterService) ListQdrantReleases(ctx context.Context, req *clusterv1.ListQdrantReleasesRequest) (*clusterv1.ListQdrantReleasesResponse, error) {
	if f.ListQdrantReleasesFunc != nil {
		return f.ListQdrantReleasesFunc(ctx, req)
	}
	return f.UnimplementedClusterServiceServer.ListQdrantReleases(ctx, req)
}
