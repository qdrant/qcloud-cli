package testutil

import (
	"context"
	"net"
	"sync"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"
	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/qcloudapi"
	"github.com/qdrant/qcloud-cli/internal/state"
	"github.com/qdrant/qcloud-cli/internal/state/config"
)

const bufSize = 1024 * 1024

// FakeClusterService is a test fake that implements ClusterServiceServer.
// Set the function fields to control responses per test.
type FakeClusterService struct {
	clusterv1.UnimplementedClusterServiceServer

	ListClustersFunc  func(context.Context, *clusterv1.ListClustersRequest) (*clusterv1.ListClustersResponse, error)
	GetClusterFunc    func(context.Context, *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error)
	CreateClusterFunc func(context.Context, *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error)
	DeleteClusterFunc func(context.Context, *clusterv1.DeleteClusterRequest) (*clusterv1.DeleteClusterResponse, error)
}

// ListClusters delegates to ListClustersFunc if set.
func (f *FakeClusterService) ListClusters(ctx context.Context, req *clusterv1.ListClustersRequest) (*clusterv1.ListClustersResponse, error) {
	if f.ListClustersFunc != nil {
		return f.ListClustersFunc(ctx, req)
	}
	return f.UnimplementedClusterServiceServer.ListClusters(ctx, req)
}

// GetCluster delegates to GetClusterFunc if set.
func (f *FakeClusterService) GetCluster(ctx context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
	if f.GetClusterFunc != nil {
		return f.GetClusterFunc(ctx, req)
	}
	return f.UnimplementedClusterServiceServer.GetCluster(ctx, req)
}

// CreateCluster delegates to CreateClusterFunc if set.
func (f *FakeClusterService) CreateCluster(ctx context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error) {
	if f.CreateClusterFunc != nil {
		return f.CreateClusterFunc(ctx, req)
	}
	return f.UnimplementedClusterServiceServer.CreateCluster(ctx, req)
}

// DeleteCluster delegates to DeleteClusterFunc if set.
func (f *FakeClusterService) DeleteCluster(ctx context.Context, req *clusterv1.DeleteClusterRequest) (*clusterv1.DeleteClusterResponse, error) {
	if f.DeleteClusterFunc != nil {
		return f.DeleteClusterFunc(ctx, req)
	}
	return f.UnimplementedClusterServiceServer.DeleteCluster(ctx, req)
}

// FakeBookingService is a test fake that implements BookingServiceServer.
// Set the function fields to control responses per test.
type FakeBookingService struct {
	bookingv1.UnimplementedBookingServiceServer

	ListPackagesFunc func(context.Context, *bookingv1.ListPackagesRequest) (*bookingv1.ListPackagesResponse, error)
}

// ListPackages delegates to ListPackagesFunc if set.
func (f *FakeBookingService) ListPackages(ctx context.Context, req *bookingv1.ListPackagesRequest) (*bookingv1.ListPackagesResponse, error) {
	if f.ListPackagesFunc != nil {
		return f.ListPackagesFunc(ctx, req)
	}
	return f.UnimplementedBookingServiceServer.ListPackages(ctx, req)
}

// RequestCapture is a server-side unary interceptor that records incoming metadata.
type RequestCapture struct {
	mu   sync.Mutex
	last metadata.MD
}

// Last returns the metadata from the most recent request.
func (rc *RequestCapture) Last() metadata.MD {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	return rc.last
}

func (rc *RequestCapture) intercept(
	ctx context.Context,
	req any,
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	rc.mu.Lock()
	rc.last = md
	rc.mu.Unlock()
	return handler(ctx, req)
}

// TestEnv bundles everything a test needs.
type TestEnv struct {
	State         *state.State
	Server        *FakeClusterService
	BookingServer *FakeBookingService
	Capture       *RequestCapture
	Cleanup       func()
}

// Option configures a TestEnv.
type Option func(*envConfig)

type envConfig struct {
	apiKey    string
	accountID string
}

// WithAPIKey sets the API key used by the test client's auth interceptor.
func WithAPIKey(key string) Option {
	return func(c *envConfig) {
		c.apiKey = key
	}
}

// WithAccountID sets the default account ID pre-configured on the test state.
func WithAccountID(id string) Option {
	return func(c *envConfig) {
		c.accountID = id
	}
}

// NewTestEnv creates a test environment with a bufconn-backed gRPC server.
func NewTestEnv(t *testing.T, opts ...Option) *TestEnv {
	t.Helper()

	cfg := &envConfig{apiKey: "test-api-key", accountID: "test-account-id"}
	for _, o := range opts {
		o(cfg)
	}

	fake := &FakeClusterService{}
	fakeBooking := &FakeBookingService{}
	capture := &RequestCapture{}

	// Start gRPC server on bufconn.
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer(grpc.UnaryInterceptor(capture.intercept))
	clusterv1.RegisterClusterServiceServer(srv, fake)
	bookingv1.RegisterBookingServiceServer(srv, fakeBooking)

	go func() {
		_ = srv.Serve(lis)
	}()

	// Dial bufconn with auth interceptor (same as production).
	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(authInterceptor(cfg.apiKey)),
	)
	if err != nil {
		t.Fatalf("failed to dial bufconn: %v", err)
	}

	client := qcloudapi.NewFromConn(conn)

	s := state.New("test")
	s.SetClient(client)

	stateCfg := config.New()
	stateCfg.SetDefault(config.KeyAccountID, cfg.accountID)
	s.Config = stateCfg

	return &TestEnv{
		State:         s,
		Server:        fake,
		BookingServer: fakeBooking,
		Capture:       capture,
		Cleanup: func() {
			_ = conn.Close()
			srv.Stop()
		},
	}
}

// authInterceptor mirrors the production auth interceptor for test clients.
func authInterceptor(apiKey string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "apikey "+apiKey)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
