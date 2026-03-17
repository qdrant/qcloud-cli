package testutil

import (
	"context"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"
)

// FakeBookingService is a test fake that implements BookingServiceServer.
// Use the *Calls fields to configure responses and inspect captured requests.
type FakeBookingService struct {
	bookingv1.UnimplementedBookingServiceServer

	ListPackagesCalls MethodSpy[*bookingv1.ListPackagesRequest, *bookingv1.ListPackagesResponse]
	GetPackageCalls   MethodSpy[*bookingv1.GetPackageRequest, *bookingv1.GetPackageResponse]
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
