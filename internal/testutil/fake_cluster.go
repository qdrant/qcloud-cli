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

	ListClustersCalls       MethodSpy[*clusterv1.ListClustersRequest, *clusterv1.ListClustersResponse]
	GetClusterCalls         MethodSpy[*clusterv1.GetClusterRequest, *clusterv1.GetClusterResponse]
	CreateClusterCalls      MethodSpy[*clusterv1.CreateClusterRequest, *clusterv1.CreateClusterResponse]
	UpdateClusterCalls      MethodSpy[*clusterv1.UpdateClusterRequest, *clusterv1.UpdateClusterResponse]
	DeleteClusterCalls      MethodSpy[*clusterv1.DeleteClusterRequest, *clusterv1.DeleteClusterResponse]
	RestartClusterCalls     MethodSpy[*clusterv1.RestartClusterRequest, *clusterv1.RestartClusterResponse]
	SuspendClusterCalls     MethodSpy[*clusterv1.SuspendClusterRequest, *clusterv1.SuspendClusterResponse]
	UnsuspendClusterCalls   MethodSpy[*clusterv1.UnsuspendClusterRequest, *clusterv1.UnsuspendClusterResponse]
	SuggestClusterNameCalls MethodSpy[*clusterv1.SuggestClusterNameRequest, *clusterv1.SuggestClusterNameResponse]
	ListQdrantReleasesCalls MethodSpy[*clusterv1.ListQdrantReleasesRequest, *clusterv1.ListQdrantReleasesResponse]
}

// ListClusters delegates to ListClustersFunc if set, otherwise dispatches via ListClustersCalls.
func (f *FakeClusterService) ListClusters(ctx context.Context, req *clusterv1.ListClustersRequest) (*clusterv1.ListClustersResponse, error) {
	f.ListClustersCalls.record(req)
	if f.ListClustersFunc != nil {
		return f.ListClustersFunc(ctx, req)
	}
	return f.ListClustersCalls.dispatch(ctx, req, f.UnimplementedClusterServiceServer.ListClusters)
}

// GetCluster delegates to GetClusterFunc if set, otherwise dispatches via GetClusterCalls.
func (f *FakeClusterService) GetCluster(ctx context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
	f.GetClusterCalls.record(req)
	if f.GetClusterFunc != nil {
		return f.GetClusterFunc(ctx, req)
	}
	return f.GetClusterCalls.dispatch(ctx, req, f.UnimplementedClusterServiceServer.GetCluster)
}

// CreateCluster delegates to CreateClusterFunc if set, otherwise dispatches via CreateClusterCalls.
func (f *FakeClusterService) CreateCluster(ctx context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error) {
	f.CreateClusterCalls.record(req)
	if f.CreateClusterFunc != nil {
		return f.CreateClusterFunc(ctx, req)
	}
	return f.CreateClusterCalls.dispatch(ctx, req, f.UnimplementedClusterServiceServer.CreateCluster)
}

// UpdateCluster delegates to UpdateClusterFunc if set, otherwise dispatches via UpdateClusterCalls.
func (f *FakeClusterService) UpdateCluster(ctx context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
	f.UpdateClusterCalls.record(req)
	if f.UpdateClusterFunc != nil {
		return f.UpdateClusterFunc(ctx, req)
	}
	return f.UpdateClusterCalls.dispatch(ctx, req, f.UnimplementedClusterServiceServer.UpdateCluster)
}

// DeleteCluster delegates to DeleteClusterFunc if set, otherwise dispatches via DeleteClusterCalls.
func (f *FakeClusterService) DeleteCluster(ctx context.Context, req *clusterv1.DeleteClusterRequest) (*clusterv1.DeleteClusterResponse, error) {
	f.DeleteClusterCalls.record(req)
	if f.DeleteClusterFunc != nil {
		return f.DeleteClusterFunc(ctx, req)
	}
	return f.DeleteClusterCalls.dispatch(ctx, req, f.UnimplementedClusterServiceServer.DeleteCluster)
}

// RestartCluster delegates to RestartClusterFunc if set, otherwise dispatches via RestartClusterCalls.
func (f *FakeClusterService) RestartCluster(ctx context.Context, req *clusterv1.RestartClusterRequest) (*clusterv1.RestartClusterResponse, error) {
	f.RestartClusterCalls.record(req)
	if f.RestartClusterFunc != nil {
		return f.RestartClusterFunc(ctx, req)
	}
	return f.RestartClusterCalls.dispatch(ctx, req, f.UnimplementedClusterServiceServer.RestartCluster)
}

// SuspendCluster delegates to SuspendClusterFunc if set, otherwise dispatches via SuspendClusterCalls.
func (f *FakeClusterService) SuspendCluster(ctx context.Context, req *clusterv1.SuspendClusterRequest) (*clusterv1.SuspendClusterResponse, error) {
	f.SuspendClusterCalls.record(req)
	if f.SuspendClusterFunc != nil {
		return f.SuspendClusterFunc(ctx, req)
	}
	return f.SuspendClusterCalls.dispatch(ctx, req, f.UnimplementedClusterServiceServer.SuspendCluster)
}

// UnsuspendCluster delegates to UnsuspendClusterFunc if set, otherwise dispatches via UnsuspendClusterCalls.
func (f *FakeClusterService) UnsuspendCluster(ctx context.Context, req *clusterv1.UnsuspendClusterRequest) (*clusterv1.UnsuspendClusterResponse, error) {
	f.UnsuspendClusterCalls.record(req)
	if f.UnsuspendClusterFunc != nil {
		return f.UnsuspendClusterFunc(ctx, req)
	}
	return f.UnsuspendClusterCalls.dispatch(ctx, req, f.UnimplementedClusterServiceServer.UnsuspendCluster)
}

// SuggestClusterName delegates to SuggestClusterNameFunc if set, otherwise dispatches via SuggestClusterNameCalls.
func (f *FakeClusterService) SuggestClusterName(ctx context.Context, req *clusterv1.SuggestClusterNameRequest) (*clusterv1.SuggestClusterNameResponse, error) {
	f.SuggestClusterNameCalls.record(req)
	if f.SuggestClusterNameFunc != nil {
		return f.SuggestClusterNameFunc(ctx, req)
	}
	return f.SuggestClusterNameCalls.dispatch(ctx, req, f.UnimplementedClusterServiceServer.SuggestClusterName)
}

// ListQdrantReleases delegates to ListQdrantReleasesFunc if set, otherwise dispatches via ListQdrantReleasesCalls.
func (f *FakeClusterService) ListQdrantReleases(ctx context.Context, req *clusterv1.ListQdrantReleasesRequest) (*clusterv1.ListQdrantReleasesResponse, error) {
	f.ListQdrantReleasesCalls.record(req)
	if f.ListQdrantReleasesFunc != nil {
		return f.ListQdrantReleasesFunc(ctx, req)
	}
	return f.ListQdrantReleasesCalls.dispatch(ctx, req, f.UnimplementedClusterServiceServer.ListQdrantReleases)
}
