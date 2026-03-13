package testutil

import (
	"context"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"
)

// FakeBookingService is a test fake that implements BookingServiceServer.
// Set the function fields to control responses per test.
type FakeBookingService struct {
	bookingv1.UnimplementedBookingServiceServer

	ListPackagesFunc func(context.Context, *bookingv1.ListPackagesRequest) (*bookingv1.ListPackagesResponse, error)
}

// ListPackages delegates to ListPackagesFunc if set.
func (f *FakeBookingService) ListPackages(ctx context.Context, req *bookingv1.ListPackagesRequest) (*bookingv1.ListPackagesResponse, error) {
	if f.ListPackagesFunc != nil {
		return f.ListPackagesFunc(ctx, req)
	}
	return f.UnimplementedBookingServiceServer.ListPackages(ctx, req)
}
