package testutil

import (
	"context"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"
)

// FakeBookingService is a test fake that implements BookingServiceServer.
// Use the *Calls fields to configure responses and inspect captured requests.
type FakeBookingService struct {
	bookingv1.UnimplementedBookingServiceServer

	ListPackagesCalls              MethodSpy[*bookingv1.ListPackagesRequest, *bookingv1.ListPackagesResponse]
	ListGlobalPackagesCalls        MethodSpy[*bookingv1.ListGlobalPackagesRequest, *bookingv1.ListGlobalPackagesResponse]
	GetPackageCalls                MethodSpy[*bookingv1.GetPackageRequest, *bookingv1.GetPackageResponse]
	ListGlobalInferenceModelsCalls MethodSpy[*bookingv1.ListGlobalInferenceModelsRequest, *bookingv1.ListGlobalInferenceModelsResponse]
}

// ListPackages records the call and dispatches via ListPackagesCalls.
func (f *FakeBookingService) ListPackages(ctx context.Context, req *bookingv1.ListPackagesRequest) (*bookingv1.ListPackagesResponse, error) {
	f.ListPackagesCalls.record(req)
	return f.ListPackagesCalls.dispatch(ctx, req, f.UnimplementedBookingServiceServer.ListPackages)
}

// GetPackage records the call and dispatches via GetPackageCalls.
func (f *FakeBookingService) GetPackage(ctx context.Context, req *bookingv1.GetPackageRequest) (*bookingv1.GetPackageResponse, error) {
	f.GetPackageCalls.record(req)
	return f.GetPackageCalls.dispatch(ctx, req, f.UnimplementedBookingServiceServer.GetPackage)
}

// ListGlobalPackages records the call and dispatches via ListGlobalPackagesCalls.
func (f *FakeBookingService) ListGlobalPackages(ctx context.Context, req *bookingv1.ListGlobalPackagesRequest) (*bookingv1.ListGlobalPackagesResponse, error) {
	f.ListGlobalPackagesCalls.record(req)
	return f.ListGlobalPackagesCalls.dispatch(ctx, req, f.UnimplementedBookingServiceServer.ListGlobalPackages)
}

// ListGlobalInferenceModels records the call and dispatches via ListGlobalInferenceModelsCalls.
func (f *FakeBookingService) ListGlobalInferenceModels(ctx context.Context, req *bookingv1.ListGlobalInferenceModelsRequest) (*bookingv1.ListGlobalInferenceModelsResponse, error) {
	f.ListGlobalInferenceModelsCalls.record(req)
	return f.ListGlobalInferenceModelsCalls.dispatch(ctx, req, f.UnimplementedBookingServiceServer.ListGlobalInferenceModels)
}
