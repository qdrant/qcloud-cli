package testutil

import (
	"context"

	clusterauthv2 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/auth/v2"
)

// FakeDatabaseApiKeyService is a test fake that implements DatabaseApiKeyServiceServer.
// Set the function fields to control responses per test.
type FakeDatabaseApiKeyService struct {
	clusterauthv2.UnimplementedDatabaseApiKeyServiceServer

	ListDatabaseApiKeysFunc  func(context.Context, *clusterauthv2.ListDatabaseApiKeysRequest) (*clusterauthv2.ListDatabaseApiKeysResponse, error)
	CreateDatabaseApiKeyFunc func(context.Context, *clusterauthv2.CreateDatabaseApiKeyRequest) (*clusterauthv2.CreateDatabaseApiKeyResponse, error)
	DeleteDatabaseApiKeyFunc func(context.Context, *clusterauthv2.DeleteDatabaseApiKeyRequest) (*clusterauthv2.DeleteDatabaseApiKeyResponse, error)

	ListDatabaseApiKeysCalls  MethodSpy[*clusterauthv2.ListDatabaseApiKeysRequest, *clusterauthv2.ListDatabaseApiKeysResponse]
	CreateDatabaseApiKeyCalls MethodSpy[*clusterauthv2.CreateDatabaseApiKeyRequest, *clusterauthv2.CreateDatabaseApiKeyResponse]
	DeleteDatabaseApiKeyCalls MethodSpy[*clusterauthv2.DeleteDatabaseApiKeyRequest, *clusterauthv2.DeleteDatabaseApiKeyResponse]
}

// ListDatabaseApiKeys delegates to ListDatabaseApiKeysFunc if set, otherwise dispatches via ListDatabaseApiKeysCalls.
func (f *FakeDatabaseApiKeyService) ListDatabaseApiKeys(ctx context.Context, req *clusterauthv2.ListDatabaseApiKeysRequest) (*clusterauthv2.ListDatabaseApiKeysResponse, error) {
	f.ListDatabaseApiKeysCalls.record(req)
	if f.ListDatabaseApiKeysFunc != nil {
		return f.ListDatabaseApiKeysFunc(ctx, req)
	}
	return f.ListDatabaseApiKeysCalls.dispatch(ctx, req, f.UnimplementedDatabaseApiKeyServiceServer.ListDatabaseApiKeys)
}

// CreateDatabaseApiKey delegates to CreateDatabaseApiKeyFunc if set, otherwise dispatches via CreateDatabaseApiKeyCalls.
func (f *FakeDatabaseApiKeyService) CreateDatabaseApiKey(ctx context.Context, req *clusterauthv2.CreateDatabaseApiKeyRequest) (*clusterauthv2.CreateDatabaseApiKeyResponse, error) {
	f.CreateDatabaseApiKeyCalls.record(req)
	if f.CreateDatabaseApiKeyFunc != nil {
		return f.CreateDatabaseApiKeyFunc(ctx, req)
	}
	return f.CreateDatabaseApiKeyCalls.dispatch(ctx, req, f.UnimplementedDatabaseApiKeyServiceServer.CreateDatabaseApiKey)
}

// DeleteDatabaseApiKey delegates to DeleteDatabaseApiKeyFunc if set, otherwise dispatches via DeleteDatabaseApiKeyCalls.
func (f *FakeDatabaseApiKeyService) DeleteDatabaseApiKey(ctx context.Context, req *clusterauthv2.DeleteDatabaseApiKeyRequest) (*clusterauthv2.DeleteDatabaseApiKeyResponse, error) {
	f.DeleteDatabaseApiKeyCalls.record(req)
	if f.DeleteDatabaseApiKeyFunc != nil {
		return f.DeleteDatabaseApiKeyFunc(ctx, req)
	}
	return f.DeleteDatabaseApiKeyCalls.dispatch(ctx, req, f.UnimplementedDatabaseApiKeyServiceServer.DeleteDatabaseApiKey)
}
