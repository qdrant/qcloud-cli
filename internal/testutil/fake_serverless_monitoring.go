package testutil

import (
	"context"

	serverlessmonitoringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/monitoring/v1"
)

// FakeServerlessMonitoringService is a test fake that implements the serverless MonitoringServiceServer.
// Use the *Calls fields to configure responses and inspect captured requests.
type FakeServerlessMonitoringService struct {
	serverlessmonitoringv1.UnimplementedMonitoringServiceServer
	GetSpaceSummaryMetricsCalls   MethodSpy[*serverlessmonitoringv1.GetSpaceSummaryMetricsRequest, *serverlessmonitoringv1.GetSpaceSummaryMetricsResponse]
	GetSpaceUsageMetricsCalls     MethodSpy[*serverlessmonitoringv1.GetSpaceUsageMetricsRequest, *serverlessmonitoringv1.GetSpaceUsageMetricsResponse]
	GetSpaceInferenceMetricsCalls MethodSpy[*serverlessmonitoringv1.GetSpaceInferenceMetricsRequest, *serverlessmonitoringv1.GetSpaceInferenceMetricsResponse]
	ListSpaceAlertsCalls          MethodSpy[*serverlessmonitoringv1.ListSpaceAlertsRequest, *serverlessmonitoringv1.ListSpaceAlertsResponse]
}

// GetSpaceSummaryMetrics records the call and dispatches via GetSpaceSummaryMetricsCalls.
func (f *FakeServerlessMonitoringService) GetSpaceSummaryMetrics(ctx context.Context, req *serverlessmonitoringv1.GetSpaceSummaryMetricsRequest) (*serverlessmonitoringv1.GetSpaceSummaryMetricsResponse, error) {
	f.GetSpaceSummaryMetricsCalls.record(req)
	return f.GetSpaceSummaryMetricsCalls.dispatch(ctx, req, f.UnimplementedMonitoringServiceServer.GetSpaceSummaryMetrics)
}

// GetSpaceUsageMetrics records the call and dispatches via GetSpaceUsageMetricsCalls.
func (f *FakeServerlessMonitoringService) GetSpaceUsageMetrics(ctx context.Context, req *serverlessmonitoringv1.GetSpaceUsageMetricsRequest) (*serverlessmonitoringv1.GetSpaceUsageMetricsResponse, error) {
	f.GetSpaceUsageMetricsCalls.record(req)
	return f.GetSpaceUsageMetricsCalls.dispatch(ctx, req, f.UnimplementedMonitoringServiceServer.GetSpaceUsageMetrics)
}

// GetSpaceInferenceMetrics records the call and dispatches via GetSpaceInferenceMetricsCalls.
func (f *FakeServerlessMonitoringService) GetSpaceInferenceMetrics(ctx context.Context, req *serverlessmonitoringv1.GetSpaceInferenceMetricsRequest) (*serverlessmonitoringv1.GetSpaceInferenceMetricsResponse, error) {
	f.GetSpaceInferenceMetricsCalls.record(req)
	return f.GetSpaceInferenceMetricsCalls.dispatch(ctx, req, f.UnimplementedMonitoringServiceServer.GetSpaceInferenceMetrics)
}

// ListSpaceAlerts records the call and dispatches via ListSpaceAlertsCalls.
func (f *FakeServerlessMonitoringService) ListSpaceAlerts(ctx context.Context, req *serverlessmonitoringv1.ListSpaceAlertsRequest) (*serverlessmonitoringv1.ListSpaceAlertsResponse, error) {
	f.ListSpaceAlertsCalls.record(req)
	return f.ListSpaceAlertsCalls.dispatch(ctx, req, f.UnimplementedMonitoringServiceServer.ListSpaceAlerts)
}
