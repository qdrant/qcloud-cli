package oauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultClientID(t *testing.T) {
	t.Parallel()

	id, ok := defaultClientID("https://login.development-cloud.qdrant.io")
	assert.True(t, ok)
	assert.Equal(t, "CjeXCZKv1plzIF9h3vZBGAhM6MLE68fI", id)

	id, ok = defaultClientID("https://login.staging-cloud.qdrant.io/")
	assert.True(t, ok)
	assert.Equal(t, "LNY7TjjIPiTvLAKjPVsGNlfUz5EdStG9", id)

	for _, issuer := range []string{
		"https://login.cloud.qdrant.io/", // TBD
		"https://login.example.com/",     // unknown
	} {
		_, ok := defaultClientID(issuer)
		assert.False(t, ok, issuer)
	}
}
