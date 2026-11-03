package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	sourceProtectedResource = "protected-resource metadata"
	sourceEndpointInference = "inferred from endpoint"
	sourceLoginURLFlag      = "flag --login-url"
)

// Discovery is the OAuth setup for one Cloud environment.
type Discovery struct {
	// Endpoint is the gRPC endpoint the setup is for.
	Endpoint string `json:"endpoint"`
	// Resource is the API the access token is for (its audience).
	Resource string `json:"resource"`
	// Issuer is the login server.
	Issuer string `json:"issuer"`
	// Source says where Issuer comes from.
	Source string   `json:"source"`
	Scopes []string `json:"scopes"`
	// MetadataError is why the protected-resource metadata could not be used, when Source is
	// endpoint inference. The inferred values work too; this explains why they were inferred.
	MetadataError string `json:"metadata_error,omitempty"`
}

// Discover maps endpoint → API host, then prefers the gateway's unauthenticated
// PRM document for the issuer. If OAuth is not enabled yet, it infers login.<cluster>.
func (c *Client) Discover(ctx context.Context, endpoint string) (Discovery, error) {
	h, err := parseHosts(endpoint)
	if err != nil {
		return Discovery{}, err
	}

	out := Discovery{
		Endpoint: h.grpcEndpoint,
		Resource: h.resource(),
		Issuer:   h.inferredIssuer(),
		Scopes:   []string{"read-only", "manage"},
		Source:   sourceEndpointInference,
	}

	prm, err := c.fetchPRM(ctx, h.protectedResourceMetadataURL())
	if err != nil {
		// OAuth may not be enabled yet; hostname inference still works.
		out.MetadataError = err.Error()

		return out, nil //nolint:nilerr // the inferred values are the fallback; the reason is in MetadataError
	}

	// RFC 9728 §3.3: a document for another resource must not be used, or an endpoint
	// could point the login at a real issuer and have the token sent back to it.
	if got := strings.TrimRight(prm.Resource, "/"); got != out.Resource {
		out.MetadataError = fmt.Sprintf("protected-resource metadata: resource %q does not match %q", prm.Resource, out.Resource)

		return out, nil
	}

	if len(prm.AuthorizationServers) == 0 || prm.AuthorizationServers[0] == "" {
		out.MetadataError = "protected-resource metadata lists no authorization_servers"

		return out, nil
	}

	out.Issuer = canonicalizeIssuer(prm.AuthorizationServers[0])
	out.Source = sourceProtectedResource
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

func (c *Client) fetchPRM(ctx context.Context, url string) (prmDocument, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return prmDocument{}, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
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
