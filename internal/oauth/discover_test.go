package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscover_FallsBackToInferredIssuer(t *testing.T) {
	t.Parallel()

	d := Discoverer{HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, assert.AnError
	})}}

	disc, err := d.Discover(context.Background(), "grpc.development-cloud.qdrant.io:443")
	require.NoError(t, err)
	assert.Equal(t, "https://login.development-cloud.qdrant.io/", disc.Issuer)
	assert.Equal(t, "https://api.development-cloud.qdrant.io", disc.Resource)
	assert.Equal(t, sourceEndpointInference, disc.Source)
}

func TestDiscover_UsesProtectedResourceMetadata(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, wellKnownPRM, r.URL.Path)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"resource":              "https://api.development-cloud.qdrant.io",
			"authorization_servers": []string{"https://login.development-cloud.qdrant.io/"},
			"scopes_supported":      []string{"read-only", "manage"},
		})
	}))
	t.Cleanup(srv.Close)

	d := Discoverer{HTTPClient: rewriteClient(t, srv.URL)}
	disc, err := d.Discover(context.Background(), "grpc.development-cloud.qdrant.io:443")
	require.NoError(t, err)
	assert.Equal(t, "https://login.development-cloud.qdrant.io/", disc.Issuer)
	assert.Equal(t, sourceProtectedResource, disc.Source)
	assert.Equal(t, []string{"read-only", "manage"}, disc.Scopes)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
