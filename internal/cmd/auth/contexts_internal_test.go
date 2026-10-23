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

	assert.Equal(t, "jane-doe-base-account", slug("Jane Doe - Base Account"))
	assert.Equal(t, "dummy-test-example-com-base-account", slug("dummy+test@example.com - Base Account"))
	assert.Empty(t, slug("  --  "))
}
