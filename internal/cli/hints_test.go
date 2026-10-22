package cli

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrorHint(t *testing.T) {
	t.Parallel()

	st, err := status.New(codes.PermissionDenied, "read only").WithDetails(&errdetails.ErrorInfo{
		Reason:   "INSUFFICIENT_SCOPE",
		Metadata: map[string]string{"required_scope": "manage"},
	})
	require.NoError(t, err)

	// Commands wrap the gRPC error.
	wrapped := fmt.Errorf("failed to create backup: %w", st.Err())
	assert.Equal(t, `Your login only has read-only access. Run "qcloud auth login --scope manage" and allow manage on the consent page.`, ErrorHint(wrapped))

	assert.Empty(t, ErrorHint(status.Error(codes.PermissionDenied, "permission denied - not granted")), "a role denial has no hint")
	assert.Empty(t, ErrorHint(fmt.Errorf("plain error")))
}
