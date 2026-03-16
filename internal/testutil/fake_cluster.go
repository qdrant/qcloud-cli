package testutil

import (
	"context"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
)

// FakeClusterService is a test fake that implements ClusterServiceServer.
// Use the *Calls fields to configure responses and inspect captured requests.
type FakeClusterService struct {
	clusterv1.UnimplementedClusterServiceServer

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

// ListClusters records the call and dispatches via ListClustersCalls.
func (f *FakeClusterService) ListClusters(ctx context.Context, req *clusterv1.ListClustersRequest) (*clusterv1.ListClustersResponse, error) {
	f.ListClustersCalls.record(req)
	return f.ListClustersCalls.dispatch(ctx, req, f.UnimplementedClusterServiceServer.ListClusters)
}

// GetCluster records the call and dispatches via GetClusterCalls.
func (f *FakeClusterService) GetCluster(ctx context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
	f.GetClusterCalls.record(req)
	return f.GetClusterCalls.dispatch(ctx, req, f.UnimplementedClusterServiceServer.GetCluster)
}

// CreateCluster records the call and dispatches via CreateClusterCalls.
func (f *FakeClusterService) CreateCluster(ctx context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error) {
	f.CreateClusterCalls.record(req)
	return f.CreateClusterCalls.dispatch(ctx, req, f.UnimplementedClusterServiceServer.CreateCluster)
}

// UpdateCluster records the call and dispatches via UpdateClusterCalls.
func (f *FakeClusterService) UpdateCluster(ctx context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
	f.UpdateClusterCalls.record(req)
	return f.UpdateClusterCalls.dispatch(ctx, req, f.UnimplementedClusterServiceServer.UpdateCluster)
}

// DeleteCluster records the call and dispatches via DeleteClusterCalls.
func (f *FakeClusterService) DeleteCluster(ctx context.Context, req *clusterv1.DeleteClusterRequest) (*clusterv1.DeleteClusterResponse, error) {
	f.DeleteClusterCalls.record(req)
	return f.DeleteClusterCalls.dispatch(ctx, req, f.UnimplementedClusterServiceServer.DeleteCluster)
}

// RestartCluster records the call and dispatches via RestartClusterCalls.
func (f *FakeClusterService) RestartCluster(ctx context.Context, req *clusterv1.RestartClusterRequest) (*clusterv1.RestartClusterResponse, error) {
	f.RestartClusterCalls.record(req)
	return f.RestartClusterCalls.dispatch(ctx, req, f.UnimplementedClusterServiceServer.RestartCluster)
}

// SuspendCluster records the call and dispatches via SuspendClusterCalls.
func (f *FakeClusterService) SuspendCluster(ctx context.Context, req *clusterv1.SuspendClusterRequest) (*clusterv1.SuspendClusterResponse, error) {
	f.SuspendClusterCalls.record(req)
	return f.SuspendClusterCalls.dispatch(ctx, req, f.UnimplementedClusterServiceServer.SuspendCluster)
}

// UnsuspendCluster records the call and dispatches via UnsuspendClusterCalls.
func (f *FakeClusterService) UnsuspendCluster(ctx context.Context, req *clusterv1.UnsuspendClusterRequest) (*clusterv1.UnsuspendClusterResponse, error) {
	f.UnsuspendClusterCalls.record(req)
	return f.UnsuspendClusterCalls.dispatch(ctx, req, f.UnimplementedClusterServiceServer.UnsuspendCluster)
}

// SuggestClusterName records the call and dispatches via SuggestClusterNameCalls.
func (f *FakeClusterService) SuggestClusterName(ctx context.Context, req *clusterv1.SuggestClusterNameRequest) (*clusterv1.SuggestClusterNameResponse, error) {
	f.SuggestClusterNameCalls.record(req)
	return f.SuggestClusterNameCalls.dispatch(ctx, req, f.UnimplementedClusterServiceServer.SuggestClusterName)
}

// ListQdrantReleases records the call and dispatches via ListQdrantReleasesCalls.
func (f *FakeClusterService) ListQdrantReleases(ctx context.Context, req *clusterv1.ListQdrantReleasesRequest) (*clusterv1.ListQdrantReleasesResponse, error) {
	f.ListQdrantReleasesCalls.record(req)
	return f.ListQdrantReleasesCalls.dispatch(ctx, req, f.UnimplementedClusterServiceServer.ListQdrantReleases)
}
