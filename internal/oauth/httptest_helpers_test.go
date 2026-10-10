package oauth

import (
	"net/http"
	"net/url"
	"testing"
)

func rewriteClient(t *testing.T, target string) *http.Client {
	t.Helper()

	u, err := url.Parse(target)
	if err != nil {
		t.Fatal(err)
	}

	return &http.Client{Transport: rewriteTransport{target: u, base: http.DefaultTransport}}
}

type rewriteTransport struct {
	target *url.URL
	base   http.RoundTripper
}

func (r rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = r.target.Scheme
	clone.URL.Host = r.target.Host
	clone.Host = r.target.Host
	return r.base.RoundTrip(clone)
}
