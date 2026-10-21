package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	sourceProtectedResource = "protected-resource metadata"
	sourceEndpointInference = "inferred from endpoint"
)

// Discovery is the OAuth setup for one Cloud environment.
type Discovery struct {
	Hosts    Hosts
	Resource string
	Issuer   string
	Scopes   []string
	Source   string
	// MetadataError is why the protected-resource metadata could not be used, when Source is
	// endpoint inference. The inferred values work too; this explains why they were inferred.
	MetadataError string
}

// Discoverer resolves the authorization server for a gRPC endpoint.
type Discoverer struct {
	HTTPClient *http.Client
}

func (d Discoverer) client() *http.Client {
	if d.HTTPClient != nil {
		return d.HTTPClient
	}

	return &http.Client{Timeout: 10 * time.Second}
}

// Discover maps endpoint → API host, then prefers the gateway's unauthenticated
// PRM document for the issuer. If OAuth is not enabled yet, it infers login.<cluster>.
func (d Discoverer) Discover(ctx context.Context, endpoint string) (Discovery, error) {
	hosts, err := ParseHosts(endpoint)
	if err != nil {
		return Discovery{}, err
	}

	out := Discovery{
		Hosts:    hosts,
		Resource: hosts.Resource(),
		Issuer:   hosts.InferredIssuer(),
		Scopes:   []string{"read-only", "manage"},
		Source:   sourceEndpointInference,
	}

	prm, err := d.fetchPRM(ctx, hosts.ProtectedResourceMetadataURL())
	if err != nil {
		// OAuth may not be enabled yet; hostname inference still works.
		out.MetadataError = err.Error()

		return out, nil //nolint:nilerr // the inferred values are the fallback; the reason is in MetadataError
	}

	if len(prm.AuthorizationServers) > 0 && prm.AuthorizationServers[0] != "" {
		out.Issuer = canonicalizeIssuer(prm.AuthorizationServers[0])
		out.Source = sourceProtectedResource
	}

	if prm.Resource != "" {
		out.Resource = strings.TrimRight(prm.Resource, "/")
	}

	if len(prm.ScopesSupported) > 0 {
		out.Scopes = prm.ScopesSupported
	}

	return out, nil
}

type prmDocument struct {
	Resource             string   `json:"resource"`
	AuthorizationServers []string `json:"authorization_servers"`
	ScopesSupported      []string `json:"scopes_supported"`
}

func (d Discoverer) fetchPRM(ctx context.Context, url string) (prmDocument, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return prmDocument{}, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := d.client().Do(req)
	if err != nil {
		return prmDocument{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		return prmDocument{}, fmt.Errorf("protected-resource metadata: HTTP %s", resp.Status)
	}

	var doc prmDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return prmDocument{}, fmt.Errorf("decode protected-resource metadata: %w", err)
	}

	return doc, nil
}

func canonicalizeIssuer(issuer string) string {
	issuer = strings.TrimSpace(issuer)
	if issuer == "" {
		return issuer
	}

	if !strings.HasSuffix(issuer, "/") {
		return issuer + "/"
	}

	return issuer
}
