package oauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHosts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		endpoint string
		api      string
		login    string
	}{
		{"grpc.cloud.qdrant.io:443", "api.cloud.qdrant.io", "login.cloud.qdrant.io"},
		{"grpc.staging-cloud.qdrant.io:443", "api.staging-cloud.qdrant.io", "login.staging-cloud.qdrant.io"},
		{"grpc.development-cloud.qdrant.io:443", "api.development-cloud.qdrant.io", "login.development-cloud.qdrant.io"},
		{"grpc.cloud.qdrant.io", "api.cloud.qdrant.io", "login.cloud.qdrant.io"},
		{"api.staging-cloud.qdrant.io:443", "api.staging-cloud.qdrant.io", "login.staging-cloud.qdrant.io"},
		{"", "api.cloud.qdrant.io", "login.cloud.qdrant.io"},
	}

	for _, tt := range tests {
		t.Run(tt.endpoint, func(t *testing.T) {
			t.Parallel()
			h, err := parseHosts(tt.endpoint)
			require.NoError(t, err)
			assert.Equal(t, tt.api, h.apiHost)
			assert.Equal(t, tt.login, h.loginHost)
			assert.Equal(t, "https://"+tt.api, h.resource())
			assert.Equal(t, "https://"+tt.login+"/", h.inferredIssuer())
		})
	}
}

func TestParseHosts_RejectsURL(t *testing.T) {
	t.Parallel()

	_, err := parseHosts("https://grpc.cloud.qdrant.io:443")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "looks like a URL")
}

func TestParseHosts_RejectsUnknownHost(t *testing.T) {
	t.Parallel()

	_, err := parseHosts("localhost:50051")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot infer")
}
