package testutil

import (
	"context"

	serverlessquotav1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/quota/v1"
)

// FakeServerlessQuotaService is a test fake that implements the serverless QuotaServiceServer.
// Use the *Calls fields to configure responses and inspect captured requests.
type FakeServerlessQuotaService struct {
	serverlessquotav1.UnimplementedQuotaServiceServer
	GetQuotasCalls MethodSpy[*serverlessquotav1.GetQuotasRequest, *serverlessquotav1.GetQuotasResponse]
}

// GetQuotas records the call and dispatches via GetQuotasCalls.
func (f *FakeServerlessQuotaService) GetQuotas(ctx context.Context, req *serverlessquotav1.GetQuotasRequest) (*serverlessquotav1.GetQuotasResponse, error) {
	f.GetQuotasCalls.record(req)
	return f.GetQuotasCalls.dispatch(ctx, req, f.UnimplementedQuotaServiceServer.GetQuotas)
}
