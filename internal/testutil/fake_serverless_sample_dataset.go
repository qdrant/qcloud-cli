package testutil

import (
	"context"

	spacedatasetv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/dataset/v1"
)

// FakeServerlessSampleDatasetService is a test fake that implements the serverless SampleDatasetServiceServer.
// Use the *Calls fields to configure responses and inspect captured requests.
type FakeServerlessSampleDatasetService struct {
	spacedatasetv1.UnimplementedSampleDatasetServiceServer

	ListSampleDatasetsCalls     MethodSpy[*spacedatasetv1.ListSampleDatasetsRequest, *spacedatasetv1.ListSampleDatasetsResponse]
	ListSampleDatasetLoadsCalls MethodSpy[*spacedatasetv1.ListSampleDatasetLoadsRequest, *spacedatasetv1.ListSampleDatasetLoadsResponse]
	GetSampleDatasetLoadCalls   MethodSpy[*spacedatasetv1.GetSampleDatasetLoadRequest, *spacedatasetv1.GetSampleDatasetLoadResponse]
	LoadSampleDatasetCalls      MethodSpy[*spacedatasetv1.LoadSampleDatasetRequest, *spacedatasetv1.LoadSampleDatasetResponse]
}

// ListSampleDatasets records the call and dispatches via ListSampleDatasetsCalls.
func (f *FakeServerlessSampleDatasetService) ListSampleDatasets(ctx context.Context, req *spacedatasetv1.ListSampleDatasetsRequest) (*spacedatasetv1.ListSampleDatasetsResponse, error) {
	f.ListSampleDatasetsCalls.record(req)
	return f.ListSampleDatasetsCalls.dispatch(ctx, req, f.UnimplementedSampleDatasetServiceServer.ListSampleDatasets)
}

// ListSampleDatasetLoads records the call and dispatches via ListSampleDatasetLoadsCalls.
func (f *FakeServerlessSampleDatasetService) ListSampleDatasetLoads(ctx context.Context, req *spacedatasetv1.ListSampleDatasetLoadsRequest) (*spacedatasetv1.ListSampleDatasetLoadsResponse, error) {
	f.ListSampleDatasetLoadsCalls.record(req)
	return f.ListSampleDatasetLoadsCalls.dispatch(ctx, req, f.UnimplementedSampleDatasetServiceServer.ListSampleDatasetLoads)
}

// GetSampleDatasetLoad records the call and dispatches via GetSampleDatasetLoadCalls.
func (f *FakeServerlessSampleDatasetService) GetSampleDatasetLoad(ctx context.Context, req *spacedatasetv1.GetSampleDatasetLoadRequest) (*spacedatasetv1.GetSampleDatasetLoadResponse, error) {
	f.GetSampleDatasetLoadCalls.record(req)
	return f.GetSampleDatasetLoadCalls.dispatch(ctx, req, f.UnimplementedSampleDatasetServiceServer.GetSampleDatasetLoad)
}

// LoadSampleDataset records the call and dispatches via LoadSampleDatasetCalls.
func (f *FakeServerlessSampleDatasetService) LoadSampleDataset(ctx context.Context, req *spacedatasetv1.LoadSampleDatasetRequest) (*spacedatasetv1.LoadSampleDatasetResponse, error) {
	f.LoadSampleDatasetCalls.record(req)
	return f.LoadSampleDatasetCalls.dispatch(ctx, req, f.UnimplementedSampleDatasetServiceServer.LoadSampleDataset)
}
