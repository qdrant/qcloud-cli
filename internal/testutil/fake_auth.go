package testutil

import (
	"context"

	authv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/auth/v1"
)

// FakeAuthService is a test fake that implements AuthServiceServer.
// Use the *Calls fields to configure responses and inspect captured requests.
type FakeAuthService struct {
	authv1.UnimplementedAuthServiceServer

	ListManagementKeysCalls  MethodSpy[*authv1.ListManagementKeysRequest, *authv1.ListManagementKeysResponse]
	CreateManagementKeyCalls MethodSpy[*authv1.CreateManagementKeyRequest, *authv1.CreateManagementKeyResponse]
	DeleteManagementKeyCalls MethodSpy[*authv1.DeleteManagementKeyRequest, *authv1.DeleteManagementKeyResponse]
}

// ListManagementKeys records the call and dispatches via ListManagementKeysCalls.
func (f *FakeAuthService) ListManagementKeys(ctx context.Context, req *authv1.ListManagementKeysRequest) (*authv1.ListManagementKeysResponse, error) {
	f.ListManagementKeysCalls.record(req)
	return f.ListManagementKeysCalls.dispatch(ctx, req, f.UnimplementedAuthServiceServer.ListManagementKeys)
}

// CreateManagementKey records the call and dispatches via CreateManagementKeyCalls.
func (f *FakeAuthService) CreateManagementKey(ctx context.Context, req *authv1.CreateManagementKeyRequest) (*authv1.CreateManagementKeyResponse, error) {
	f.CreateManagementKeyCalls.record(req)
	return f.CreateManagementKeyCalls.dispatch(ctx, req, f.UnimplementedAuthServiceServer.CreateManagementKey)
}

// DeleteManagementKey records the call and dispatches via DeleteManagementKeyCalls.
func (f *FakeAuthService) DeleteManagementKey(ctx context.Context, req *authv1.DeleteManagementKeyRequest) (*authv1.DeleteManagementKeyResponse, error) {
	f.DeleteManagementKeyCalls.record(req)
	return f.DeleteManagementKeyCalls.dispatch(ctx, req, f.UnimplementedAuthServiceServer.DeleteManagementKey)
}
