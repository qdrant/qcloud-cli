package testutil

import (
	"context"

	platformv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/platform/v1"
)

// FakePlatformService is a test fake that implements PlatformServiceServer.
// Use the *Calls fields to configure responses and inspect captured requests.
type FakePlatformService struct {
	platformv1.UnimplementedPlatformServiceServer

	ListCloudProvidersCalls       MethodSpy[*platformv1.ListCloudProvidersRequest, *platformv1.ListCloudProvidersResponse]
	ListCloudProviderRegionsCalls MethodSpy[*platformv1.ListCloudProviderRegionsRequest, *platformv1.ListCloudProviderRegionsResponse]
}

// ListCloudProviders records the call and dispatches via ListCloudProvidersCalls.
func (f *FakePlatformService) ListCloudProviders(ctx context.Context, req *platformv1.ListCloudProvidersRequest) (*platformv1.ListCloudProvidersResponse, error) {
	f.ListCloudProvidersCalls.record(req)
	return f.ListCloudProvidersCalls.dispatch(ctx, req, f.UnimplementedPlatformServiceServer.ListCloudProviders)
}

// ListCloudProviderRegions records the call and dispatches via ListCloudProviderRegionsCalls.
func (f *FakePlatformService) ListCloudProviderRegions(ctx context.Context, req *platformv1.ListCloudProviderRegionsRequest) (*platformv1.ListCloudProviderRegionsResponse, error) {
	f.ListCloudProviderRegionsCalls.record(req)
	return f.ListCloudProviderRegionsCalls.dispatch(ctx, req, f.UnimplementedPlatformServiceServer.ListCloudProviderRegions)
}
