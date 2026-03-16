package testutil

import (
	"context"

	platformv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/platform/v1"
)

// FakePlatformService is a test fake that implements PlatformServiceServer.
// Set the function fields to control responses per test.
type FakePlatformService struct {
	platformv1.UnimplementedPlatformServiceServer

	ListCloudProvidersFunc       func(context.Context, *platformv1.ListCloudProvidersRequest) (*platformv1.ListCloudProvidersResponse, error)
	ListCloudProviderRegionsFunc func(context.Context, *platformv1.ListCloudProviderRegionsRequest) (*platformv1.ListCloudProviderRegionsResponse, error)

	ListCloudProvidersCalls       MethodSpy[*platformv1.ListCloudProvidersRequest, *platformv1.ListCloudProvidersResponse]
	ListCloudProviderRegionsCalls MethodSpy[*platformv1.ListCloudProviderRegionsRequest, *platformv1.ListCloudProviderRegionsResponse]
}

// ListCloudProviders delegates to ListCloudProvidersFunc if set, otherwise dispatches via ListCloudProvidersCalls.
func (f *FakePlatformService) ListCloudProviders(ctx context.Context, req *platformv1.ListCloudProvidersRequest) (*platformv1.ListCloudProvidersResponse, error) {
	f.ListCloudProvidersCalls.record(req)
	if f.ListCloudProvidersFunc != nil {
		return f.ListCloudProvidersFunc(ctx, req)
	}
	return f.ListCloudProvidersCalls.dispatch(ctx, req, f.UnimplementedPlatformServiceServer.ListCloudProviders)
}

// ListCloudProviderRegions delegates to ListCloudProviderRegionsFunc if set, otherwise dispatches via ListCloudProviderRegionsCalls.
func (f *FakePlatformService) ListCloudProviderRegions(ctx context.Context, req *platformv1.ListCloudProviderRegionsRequest) (*platformv1.ListCloudProviderRegionsResponse, error) {
	f.ListCloudProviderRegionsCalls.record(req)
	if f.ListCloudProviderRegionsFunc != nil {
		return f.ListCloudProviderRegionsFunc(ctx, req)
	}
	return f.ListCloudProviderRegionsCalls.dispatch(ctx, req, f.UnimplementedPlatformServiceServer.ListCloudProviderRegions)
}
