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
}

// ListDatabaseApiKeys delegates to ListDatabaseApiKeysFunc if set.
func (f *FakeDatabaseApiKeyService) ListDatabaseApiKeys(ctx context.Context, req *clusterauthv2.ListDatabaseApiKeysRequest) (*clusterauthv2.ListDatabaseApiKeysResponse, error) {
	if f.ListDatabaseApiKeysFunc != nil {
		return f.ListDatabaseApiKeysFunc(ctx, req)
	}
	return f.UnimplementedDatabaseApiKeyServiceServer.ListDatabaseApiKeys(ctx, req)
}

// CreateDatabaseApiKey delegates to CreateDatabaseApiKeyFunc if set.
func (f *FakeDatabaseApiKeyService) CreateDatabaseApiKey(ctx context.Context, req *clusterauthv2.CreateDatabaseApiKeyRequest) (*clusterauthv2.CreateDatabaseApiKeyResponse, error) {
	if f.CreateDatabaseApiKeyFunc != nil {
		return f.CreateDatabaseApiKeyFunc(ctx, req)
	}
	return f.UnimplementedDatabaseApiKeyServiceServer.CreateDatabaseApiKey(ctx, req)
}

// DeleteDatabaseApiKey delegates to DeleteDatabaseApiKeyFunc if set.
func (f *FakeDatabaseApiKeyService) DeleteDatabaseApiKey(ctx context.Context, req *clusterauthv2.DeleteDatabaseApiKeyRequest) (*clusterauthv2.DeleteDatabaseApiKeyResponse, error) {
	if f.DeleteDatabaseApiKeyFunc != nil {
		return f.DeleteDatabaseApiKeyFunc(ctx, req)
	}
	return f.UnimplementedDatabaseApiKeyServiceServer.DeleteDatabaseApiKey(ctx, req)
}
