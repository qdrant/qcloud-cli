package testutil

import (
	"context"

	spaceauthv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/auth/v1"
)

// FakeServerlessSpaceApiKeyService is a test fake that implements the serverless SpaceApiKeyServiceServer.
// Use the *Calls fields to configure responses and inspect captured requests.
type FakeServerlessSpaceApiKeyService struct {
	spaceauthv1.UnimplementedSpaceApiKeyServiceServer

	ListSpaceApiKeysCalls  MethodSpy[*spaceauthv1.ListSpaceApiKeysRequest, *spaceauthv1.ListSpaceApiKeysResponse]
	CreateSpaceApiKeyCalls MethodSpy[*spaceauthv1.CreateSpaceApiKeyRequest, *spaceauthv1.CreateSpaceApiKeyResponse]
	DeleteSpaceApiKeyCalls MethodSpy[*spaceauthv1.DeleteSpaceApiKeyRequest, *spaceauthv1.DeleteSpaceApiKeyResponse]
}

// ListSpaceApiKeys records the call and dispatches via ListSpaceApiKeysCalls.
func (f *FakeServerlessSpaceApiKeyService) ListSpaceApiKeys(ctx context.Context, req *spaceauthv1.ListSpaceApiKeysRequest) (*spaceauthv1.ListSpaceApiKeysResponse, error) {
	f.ListSpaceApiKeysCalls.record(req)
	return f.ListSpaceApiKeysCalls.dispatch(ctx, req, f.UnimplementedSpaceApiKeyServiceServer.ListSpaceApiKeys)
}

// CreateSpaceApiKey records the call and dispatches via CreateSpaceApiKeyCalls.
func (f *FakeServerlessSpaceApiKeyService) CreateSpaceApiKey(ctx context.Context, req *spaceauthv1.CreateSpaceApiKeyRequest) (*spaceauthv1.CreateSpaceApiKeyResponse, error) {
	f.CreateSpaceApiKeyCalls.record(req)
	return f.CreateSpaceApiKeyCalls.dispatch(ctx, req, f.UnimplementedSpaceApiKeyServiceServer.CreateSpaceApiKey)
}

// DeleteSpaceApiKey records the call and dispatches via DeleteSpaceApiKeyCalls.
func (f *FakeServerlessSpaceApiKeyService) DeleteSpaceApiKey(ctx context.Context, req *spaceauthv1.DeleteSpaceApiKeyRequest) (*spaceauthv1.DeleteSpaceApiKeyResponse, error) {
	f.DeleteSpaceApiKeyCalls.record(req)
	return f.DeleteSpaceApiKeyCalls.dispatch(ctx, req, f.UnimplementedSpaceApiKeyServiceServer.DeleteSpaceApiKey)
}
