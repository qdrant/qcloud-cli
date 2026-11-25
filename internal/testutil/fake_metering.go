package testutil

import (
	"context"

	meteringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/metering/v1"
)

// FakeMeteringService is a test fake that implements MeteringServiceServer.
// Use the *Calls fields to configure responses and inspect captured requests.
type FakeMeteringService struct {
	meteringv1.UnimplementedMeteringServiceServer
	GetUsageBreakdownCalls MethodSpy[*meteringv1.GetUsageBreakdownRequest, *meteringv1.GetUsageBreakdownResponse]
}

// GetUsageBreakdown records the call and dispatches via GetUsageBreakdownCalls.
func (f *FakeMeteringService) GetUsageBreakdown(ctx context.Context, req *meteringv1.GetUsageBreakdownRequest) (*meteringv1.GetUsageBreakdownResponse, error) {
	f.GetUsageBreakdownCalls.record(req)
	return f.GetUsageBreakdownCalls.dispatch(ctx, req, f.UnimplementedMeteringServiceServer.GetUsageBreakdown)
}
