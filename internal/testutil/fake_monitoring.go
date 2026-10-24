package testutil

import (
	"context"

	monitoringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/monitoring/v1"
)

// FakeMonitoringService is a test fake that implements MonitoringServiceServer.
// Use the *Calls fields to configure responses and inspect captured requests.
type FakeMonitoringService struct {
	monitoringv1.UnimplementedMonitoringServiceServer

	GetClusterLogsCalls           MethodSpy[*monitoringv1.GetClusterLogsRequest, *monitoringv1.GetClusterLogsResponse]
	GetClusterUsageMetricsCalls   MethodSpy[*monitoringv1.GetClusterUsageMetricsRequest, *monitoringv1.GetClusterUsageMetricsResponse]
	GetClusterSummaryMetricsCalls MethodSpy[*monitoringv1.GetClusterSummaryMetricsRequest, *monitoringv1.GetClusterSummaryMetricsResponse]
	ListClusterAlertsCalls        MethodSpy[*monitoringv1.ListClusterAlertsRequest, *monitoringv1.ListClusterAlertsResponse]
}

// GetClusterLogs records the call and dispatches via GetClusterLogsCalls.
func (f *FakeMonitoringService) GetClusterLogs(ctx context.Context, req *monitoringv1.GetClusterLogsRequest) (*monitoringv1.GetClusterLogsResponse, error) {
	f.GetClusterLogsCalls.record(req)
	return f.GetClusterLogsCalls.dispatch(ctx, req, f.UnimplementedMonitoringServiceServer.GetClusterLogs)
}

// GetClusterSummaryMetrics records the call and dispatches via GetClusterSummaryMetricsCalls.
func (f *FakeMonitoringService) GetClusterSummaryMetrics(ctx context.Context, req *monitoringv1.GetClusterSummaryMetricsRequest) (*monitoringv1.GetClusterSummaryMetricsResponse, error) {
	f.GetClusterSummaryMetricsCalls.record(req)
	return f.GetClusterSummaryMetricsCalls.dispatch(ctx, req, f.UnimplementedMonitoringServiceServer.GetClusterSummaryMetrics)
}

// GetClusterUsageMetrics records the call and dispatches via GetClusterUsageMetricsCalls.
func (f *FakeMonitoringService) GetClusterUsageMetrics(ctx context.Context, req *monitoringv1.GetClusterUsageMetricsRequest) (*monitoringv1.GetClusterUsageMetricsResponse, error) {
	f.GetClusterUsageMetricsCalls.record(req)
	return f.GetClusterUsageMetricsCalls.dispatch(ctx, req, f.UnimplementedMonitoringServiceServer.GetClusterUsageMetrics)
}

// ListClusterAlerts records the call and dispatches via ListClusterAlertsCalls.
func (f *FakeMonitoringService) ListClusterAlerts(ctx context.Context, req *monitoringv1.ListClusterAlertsRequest) (*monitoringv1.ListClusterAlertsResponse, error) {
	f.ListClusterAlertsCalls.record(req)
	return f.ListClusterAlertsCalls.dispatch(ctx, req, f.UnimplementedMonitoringServiceServer.ListClusterAlerts)
}
