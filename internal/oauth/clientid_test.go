package oauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultClientID(t *testing.T) {
	t.Parallel()

	id, ok := DefaultClientID("https://login.development-cloud.qdrant.io")
	assert.True(t, ok)
	assert.Equal(t, "CjeXCZKv1plzIF9h3vZBGAhM6MLE68fI", id)

	for _, issuer := range []string{
		"https://login.cloud.qdrant.io/",         // TBD
		"https://login.staging-cloud.qdrant.io/", // TBD
		"https://login.example.com/",             // unknown
	} {
		_, ok := DefaultClientID(issuer)
		assert.False(t, ok, issuer)
	}
}
