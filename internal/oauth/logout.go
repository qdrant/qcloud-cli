package oauth

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Logout removes endpoint's session. It is best effort: when the refresh token can be read, it is
// revoked first, and a failure to read or revoke it is returned as a warning, not an error, so that
// logout still works when the keychain entry is gone or the login server cannot be reached. err is
// set only when the session could not be removed.
func (c *Client) Logout(ctx context.Context, endpoint string) (warnings []error, err error) {
	tok, ok, loadErr := c.store.load(endpoint)
	switch {
	case loadErr != nil:
		warnings = append(warnings, fmt.Errorf("could not read the refresh token, so it was not revoked: %w", loadErr))
	case ok && tok.RefreshToken != "":
		if err := c.revoke(ctx, tok); err != nil {
			warnings = append(warnings, fmt.Errorf("could not revoke the refresh token: %w", err))
		}
	}

	keychainErr, err := c.store.delete(endpoint)
	if keychainErr != nil {
		warnings = append(warnings, keychainErr)
	}

	return warnings, err
}

// revoke revokes the refresh token at the login server, when it advertises a revocation endpoint.
func (c *Client) revoke(ctx context.Context, tok Token) error {
	as, err := c.fetchASMetadata(ctx, tok.Issuer)
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

	resp, err := c.http.Do(req)
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
