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

// ASMetadata is the subset of OIDC / OAuth AS metadata the CLI needs.
type ASMetadata struct {
	Issuer                        string   `json:"issuer"`
	AuthorizationEndpoint         string   `json:"authorization_endpoint"`
	TokenEndpoint                 string   `json:"token_endpoint"`
	DeviceAuthorizationEndpoint   string   `json:"device_authorization_endpoint"`
	RevocationEndpoint            string   `json:"revocation_endpoint"`
	CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported"`
}

func (d Discoverer) fetchASMetadata(ctx context.Context, issuer string) (ASMetadata, error) {
	metaURL, err := url.JoinPath(strings.TrimRight(issuer, "/")+"/", ".well-known/openid-configuration")
	if err != nil {
		return ASMetadata{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, metaURL, nil)
	if err != nil {
		return ASMetadata{}, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := d.client().Do(req)
	if err != nil {
		return ASMetadata{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		return ASMetadata{}, fmt.Errorf("authorization server metadata: HTTP %s", resp.Status)
	}

	var meta ASMetadata
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return ASMetadata{}, fmt.Errorf("decode authorization server metadata: %w", err)
	}

	if meta.AuthorizationEndpoint == "" || meta.TokenEndpoint == "" {
		return ASMetadata{}, fmt.Errorf("authorization server metadata missing authorization or token endpoint")
	}

	return meta, nil
}
