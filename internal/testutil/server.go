package testutil

import (
	"context"
	"net"
	"sync"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"
	clusterauthv2 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/auth/v2"
	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"
	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	platformv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/platform/v1"

	"github.com/qdrant/qcloud-cli/internal/qcloudapi"
	"github.com/qdrant/qcloud-cli/internal/state"
	"github.com/qdrant/qcloud-cli/internal/testutil/mocks"
)

const bufSize = 1024 * 1024

// gRPC server interfaces contain an unexported mustEmbedUnimplemented* method
// that can only be satisfied from outside the package by embedding the
// corresponding Unimplemented* struct. The helper structs below wrap each
// Unimplemented* at depth 2 so that the mock's methods at depth 1 win via
// Go's promotion rules, while the package-qualified mustEmbed method is still
// included in the proxy's method set.

type clusterUnimplHelper struct {
	clusterv1.UnimplementedClusterServiceServer
}
type backupUnimplHelper struct {
	backupv1.UnimplementedBackupServiceServer
}
type bookingUnimplHelper struct {
	bookingv1.UnimplementedBookingServiceServer
}
type platformUnimplHelper struct {
	platformv1.UnimplementedPlatformServiceServer
}
type dbKeyUnimplHelper struct {
	clusterauthv2.UnimplementedDatabaseApiKeyServiceServer
}

type clusterProxy struct {
	clusterUnimplHelper
	*mocks.MockClusterServiceServer
}

type backupProxy struct {
	backupUnimplHelper
	*mocks.MockBackupServiceServer
}

type bookingProxy struct {
	bookingUnimplHelper
	*mocks.MockBookingServiceServer
}

type platformProxy struct {
	platformUnimplHelper
	*mocks.MockPlatformServiceServer
}

type dbKeyProxy struct {
	dbKeyUnimplHelper
	*mocks.MockDatabaseApiKeyServiceServer
}

// TestEnv bundles everything a test needs.
type TestEnv struct {
	State                *state.State
	ClusterServer        *mocks.MockClusterServiceServer
	BookingServer        *mocks.MockBookingServiceServer
	PlatformServer       *mocks.MockPlatformServiceServer
	DatabaseApiKeyServer *mocks.MockDatabaseApiKeyServiceServer
	BackupServer         *mocks.MockBackupServiceServer
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

	server := mocks.NewMockClusterServiceServer(t)
	bookingServer := mocks.NewMockBookingServiceServer(t)
	platformServer := mocks.NewMockPlatformServiceServer(t)
	databaseApiKeyServer := mocks.NewMockDatabaseApiKeyServiceServer(t)
	backupServer := mocks.NewMockBackupServiceServer(t)

	// Start gRPC server on bufconn.
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	clusterv1.RegisterClusterServiceServer(srv, &clusterProxy{MockClusterServiceServer: server})
	bookingv1.RegisterBookingServiceServer(srv, &bookingProxy{MockBookingServiceServer: bookingServer})
	platformv1.RegisterPlatformServiceServer(srv, &platformProxy{MockPlatformServiceServer: platformServer})
	clusterauthv2.RegisterDatabaseApiKeyServiceServer(srv, &dbKeyProxy{MockDatabaseApiKeyServiceServer: databaseApiKeyServer})
	backupv1.RegisterBackupServiceServer(srv, &backupProxy{MockBackupServiceServer: backupServer})

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

	var once sync.Once
	cleanup := func() {
		once.Do(func() {
			_ = client.Close()
			srv.Stop()
		})
	}
	t.Cleanup(cleanup)

	return &TestEnv{
		State:                s,
		ClusterServer:        server,
		BookingServer:        bookingServer,
		PlatformServer:       platformServer,
		DatabaseApiKeyServer: databaseApiKeyServer,
		BackupServer:         backupServer,
		Cleanup:              cleanup,
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
