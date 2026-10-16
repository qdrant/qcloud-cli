package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type deviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
	Error                   string `json:"error"`
	ErrorDesc               string `json:"error_description"`
}

func (d Discoverer) deviceLogin(ctx context.Context, opts LoginOptions, disc Discovery, as ASMetadata, scope string) (Token, error) {
	if as.DeviceAuthorizationEndpoint == "" {
		return Token{}, fmt.Errorf("authorization server does not advertise device_authorization_endpoint")
	}

	form := url.Values{
		"client_id": {opts.ClientID},
		"scope":     {scope},
		"resource":  {disc.Resource},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, as.DeviceAuthorizationEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := d.client().Do(req)
	if err != nil {
		return Token{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Token{}, fmt.Errorf("read device code response: %w", err)
	}

	var dc deviceCodeResponse
	if err := json.Unmarshal(body, &dc); err != nil {
		return Token{}, fmt.Errorf("decode device code response: %w", err)
	}

	if dc.Error != "" {
		return Token{}, fmt.Errorf("device authorization: %s", strings.TrimSpace(dc.Error+" "+dc.ErrorDesc))
	}

	if dc.DeviceCode == "" || dc.UserCode == "" {
		return Token{}, fmt.Errorf("device authorization: missing device_code or user_code")
	}

	verify := dc.VerificationURIComplete
	if verify == "" {
		verify = dc.VerificationURI
	}

	if opts.Output != nil {
		fmt.Fprintf(opts.Output, "Open %s and enter code %s\n", verify, dc.UserCode)
	}

	if opts.OpenURL != nil && verify != "" {
		_ = opts.OpenURL(verify)
	}

	interval := time.Duration(dc.Interval) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}

	deadline := opts.Now().Add(10 * time.Minute)
	if dc.ExpiresIn > 0 {
		deadline = opts.Now().Add(time.Duration(dc.ExpiresIn) * time.Second)
	}

	base := Token{
		Endpoint: disc.Hosts.GRPCEndpoint,
		Issuer:   disc.Issuer,
		ClientID: opts.ClientID,
		Resource: disc.Resource,
		Scope:    opts.Scope,
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if opts.Now().After(deadline) {
			return Token{}, fmt.Errorf("device login timed out")
		}

		select {
		case <-ctx.Done():
			return Token{}, ctx.Err()
		case <-ticker.C:
		}

		tokForm := url.Values{
			"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
			"device_code": {dc.DeviceCode},
			"client_id":   {opts.ClientID},
			"resource":    {disc.Resource},
		}

		tok, err := d.exchange(ctx, as.TokenEndpoint, tokForm, opts.Now(), base)
		if err == nil {
			return tok, nil
		}

		if !strings.Contains(err.Error(), "authorization_pending") && !strings.Contains(err.Error(), "slow_down") {
			return Token{}, err
		}
	}
}

// RevokeRefreshToken calls the AS revocation endpoint when advertised.
func (d Discoverer) RevokeRefreshToken(ctx context.Context, tok Token) error {
	if tok.RefreshToken == "" || tok.Issuer == "" {
		return nil
	}

	as, err := d.fetchASMetadata(ctx, tok.Issuer)
	if err != nil {
		return err
	}

	if as.RevocationEndpoint == "" {
		return nil
	}

	form := url.Values{
		"token":           {tok.RefreshToken},
		"token_type_hint": {"refresh_token"},
		"client_id":       {tok.ClientID},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, as.RevocationEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := d.client().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("revoke: HTTP %s", resp.Status)
	}

	return nil
}
