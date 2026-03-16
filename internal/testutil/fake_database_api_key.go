package testutil

import (
	"context"

	clusterauthv2 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/auth/v2"
)

// FakeDatabaseApiKeyService is a test fake that implements DatabaseApiKeyServiceServer.
// Use the *Calls fields to configure responses and inspect captured requests.
type FakeDatabaseApiKeyService struct {
	clusterauthv2.UnimplementedDatabaseApiKeyServiceServer

	ListDatabaseApiKeysCalls  MethodSpy[*clusterauthv2.ListDatabaseApiKeysRequest, *clusterauthv2.ListDatabaseApiKeysResponse]
	CreateDatabaseApiKeyCalls MethodSpy[*clusterauthv2.CreateDatabaseApiKeyRequest, *clusterauthv2.CreateDatabaseApiKeyResponse]
	DeleteDatabaseApiKeyCalls MethodSpy[*clusterauthv2.DeleteDatabaseApiKeyRequest, *clusterauthv2.DeleteDatabaseApiKeyResponse]
}

// ListDatabaseApiKeys records the call and dispatches via ListDatabaseApiKeysCalls.
func (f *FakeDatabaseApiKeyService) ListDatabaseApiKeys(ctx context.Context, req *clusterauthv2.ListDatabaseApiKeysRequest) (*clusterauthv2.ListDatabaseApiKeysResponse, error) {
	f.ListDatabaseApiKeysCalls.record(req)
	return f.ListDatabaseApiKeysCalls.dispatch(ctx, req, f.UnimplementedDatabaseApiKeyServiceServer.ListDatabaseApiKeys)
}

// CreateDatabaseApiKey records the call and dispatches via CreateDatabaseApiKeyCalls.
func (f *FakeDatabaseApiKeyService) CreateDatabaseApiKey(ctx context.Context, req *clusterauthv2.CreateDatabaseApiKeyRequest) (*clusterauthv2.CreateDatabaseApiKeyResponse, error) {
	f.CreateDatabaseApiKeyCalls.record(req)
	return f.CreateDatabaseApiKeyCalls.dispatch(ctx, req, f.UnimplementedDatabaseApiKeyServiceServer.CreateDatabaseApiKey)
}

// DeleteDatabaseApiKey records the call and dispatches via DeleteDatabaseApiKeyCalls.
func (f *FakeDatabaseApiKeyService) DeleteDatabaseApiKey(ctx context.Context, req *clusterauthv2.DeleteDatabaseApiKeyRequest) (*clusterauthv2.DeleteDatabaseApiKeyResponse, error) {
	f.DeleteDatabaseApiKeyCalls.record(req)
	return f.DeleteDatabaseApiKeyCalls.dispatch(ctx, req, f.UnimplementedDatabaseApiKeyServiceServer.DeleteDatabaseApiKey)
}
