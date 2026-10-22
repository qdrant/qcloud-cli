package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextPrefix(t *testing.T) {
	t.Parallel()

	for endpoint, want := range map[string]string{
		"grpc.cloud.qdrant.io:443":             "",
		"grpc.development-cloud.qdrant.io:443": "dev-",
		"grpc.staging-cloud.qdrant.io:443":     "staging-",
		"localhost:50051":                      "localhost-",
	} {
		assert.Equal(t, want, contextPrefix(endpoint), endpoint)
	}
}

func TestSlug(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "robert-stam-base-account", slug("Robert Stam - Base Account"))
	assert.Equal(t, "cloud-test-functional-qdrant-com-base-account", slug("cloud-test+functional@qdrant.com - Base Account"))
	assert.Empty(t, slug("  --  "))
}
