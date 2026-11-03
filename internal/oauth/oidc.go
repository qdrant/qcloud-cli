package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// asMetadata is the subset of the login server's OpenID Connect metadata the CLI needs.
type asMetadata struct {
	Issuer                      string `json:"issuer"`
	AuthorizationEndpoint       string `json:"authorization_endpoint"`
	TokenEndpoint               string `json:"token_endpoint"`
	DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint"`
	RevocationEndpoint          string `json:"revocation_endpoint"`
}

func (c *Client) fetchASMetadata(ctx context.Context, issuer string) (asMetadata, error) {
	metaURL, err := url.JoinPath(strings.TrimRight(issuer, "/")+"/", ".well-known/openid-configuration")
	if err != nil {
		return asMetadata{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, metaURL, nil)
	if err != nil {
		return asMetadata{}, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return asMetadata{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		return asMetadata{}, fmt.Errorf("authorization server metadata: HTTP %s", resp.Status)
	}

	var meta asMetadata
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return asMetadata{}, fmt.Errorf("decode authorization server metadata: %w", err)
	}

	if meta.AuthorizationEndpoint == "" || meta.TokenEndpoint == "" {
		return asMetadata{}, fmt.Errorf("authorization server metadata missing authorization or token endpoint")
	}

	return meta, nil
}
