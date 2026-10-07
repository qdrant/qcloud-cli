package output_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
)

func TestRetentionPeriod(t *testing.T) {
	assert.Equal(t, "indefinite", output.RetentionPeriod(nil))
	assert.Equal(t, "1 day", output.RetentionPeriod(durationpb.New(24*time.Hour)))
	assert.Equal(t, "30 days", output.RetentionPeriod(durationpb.New(30*24*time.Hour)))
	assert.Equal(t, "1 day 12 hours", output.RetentionPeriod(durationpb.New(36*time.Hour)))
}
