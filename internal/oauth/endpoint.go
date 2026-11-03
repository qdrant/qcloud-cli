package oauth

import (
	"fmt"
	"net"
	"strings"
)

const (
	defaultGRPCHost = "grpc.cloud.qdrant.io"
	wellKnownPRM    = "/.well-known/oauth-protected-resource"
)

// hosts are the HTTPS API and login hosts derived from a gRPC endpoint.
type hosts struct {
	grpcEndpoint string
	apiHost      string
	loginHost    string
}

// resource is the generic Cloud API OAuth resource identifier (JWT aud): https://api.<cluster>.qdrant.io.
func (h hosts) resource() string {
	return "https://" + h.apiHost
}

// inferredIssuer is https://login.<cluster>.qdrant.io/ (trailing slash as the login server issues it).
func (h hosts) inferredIssuer() string {
	return "https://" + h.loginHost + "/"
}

// protectedResourceMetadataURL is the unauthenticated RFC 9728 document on the API host.
func (h hosts) protectedResourceMetadataURL() string {
	return h.resource() + wellKnownPRM
}

// parseHosts maps a gRPC endpoint (QDRANT_CLOUD_ENDPOINT) to API and login hosts.
// grpc.cloud.qdrant.io:443 → api.cloud.qdrant.io / login.cloud.qdrant.io.
func parseHosts(endpoint string) (hosts, error) {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		endpoint = defaultGRPCHost + ":443"
	}

	host := endpoint
	if strings.Contains(endpoint, "://") {
		return hosts{}, fmt.Errorf("endpoint %q looks like a URL; use a gRPC host:port (e.g. grpc.cloud.qdrant.io:443)", endpoint)
	}

	if h, _, err := net.SplitHostPort(endpoint); err == nil {
		host = h
	}

	host = strings.TrimSuffix(host, ".")
	if host == "" {
		return hosts{}, fmt.Errorf("empty host in endpoint %q", endpoint)
	}

	apiHost, ok := rewriteFirstLabel(host, "api")
	if !ok {
		return hosts{}, fmt.Errorf("cannot infer API/login hosts from endpoint %q; expected grpc.<cluster>.qdrant.io", endpoint)
	}

	loginHost, ok := rewriteFirstLabel(host, "login")
	if !ok {
		return hosts{}, fmt.Errorf("cannot infer login host from endpoint %q", endpoint)
	}

	return hosts{
		grpcEndpoint: endpoint,
		apiHost:      apiHost,
		loginHost:    loginHost,
	}, nil
}

func rewriteFirstLabel(host, newLabel string) (string, bool) {
	labels := strings.Split(host, ".")
	if len(labels) < 2 {
		return "", false
	}

	switch labels[0] {
	case "grpc", "api", "login":
		labels[0] = newLabel
		return strings.Join(labels, "."), true
	default:
		return "", false
	}
}
