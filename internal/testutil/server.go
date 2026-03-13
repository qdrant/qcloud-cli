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
	clusterauthv2 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/auth/v2"
	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	platformv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/platform/v1"

	"github.com/qdrant/qcloud-cli/internal/qcloudapi"
	"github.com/qdrant/qcloud-cli/internal/state"
)

const bufSize = 1024 * 1024

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
	State                *state.State
	Server               *FakeClusterService
	BookingServer        *FakeBookingService
	PlatformServer       *FakePlatformService
	DatabaseApiKeyServer *FakeDatabaseApiKeyService
	Capture              *RequestCapture
	Cleanup              func()
}

// Option configures a TestEnv.
type Option func(*envConfig)

type envConfig struct {
	apiKey    string
	accountID string
	version   string
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

// WithVersion sets the CLI version used in the user-agent header.
func WithVersion(v string) Option {
	return func(c *envConfig) { c.version = v }
}

// newBaseTestEnv sets up the bufconn-backed gRPC server and wires the state,
// but does not pre-populate any config values. Both public constructors call this.
func newBaseTestEnv(t *testing.T, cfg *envConfig) *TestEnv {
	t.Helper()

	fake := &FakeClusterService{}
	fakeBooking := &FakeBookingService{}
	fakePlatform := &FakePlatformService{}
	fakeDatabaseApiKey := &FakeDatabaseApiKeyService{}
	capture := &RequestCapture{}

	// Start gRPC server on bufconn.
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer(grpc.UnaryInterceptor(capture.intercept))
	clusterv1.RegisterClusterServiceServer(srv, fake)
	bookingv1.RegisterBookingServiceServer(srv, fakeBooking)
	platformv1.RegisterPlatformServiceServer(srv, fakePlatform)
	clusterauthv2.RegisterDatabaseApiKeyServiceServer(srv, fakeDatabaseApiKey)

	go func() {
		_ = srv.Serve(lis)
	}()

	// Dial the in-memory server. A few things here that may be surprising:
	//
	// "passthrough:///bufnet" is a gRPC target URI. The "passthrough" scheme
	// tells gRPC's name resolver to skip DNS and use the address as-is.
	// "bufnet" is a throwaway label — it is never resolved. The actual
	// connection is made by WithContextDialer below, which ignores the address
	// and always dials the bufconn listener directly.
	//
	// WithTransportCredentials(insecure) skips TLS. Together these let us run
	// a full gRPC stack in-process without any network or certificate setup.
	dialOpts := []grpc.DialOption{
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	if cfg.version != "" {
		dialOpts = append(dialOpts, grpc.WithUserAgent("qcloud-cli/"+cfg.version))
	}
	client, err := qcloudapi.NewWithDialOptions("passthrough:///bufnet", cfg.apiKey, dialOpts...)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}

	s := state.New(cfg.version)
	s.SetClient(client)

	return &TestEnv{
		State:                s,
		Server:               fake,
		BookingServer:        fakeBooking,
		PlatformServer:       fakePlatform,
		DatabaseApiKeyServer: fakeDatabaseApiKey,
		Capture:              capture,
		Cleanup: func() {
			_ = client.Close()
			srv.Stop()
		},
	}
}

// NewTestEnv creates a test environment with a bufconn-backed gRPC server and
// a pre-populated config. account_id and api_key are set via the highest viper
// priority (Set), so they reliably override any machine environment variables.
// Use this for testing command behaviour where a valid account ID is needed but
// its specific source doesn't matter.
// Defaults: account_id="test-account-id". Override with WithAccountID / WithAPIKey.
func NewTestEnv(t *testing.T, opts ...Option) *TestEnv {
	t.Helper()

	cfg := &envConfig{apiKey: "test-api-key", accountID: "test-account-id"}
	for _, o := range opts {
		o(cfg)
	}

	env := newBaseTestEnv(t, cfg)
	env.State.Config.SetAccountID(cfg.accountID)
	env.State.Config.SetAPIKey(cfg.apiKey)

	return env
}

// NewBareTestEnv creates a test environment with a bufconn-backed gRPC server
// but with no config values pre-populated in viper. Use this when the test
// itself controls how config is loaded — for example, to verify that account_id
// is read from a config file, an environment variable, or a CLI flag.
func NewBareTestEnv(t *testing.T) *TestEnv {
	t.Helper()

	cfg := &envConfig{}
	return newBaseTestEnv(t, cfg)
}
