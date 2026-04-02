package testutil

import (
	"context"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"
)

// FakeIAMService is a test fake that implements IAMServiceServer.
// Use the *Calls fields to configure responses and inspect captured requests.
type FakeIAMService struct {
	iamv1.UnimplementedIAMServiceServer

	GetAuthenticatedUserCalls MethodSpy[*iamv1.GetAuthenticatedUserRequest, *iamv1.GetAuthenticatedUserResponse]
	ListUsersCalls            MethodSpy[*iamv1.ListUsersRequest, *iamv1.ListUsersResponse]
	ListUserRolesCalls        MethodSpy[*iamv1.ListUserRolesRequest, *iamv1.ListUserRolesResponse]
	AssignUserRolesCalls      MethodSpy[*iamv1.AssignUserRolesRequest, *iamv1.AssignUserRolesResponse]
	ListRolesCalls            MethodSpy[*iamv1.ListRolesRequest, *iamv1.ListRolesResponse]
}

// GetAuthenticatedUser records the call and dispatches via GetAuthenticatedUserCalls.
func (f *FakeIAMService) GetAuthenticatedUser(ctx context.Context, req *iamv1.GetAuthenticatedUserRequest) (*iamv1.GetAuthenticatedUserResponse, error) {
	f.GetAuthenticatedUserCalls.record(req)
	return f.GetAuthenticatedUserCalls.dispatch(ctx, req, f.UnimplementedIAMServiceServer.GetAuthenticatedUser)
}

// ListUsers records the call and dispatches via ListUsersCalls.
func (f *FakeIAMService) ListUsers(ctx context.Context, req *iamv1.ListUsersRequest) (*iamv1.ListUsersResponse, error) {
	f.ListUsersCalls.record(req)
	return f.ListUsersCalls.dispatch(ctx, req, f.UnimplementedIAMServiceServer.ListUsers)
}

// ListUserRoles records the call and dispatches via ListUserRolesCalls.
func (f *FakeIAMService) ListUserRoles(ctx context.Context, req *iamv1.ListUserRolesRequest) (*iamv1.ListUserRolesResponse, error) {
	f.ListUserRolesCalls.record(req)
	return f.ListUserRolesCalls.dispatch(ctx, req, f.UnimplementedIAMServiceServer.ListUserRoles)
}

// AssignUserRoles records the call and dispatches via AssignUserRolesCalls.
func (f *FakeIAMService) AssignUserRoles(ctx context.Context, req *iamv1.AssignUserRolesRequest) (*iamv1.AssignUserRolesResponse, error) {
	f.AssignUserRolesCalls.record(req)
	return f.AssignUserRolesCalls.dispatch(ctx, req, f.UnimplementedIAMServiceServer.AssignUserRoles)
}

// ListRoles records the call and dispatches via ListRolesCalls.
func (f *FakeIAMService) ListRoles(ctx context.Context, req *iamv1.ListRolesRequest) (*iamv1.ListRolesResponse, error) {
	f.ListRolesCalls.record(req)
	return f.ListRolesCalls.dispatch(ctx, req, f.UnimplementedIAMServiceServer.ListRoles)
}
