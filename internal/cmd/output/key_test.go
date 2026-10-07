package output_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
)

type fakeKey struct{ id, name, key string }

func (k fakeKey) GetId() string   { return k.id }
func (k fakeKey) GetName() string { return k.name }
func (k fakeKey) GetKey() string  { return k.key }

func TestCreatedAPIKey(t *testing.T) {
	var buf bytes.Buffer
	output.CreatedAPIKey(&buf, fakeKey{id: "key-1", name: "reader", key: "s3cr3t"})
	assert.Equal(t, "API key key-1 (reader) created.\n\nSave this key now — it will not be shown again:\n  s3cr3t\n", buf.String())
}

func TestKeySecret_Empty(t *testing.T) {
	var buf bytes.Buffer
	output.KeySecret(&buf, "")
	assert.Empty(t, buf.String())
}
