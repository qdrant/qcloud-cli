package testutil

import (
	"context"

	spacev1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/v1"
)

// FakeServerlessSpaceService is a test fake that implements the serverless SpaceServiceServer.
// Use the *Calls fields to configure responses and inspect captured requests.
type FakeServerlessSpaceService struct {
	spacev1.UnimplementedSpaceServiceServer

	ListSpacesCalls            MethodSpy[*spacev1.ListSpacesRequest, *spacev1.ListSpacesResponse]
	GetSpaceCalls              MethodSpy[*spacev1.GetSpaceRequest, *spacev1.GetSpaceResponse]
	CreateSpaceCalls           MethodSpy[*spacev1.CreateSpaceRequest, *spacev1.CreateSpaceResponse]
	CreateSpaceFromBackupCalls MethodSpy[*spacev1.CreateSpaceFromBackupRequest, *spacev1.CreateSpaceFromBackupResponse]
	UpdateSpaceCalls           MethodSpy[*spacev1.UpdateSpaceRequest, *spacev1.UpdateSpaceResponse]
	DeleteSpaceCalls           MethodSpy[*spacev1.DeleteSpaceRequest, *spacev1.DeleteSpaceResponse]
	SuggestSpaceNameCalls      MethodSpy[*spacev1.SuggestSpaceNameRequest, *spacev1.SuggestSpaceNameResponse]
}

// ListSpaces records the call and dispatches via ListSpacesCalls.
func (f *FakeServerlessSpaceService) ListSpaces(ctx context.Context, req *spacev1.ListSpacesRequest) (*spacev1.ListSpacesResponse, error) {
	f.ListSpacesCalls.record(req)
	return f.ListSpacesCalls.dispatch(ctx, req, f.UnimplementedSpaceServiceServer.ListSpaces)
}

// GetSpace records the call and dispatches via GetSpaceCalls.
func (f *FakeServerlessSpaceService) GetSpace(ctx context.Context, req *spacev1.GetSpaceRequest) (*spacev1.GetSpaceResponse, error) {
	f.GetSpaceCalls.record(req)
	return f.GetSpaceCalls.dispatch(ctx, req, f.UnimplementedSpaceServiceServer.GetSpace)
}

// CreateSpace records the call and dispatches via CreateSpaceCalls.
func (f *FakeServerlessSpaceService) CreateSpace(ctx context.Context, req *spacev1.CreateSpaceRequest) (*spacev1.CreateSpaceResponse, error) {
	f.CreateSpaceCalls.record(req)
	return f.CreateSpaceCalls.dispatch(ctx, req, f.UnimplementedSpaceServiceServer.CreateSpace)
}

// CreateSpaceFromBackup records the call and dispatches via CreateSpaceFromBackupCalls.
func (f *FakeServerlessSpaceService) CreateSpaceFromBackup(ctx context.Context, req *spacev1.CreateSpaceFromBackupRequest) (*spacev1.CreateSpaceFromBackupResponse, error) {
	f.CreateSpaceFromBackupCalls.record(req)
	return f.CreateSpaceFromBackupCalls.dispatch(ctx, req, f.UnimplementedSpaceServiceServer.CreateSpaceFromBackup)
}

// UpdateSpace records the call and dispatches via UpdateSpaceCalls.
func (f *FakeServerlessSpaceService) UpdateSpace(ctx context.Context, req *spacev1.UpdateSpaceRequest) (*spacev1.UpdateSpaceResponse, error) {
	f.UpdateSpaceCalls.record(req)
	return f.UpdateSpaceCalls.dispatch(ctx, req, f.UnimplementedSpaceServiceServer.UpdateSpace)
}

// DeleteSpace records the call and dispatches via DeleteSpaceCalls.
func (f *FakeServerlessSpaceService) DeleteSpace(ctx context.Context, req *spacev1.DeleteSpaceRequest) (*spacev1.DeleteSpaceResponse, error) {
	f.DeleteSpaceCalls.record(req)
	return f.DeleteSpaceCalls.dispatch(ctx, req, f.UnimplementedSpaceServiceServer.DeleteSpace)
}

// SuggestSpaceName records the call and dispatches via SuggestSpaceNameCalls.
func (f *FakeServerlessSpaceService) SuggestSpaceName(ctx context.Context, req *spacev1.SuggestSpaceNameRequest) (*spacev1.SuggestSpaceNameResponse, error) {
	f.SuggestSpaceNameCalls.record(req)
	return f.SuggestSpaceNameCalls.dispatch(ctx, req, f.UnimplementedSpaceServiceServer.SuggestSpaceName)
}
