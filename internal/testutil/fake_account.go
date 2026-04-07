package testutil

import (
	"context"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"
)

// FakeAccountService is a test fake that implements AccountServiceServer.
// Use the *Calls fields to configure responses and inspect captured requests.
type FakeAccountService struct {
	accountv1.UnimplementedAccountServiceServer

	ListAccountInvitesCalls  MethodSpy[*accountv1.ListAccountInvitesRequest, *accountv1.ListAccountInvitesResponse]
	GetAccountInviteCalls    MethodSpy[*accountv1.GetAccountInviteRequest, *accountv1.GetAccountInviteResponse]
	CreateAccountInviteCalls MethodSpy[*accountv1.CreateAccountInviteRequest, *accountv1.CreateAccountInviteResponse]
	DeleteAccountInviteCalls MethodSpy[*accountv1.DeleteAccountInviteRequest, *accountv1.DeleteAccountInviteResponse]
}

// ListAccountInvites records the call and dispatches via ListAccountInvitesCalls.
func (f *FakeAccountService) ListAccountInvites(ctx context.Context, req *accountv1.ListAccountInvitesRequest) (*accountv1.ListAccountInvitesResponse, error) {
	f.ListAccountInvitesCalls.record(req)
	return f.ListAccountInvitesCalls.dispatch(ctx, req, f.UnimplementedAccountServiceServer.ListAccountInvites)
}

// GetAccountInvite records the call and dispatches via GetAccountInviteCalls.
func (f *FakeAccountService) GetAccountInvite(ctx context.Context, req *accountv1.GetAccountInviteRequest) (*accountv1.GetAccountInviteResponse, error) {
	f.GetAccountInviteCalls.record(req)
	return f.GetAccountInviteCalls.dispatch(ctx, req, f.UnimplementedAccountServiceServer.GetAccountInvite)
}

// CreateAccountInvite records the call and dispatches via CreateAccountInviteCalls.
func (f *FakeAccountService) CreateAccountInvite(ctx context.Context, req *accountv1.CreateAccountInviteRequest) (*accountv1.CreateAccountInviteResponse, error) {
	f.CreateAccountInviteCalls.record(req)
	return f.CreateAccountInviteCalls.dispatch(ctx, req, f.UnimplementedAccountServiceServer.CreateAccountInvite)
}

// DeleteAccountInvite records the call and dispatches via DeleteAccountInviteCalls.
func (f *FakeAccountService) DeleteAccountInvite(ctx context.Context, req *accountv1.DeleteAccountInviteRequest) (*accountv1.DeleteAccountInviteResponse, error) {
	f.DeleteAccountInviteCalls.record(req)
	return f.DeleteAccountInviteCalls.dispatch(ctx, req, f.UnimplementedAccountServiceServer.DeleteAccountInvite)
}
