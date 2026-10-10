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

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

func (d Discoverer) exchange(ctx context.Context, tokenURL string, form url.Values, now time.Time, meta Token) (Token, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
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
		return Token{}, fmt.Errorf("read token response: %w", err)
	}

	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return Token{}, fmt.Errorf("decode token response: %w", err)
	}

	if tr.Error != "" {
		desc := tr.ErrorDesc
		if desc == "" {
			desc = tr.Error
		}

		return Token{}, fmt.Errorf("token endpoint: %s", desc)
	}

	if resp.StatusCode != http.StatusOK {
		return Token{}, fmt.Errorf("token endpoint: HTTP %s", resp.Status)
	}

	if tr.AccessToken == "" {
		return Token{}, fmt.Errorf("token endpoint returned no access_token")
	}

	tok := meta
	tok.AccessToken = tr.AccessToken
	if tr.RefreshToken != "" {
		tok.RefreshToken = tr.RefreshToken
	}

	tok.TokenType = tr.TokenType
	if tok.TokenType == "" {
		tok.TokenType = "Bearer"
	}

	if tr.Scope != "" {
		tok.Scope = tr.Scope
	}

	if tr.ExpiresIn > 0 {
		tok.Expiry = now.Add(time.Duration(tr.ExpiresIn) * time.Second)
	}

	return tok, nil
}

// Refresh exchanges a refresh token for a new access token.
func (d Discoverer) Refresh(ctx context.Context, tok Token, now time.Time) (Token, error) {
	if tok.RefreshToken == "" {
		return Token{}, fmt.Errorf("no refresh token; run qcloud auth login")
	}

	meta, err := d.fetchASMetadata(ctx, tok.Issuer)
	if err != nil {
		return Token{}, err
	}

	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {tok.RefreshToken},
		"client_id":     {tok.ClientID},
	}
	if tok.Resource != "" {
		form.Set("resource", tok.Resource)
	}

	return d.exchange(ctx, meta.TokenEndpoint, form, now, tok)
}
