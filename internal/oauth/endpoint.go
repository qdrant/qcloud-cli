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

// Hosts are the HTTPS API and Auth0 issuer derived from a gRPC endpoint.
type Hosts struct {
	GRPCEndpoint string
	APIHost      string
	LoginHost    string
}

// APIBaseURL is https://api.<cluster>.qdrant.io (no trailing slash).
func (h Hosts) APIBaseURL() string {
	return "https://" + h.APIHost
}

// Resource is the generic Cloud API OAuth resource identifier (JWT aud).
func (h Hosts) Resource() string {
	return h.APIBaseURL()
}

// InferredIssuer is https://login.<cluster>.qdrant.io/ (trailing slash as Auth0 issues it).
func (h Hosts) InferredIssuer() string {
	return "https://" + h.LoginHost + "/"
}

// ProtectedResourceMetadataURL is the unauthenticated RFC 9728 document on the API host.
func (h Hosts) ProtectedResourceMetadataURL() string {
	return h.APIBaseURL() + wellKnownPRM
}

// ParseHosts maps a gRPC endpoint (QDRANT_CLOUD_ENDPOINT) to API and login hosts.
// grpc.cloud.qdrant.io:443 → api.cloud.qdrant.io / login.cloud.qdrant.io.
func ParseHosts(endpoint string) (Hosts, error) {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		endpoint = defaultGRPCHost + ":443"
	}

	host := endpoint
	if strings.Contains(endpoint, "://") {
		return Hosts{}, fmt.Errorf("endpoint %q looks like a URL; use a gRPC host:port (e.g. grpc.cloud.qdrant.io:443)", endpoint)
	}

	if h, _, err := net.SplitHostPort(endpoint); err == nil {
		host = h
	}

	host = strings.TrimSuffix(host, ".")
	if host == "" {
		return Hosts{}, fmt.Errorf("empty host in endpoint %q", endpoint)
	}

	apiHost, ok := rewriteFirstLabel(host, "api")
	if !ok {
		return Hosts{}, fmt.Errorf("cannot infer API/login hosts from endpoint %q; expected grpc.<cluster>.qdrant.io", endpoint)
	}

	loginHost, ok := rewriteFirstLabel(host, "login")
	if !ok {
		return Hosts{}, fmt.Errorf("cannot infer login host from endpoint %q", endpoint)
	}

	return Hosts{
		GRPCEndpoint: endpoint,
		APIHost:      apiHost,
		LoginHost:    loginHost,
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
