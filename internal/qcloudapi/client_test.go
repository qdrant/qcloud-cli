package qcloudapi

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAuthInterceptor_TraceID(t *testing.T) {
	interceptor := authInterceptor("apikey test-key")

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
		assert.Contains(t, err.Error(), "(trace ID: abc-123)")
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
		assert.Contains(t, err.Error(), "(trace ID: id-1|id-2)")
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

func TestBearerCredentials(t *testing.T) {
	t.Run("sends the current token", func(t *testing.T) {
		creds := bearerCredentials{ts: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "access-token", TokenType: "bearer"})}
		md, err := creds.GetRequestMetadata(context.Background())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"authorization": "Bearer access-token"}, md)
		assert.True(t, creds.RequireTransportSecurity())
	})

	t.Run("a token that cannot be had fails the call as unauthenticated", func(t *testing.T) {
		creds := bearerCredentials{ts: failingTokenSource{errors.New("the login has ended")}}
		_, err := creds.GetRequestMetadata(context.Background())
		require.Error(t, err)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
		assert.Contains(t, err.Error(), "the login has ended")
	})
}

type failingTokenSource struct{ err error }

func (f failingTokenSource) Token() (*oauth2.Token, error) { return nil, f.err }
