package oauth

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/oauth2"
)

// TokenSource returns the access token of endpoint's session, refreshed when it is about to expire.
// A refreshed token is written back to the store. When the login server no longer accepts the refresh
// token (invalid_grant), the session is removed, so the next command asks to log in instead of trying
// again. It returns ErrNotLoggedIn when there is no session.
func (c *Client) TokenSource(ctx context.Context, endpoint string) (oauth2.TokenSource, error) {
	tok, ok, err := c.store.load(endpoint)
	if err != nil {
		return nil, err
	}

	if !ok || tok.AccessToken == "" {
		return nil, ErrNotLoggedIn
	}

	// Sessions stored without an expiry take it from the access token.
	if tok.Expiry.IsZero() {
		tok.Expiry = readTokenClaims(tok.AccessToken).expiry
	}

	// Refreshes happen during the command, possibly long after ctx's caller returned; each request is
	// bounded by the client's timeout instead.
	src := &refreshingSource{ctx: context.WithoutCancel(ctx), client: c, stored: tok}

	return oauth2.ReuseTokenSourceWithExpiry(tok.oauth2(), src, expiryDelta), nil
}

func (tok Token) oauth2() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  tok.AccessToken,
		TokenType:    tok.TokenType,
		RefreshToken: tok.RefreshToken,
		Expiry:       tok.Expiry,
	}
}

// refreshingSource refreshes a stored session. oauth2.ReuseTokenSource calls it only when the current
// access token has expired, and serializes the calls.
type refreshingSource struct {
	// ctx is for the refresh requests: oauth2.TokenSource.Token takes no context, so the source keeps
	// one, as x/oauth2's own refreshing source does.
	ctx    context.Context //nolint:containedctx // see above
	client *Client
	stored Token
}

func (s *refreshingSource) Token() (*oauth2.Token, error) {
	if s.stored.RefreshToken == "" {
		return nil, fmt.Errorf("the login for %s has expired (run \"qcloud auth login\")", s.stored.Endpoint)
	}

	as, err := s.client.fetchASMetadata(s.ctx, s.stored.Issuer)
	if err != nil {
		return nil, fmt.Errorf("refresh the OAuth token: %w", err)
	}

	cfg := s.client.config(as, s.stored.ClientID, nil)
	t, err := cfg.TokenSource(s.client.withHTTP(s.ctx), &oauth2.Token{RefreshToken: s.stored.RefreshToken}).Token()
	if err != nil {
		if re, ok := errors.AsType[*oauth2.RetrieveError](err); ok && re.ErrorCode == "invalid_grant" {
			_, _ = s.client.store.delete(s.stored.Endpoint)

			return nil, fmt.Errorf("the login for %s has ended (run \"qcloud auth login\"): %w", s.stored.Endpoint, tokenError(err))
		}

		return nil, fmt.Errorf("refresh the OAuth token: %w", tokenError(err))
	}

	s.stored = s.stored.with(t)
	if _, err := s.client.store.save(s.stored); err != nil {
		return nil, err
	}

	return s.stored.oauth2(), nil
}

// Token returns a fresh access token for endpoint's session, with the session it belongs to.
func (c *Client) Token(ctx context.Context, endpoint string) (Token, error) {
	ts, err := c.TokenSource(ctx, endpoint)
	if err != nil {
		return Token{}, err
	}

	t, err := ts.Token()
	if err != nil {
		return Token{}, err
	}

	tok, _, err := c.store.info(endpoint)
	if err != nil {
		return Token{}, err
	}

	tok.AccessToken = t.AccessToken
	tok.TokenType = t.Type()
	tok.Expiry = t.Expiry

	return tok, nil
}
