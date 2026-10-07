package testutil

import (
	"context"

	serverlessplatformv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/platform/v1"
)

// FakeServerlessPlatformService is a test fake that implements the serverless PlatformServiceServer.
// Use the *Calls fields to configure responses and inspect captured requests.
type FakeServerlessPlatformService struct {
	serverlessplatformv1.UnimplementedPlatformServiceServer
	ListCloudRegionsCalls MethodSpy[*serverlessplatformv1.ListCloudRegionsRequest, *serverlessplatformv1.ListCloudRegionsResponse]
	GetCloudRegionCalls   MethodSpy[*serverlessplatformv1.GetCloudRegionRequest, *serverlessplatformv1.GetCloudRegionResponse]
}

// ListCloudRegions records the call and dispatches via ListCloudRegionsCalls.
func (f *FakeServerlessPlatformService) ListCloudRegions(ctx context.Context, req *serverlessplatformv1.ListCloudRegionsRequest) (*serverlessplatformv1.ListCloudRegionsResponse, error) {
	f.ListCloudRegionsCalls.record(req)
	return f.ListCloudRegionsCalls.dispatch(ctx, req, f.UnimplementedPlatformServiceServer.ListCloudRegions)
}

// GetCloudRegion records the call and dispatches via GetCloudRegionCalls.
func (f *FakeServerlessPlatformService) GetCloudRegion(ctx context.Context, req *serverlessplatformv1.GetCloudRegionRequest) (*serverlessplatformv1.GetCloudRegionResponse, error) {
	f.GetCloudRegionCalls.record(req)
	return f.GetCloudRegionCalls.dispatch(ctx, req, f.UnimplementedPlatformServiceServer.GetCloudRegion)
}
