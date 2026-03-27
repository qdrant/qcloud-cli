package testutil

import (
	"context"

	hybridv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/hybrid/v1"
)

// FakeHybridService is a test fake that implements HybridCloudServiceServer.
// Use the *Calls fields to configure responses and inspect captured requests.
type FakeHybridService struct {
	hybridv1.UnimplementedHybridCloudServiceServer

	ListHybridCloudEnvironmentsCalls  MethodSpy[*hybridv1.ListHybridCloudEnvironmentsRequest, *hybridv1.ListHybridCloudEnvironmentsResponse]
	GetHybridCloudEnvironmentCalls    MethodSpy[*hybridv1.GetHybridCloudEnvironmentRequest, *hybridv1.GetHybridCloudEnvironmentResponse]
	CreateHybridCloudEnvironmentCalls MethodSpy[*hybridv1.CreateHybridCloudEnvironmentRequest, *hybridv1.CreateHybridCloudEnvironmentResponse]
	UpdateHybridCloudEnvironmentCalls MethodSpy[*hybridv1.UpdateHybridCloudEnvironmentRequest, *hybridv1.UpdateHybridCloudEnvironmentResponse]
	DeleteHybridCloudEnvironmentCalls MethodSpy[*hybridv1.DeleteHybridCloudEnvironmentRequest, *hybridv1.DeleteHybridCloudEnvironmentResponse]
	GenerateBootstrapCommandsCalls    MethodSpy[*hybridv1.GenerateBootstrapCommandsRequest, *hybridv1.GenerateBootstrapCommandsResponse]
}

// ListHybridCloudEnvironments records the call and dispatches via ListHybridCloudEnvironmentsCalls.
func (f *FakeHybridService) ListHybridCloudEnvironments(ctx context.Context, req *hybridv1.ListHybridCloudEnvironmentsRequest) (*hybridv1.ListHybridCloudEnvironmentsResponse, error) {
	f.ListHybridCloudEnvironmentsCalls.record(req)
	return f.ListHybridCloudEnvironmentsCalls.dispatch(ctx, req, f.UnimplementedHybridCloudServiceServer.ListHybridCloudEnvironments)
}

// GetHybridCloudEnvironment records the call and dispatches via GetHybridCloudEnvironmentCalls.
func (f *FakeHybridService) GetHybridCloudEnvironment(ctx context.Context, req *hybridv1.GetHybridCloudEnvironmentRequest) (*hybridv1.GetHybridCloudEnvironmentResponse, error) {
	f.GetHybridCloudEnvironmentCalls.record(req)
	return f.GetHybridCloudEnvironmentCalls.dispatch(ctx, req, f.UnimplementedHybridCloudServiceServer.GetHybridCloudEnvironment)
}

// CreateHybridCloudEnvironment records the call and dispatches via CreateHybridCloudEnvironmentCalls.
func (f *FakeHybridService) CreateHybridCloudEnvironment(ctx context.Context, req *hybridv1.CreateHybridCloudEnvironmentRequest) (*hybridv1.CreateHybridCloudEnvironmentResponse, error) {
	f.CreateHybridCloudEnvironmentCalls.record(req)
	return f.CreateHybridCloudEnvironmentCalls.dispatch(ctx, req, f.UnimplementedHybridCloudServiceServer.CreateHybridCloudEnvironment)
}

// UpdateHybridCloudEnvironment records the call and dispatches via UpdateHybridCloudEnvironmentCalls.
func (f *FakeHybridService) UpdateHybridCloudEnvironment(ctx context.Context, req *hybridv1.UpdateHybridCloudEnvironmentRequest) (*hybridv1.UpdateHybridCloudEnvironmentResponse, error) {
	f.UpdateHybridCloudEnvironmentCalls.record(req)
	return f.UpdateHybridCloudEnvironmentCalls.dispatch(ctx, req, f.UnimplementedHybridCloudServiceServer.UpdateHybridCloudEnvironment)
}

// DeleteHybridCloudEnvironment records the call and dispatches via DeleteHybridCloudEnvironmentCalls.
func (f *FakeHybridService) DeleteHybridCloudEnvironment(ctx context.Context, req *hybridv1.DeleteHybridCloudEnvironmentRequest) (*hybridv1.DeleteHybridCloudEnvironmentResponse, error) {
	f.DeleteHybridCloudEnvironmentCalls.record(req)
	return f.DeleteHybridCloudEnvironmentCalls.dispatch(ctx, req, f.UnimplementedHybridCloudServiceServer.DeleteHybridCloudEnvironment)
}

// GenerateBootstrapCommands records the call and dispatches via GenerateBootstrapCommandsCalls.
func (f *FakeHybridService) GenerateBootstrapCommands(ctx context.Context, req *hybridv1.GenerateBootstrapCommandsRequest) (*hybridv1.GenerateBootstrapCommandsResponse, error) {
	f.GenerateBootstrapCommandsCalls.record(req)
	return f.GenerateBootstrapCommandsCalls.dispatch(ctx, req, f.UnimplementedHybridCloudServiceServer.GenerateBootstrapCommands)
}
