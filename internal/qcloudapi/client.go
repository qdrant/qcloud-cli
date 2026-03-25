package qcloudapi

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"
	clusterauthv2 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/auth/v2"
	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"
	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	hybridv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/hybrid/v1"
	platformv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/platform/v1"
)

// Client wraps a gRPC connection to the Qdrant Cloud API.
type Client struct {
	conn           *grpc.ClientConn
	cluster        clusterv1.ClusterServiceClient
	booking        bookingv1.BookingServiceClient
	platform       platformv1.PlatformServiceClient
	databaseApiKey clusterauthv2.DatabaseApiKeyServiceClient
	backup         backupv1.BackupServiceClient
	hybrid         hybridv1.HybridCloudServiceClient
}

// New creates a new gRPC client connected to the given endpoint with the given API key.
func New(ctx context.Context, endpoint, apiKey, version string) (*Client, error) {
	return NewWithDialOptions(endpoint, apiKey,
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")),
		grpc.WithUserAgent("qcloud-cli/"+version),
	)
}

// NewWithDialOptions creates a Client with the auth interceptor always applied,
// plus any additional dial options (e.g. custom transport for testing).
func NewWithDialOptions(endpoint, apiKey string, opts ...grpc.DialOption) (*Client, error) {
	all := append([]grpc.DialOption{grpc.WithUnaryInterceptor(authInterceptor(apiKey))}, opts...)
	conn, err := grpc.NewClient(endpoint, all...)
	if err != nil {
		return nil, err
	}
	return newFromConn(conn), nil
}

func newFromConn(conn *grpc.ClientConn) *Client {
	return &Client{
		conn:           conn,
		cluster:        clusterv1.NewClusterServiceClient(conn),
		booking:        bookingv1.NewBookingServiceClient(conn),
		platform:       platformv1.NewPlatformServiceClient(conn),
		databaseApiKey: clusterauthv2.NewDatabaseApiKeyServiceClient(conn),
		backup:         backupv1.NewBackupServiceClient(conn),
		hybrid:         hybridv1.NewHybridCloudServiceClient(conn),
	}
}

// Cluster returns the ClusterService gRPC client.
func (c *Client) Cluster() clusterv1.ClusterServiceClient {
	return c.cluster
}

// Booking returns the BookingService gRPC client.
func (c *Client) Booking() bookingv1.BookingServiceClient {
	return c.booking
}

// Platform returns the PlatformService gRPC client.
func (c *Client) Platform() platformv1.PlatformServiceClient {
	return c.platform
}

// DatabaseApiKey returns the DatabaseApiKeyService gRPC client.
func (c *Client) DatabaseApiKey() clusterauthv2.DatabaseApiKeyServiceClient {
	return c.databaseApiKey
}

// Backup returns the BackupService gRPC client.
func (c *Client) Backup() backupv1.BackupServiceClient {
	return c.backup
}

// Hybrid returns the HybridCloudService gRPC client.
func (c *Client) Hybrid() hybridv1.HybridCloudServiceClient {
	return c.hybrid
}

// Close closes the underlying gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

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
