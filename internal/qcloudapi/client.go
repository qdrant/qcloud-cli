package qcloudapi

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
)

// Client wraps a gRPC connection to the Qdrant Cloud API.
type Client struct {
	conn    *grpc.ClientConn
	cluster clusterv1.ClusterServiceClient
}

// New creates a new gRPC client connected to the given endpoint with the given API key.
func New(ctx context.Context, endpoint, apiKey string) (*Client, error) {
	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")),
		grpc.WithUnaryInterceptor(authInterceptor(apiKey)),
	)
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:    conn,
		cluster: clusterv1.NewClusterServiceClient(conn),
	}, nil
}

// Cluster returns the ClusterService gRPC client.
func (c *Client) Cluster() clusterv1.ClusterServiceClient {
	return c.cluster
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
