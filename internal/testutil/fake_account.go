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
	ListAccountsCalls        MethodSpy[*accountv1.ListAccountsRequest, *accountv1.ListAccountsResponse]
	GetAccountCalls          MethodSpy[*accountv1.GetAccountRequest, *accountv1.GetAccountResponse]
	UpdateAccountCalls       MethodSpy[*accountv1.UpdateAccountRequest, *accountv1.UpdateAccountResponse]
	ListAccountMembersCalls  MethodSpy[*accountv1.ListAccountMembersRequest, *accountv1.ListAccountMembersResponse]
	GetAccountMemberCalls    MethodSpy[*accountv1.GetAccountMemberRequest, *accountv1.GetAccountMemberResponse]
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

// ListAccounts records the call and dispatches via ListAccountsCalls.
func (f *FakeAccountService) ListAccounts(ctx context.Context, req *accountv1.ListAccountsRequest) (*accountv1.ListAccountsResponse, error) {
	f.ListAccountsCalls.record(req)
	return f.ListAccountsCalls.dispatch(ctx, req, f.UnimplementedAccountServiceServer.ListAccounts)
}

// GetAccount records the call and dispatches via GetAccountCalls.
func (f *FakeAccountService) GetAccount(ctx context.Context, req *accountv1.GetAccountRequest) (*accountv1.GetAccountResponse, error) {
	f.GetAccountCalls.record(req)
	return f.GetAccountCalls.dispatch(ctx, req, f.UnimplementedAccountServiceServer.GetAccount)
}

// UpdateAccount records the call and dispatches via UpdateAccountCalls.
func (f *FakeAccountService) UpdateAccount(ctx context.Context, req *accountv1.UpdateAccountRequest) (*accountv1.UpdateAccountResponse, error) {
	f.UpdateAccountCalls.record(req)
	return f.UpdateAccountCalls.dispatch(ctx, req, f.UnimplementedAccountServiceServer.UpdateAccount)
}

// ListAccountMembers records the call and dispatches via ListAccountMembersCalls.
func (f *FakeAccountService) ListAccountMembers(ctx context.Context, req *accountv1.ListAccountMembersRequest) (*accountv1.ListAccountMembersResponse, error) {
	f.ListAccountMembersCalls.record(req)
	return f.ListAccountMembersCalls.dispatch(ctx, req, f.UnimplementedAccountServiceServer.ListAccountMembers)
}

// GetAccountMember records the call and dispatches via GetAccountMemberCalls.
func (f *FakeAccountService) GetAccountMember(ctx context.Context, req *accountv1.GetAccountMemberRequest) (*accountv1.GetAccountMemberResponse, error) {
	f.GetAccountMemberCalls.record(req)
	return f.GetAccountMemberCalls.dispatch(ctx, req, f.UnimplementedAccountServiceServer.GetAccountMember)
}
