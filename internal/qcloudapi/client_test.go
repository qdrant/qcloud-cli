package qcloudapi

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestAuthInterceptor_TraceID(t *testing.T) {
	interceptor := authInterceptor("test-key")

	t.Run("no error returns nil", func(t *testing.T) {
		invoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return nil
		}
		err := interceptor(context.Background(), "/test", nil, nil, nil, invoker)
		assert.NoError(t, err)
	})

	t.Run("error without trailer returns original error", func(t *testing.T) {
		orig := errors.New("something broke")
		invoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return orig
		}
		err := interceptor(context.Background(), "/test", nil, nil, nil, invoker)
		require.ErrorIs(t, err, orig)
		assert.NotContains(t, err.Error(), "[")
	})

	t.Run("error with trace ID trailer appends it", func(t *testing.T) {
		orig := errors.New("something broke")
		invoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			for _, o := range opts {
				if t, ok := o.(grpc.TrailerCallOption); ok {
					*t.TrailerAddr = metadata.Pairs(traceIDTrailer, "abc-123")
				}
			}
			return orig
		}
		err := interceptor(context.Background(), "/test", nil, nil, nil, invoker)
		require.ErrorIs(t, err, orig)
		assert.Contains(t, err.Error(), "[abc-123]")
	})

	t.Run("error with multiple trace IDs joins them", func(t *testing.T) {
		orig := errors.New("something broke")
		invoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			for _, o := range opts {
				if t, ok := o.(grpc.TrailerCallOption); ok {
					*t.TrailerAddr = metadata.Pairs(traceIDTrailer, "id-1", traceIDTrailer, "id-2")
				}
			}
			return orig
		}
		err := interceptor(context.Background(), "/test", nil, nil, nil, invoker)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "[id-1|id-2]")
	})

	t.Run("sets authorization header", func(t *testing.T) {
		invoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			md, ok := metadata.FromOutgoingContext(ctx)
			assert.True(t, ok)
			assert.Equal(t, []string{"apikey test-key"}, md.Get("authorization"))
			return nil
		}
		err := interceptor(context.Background(), "/test", nil, nil, nil, invoker)
		assert.NoError(t, err)
	})
}
