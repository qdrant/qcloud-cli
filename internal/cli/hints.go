package cli

import (
	"errors"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

// insufficientScopeReason is the ErrorInfo reason the gateway sets when only the OAuth scope stands in
// the way: the user's role allows the call, the access granted to the CLI does not.
const insufficientScopeReason = "INSUFFICIENT_SCOPE"

// ErrorHint returns a line telling the user how to fix err, or "" when there is nothing to add.
func ErrorHint(err error) string {
	var grpcErr interface{ GRPCStatus() *status.Status }
	if !errors.As(err, &grpcErr) {
		return ""
	}

	for _, detail := range grpcErr.GRPCStatus().Details() {
		info, ok := detail.(*errdetails.ErrorInfo)
		if ok && info.GetReason() == insufficientScopeReason {
			scope := info.GetMetadata()["required_scope"]
			if scope == "" {
				scope = "manage"
			}

			return "Your login only has read-only access. Run \"qcloud auth login --scope " + scope + "\" and allow " + scope + " on the consent page."
		}
	}

	return ""
}
