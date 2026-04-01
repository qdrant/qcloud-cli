package testutil

import (
	"context"

	monitoringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/monitoring/v1"
)

// FakeMonitoringService is a test fake that implements MonitoringServiceServer.
// Use the *Calls fields to configure responses and inspect captured requests.
type FakeMonitoringService struct {
	monitoringv1.UnimplementedMonitoringServiceServer

	GetClusterLogsCalls MethodSpy[*monitoringv1.GetClusterLogsRequest, *monitoringv1.GetClusterLogsResponse]
}

// GetClusterLogs records the call and dispatches via GetClusterLogsCalls.
func (f *FakeMonitoringService) GetClusterLogs(ctx context.Context, req *monitoringv1.GetClusterLogsRequest) (*monitoringv1.GetClusterLogsResponse, error) {
	f.GetClusterLogsCalls.record(req)
	return f.GetClusterLogsCalls.dispatch(ctx, req, f.UnimplementedMonitoringServiceServer.GetClusterLogs)
}
