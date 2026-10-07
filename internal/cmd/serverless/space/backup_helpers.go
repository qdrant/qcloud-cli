package space

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/qdrant/qcloud-cli/internal/resource"
)

const (
	minRetentionDays = 1
	maxRetentionDays = 365
	allCollections   = "(all)"
)

// retentionFromDays converts a --retention-days value into a retention period.
func retentionFromDays(days uint32) (*durationpb.Duration, error) {
	if days < minRetentionDays || days > maxRetentionDays {
		return nil, fmt.Errorf("--retention-days must be between %d and %d", minRetentionDays, maxRetentionDays)
	}

	return durationpb.New(time.Duration(days) * 24 * time.Hour), nil
}

// formatCollection renders an optional collection name; unset means the whole
// space is covered.
func formatCollection(name *string) string {
	if name == nil {
		return allCollections
	}

	return *name
}

// formatOptionalBytes renders an optional byte count, or "" when unset.
func formatOptionalBytes(v *int64) string {
	if v == nil {
		return ""
	}

	return resource.ByteQuantity(*v).String()
}

// formatOptionalDuration renders an optional duration rounded to seconds, or ""
// when unset.
func formatOptionalDuration(d *durationpb.Duration) string {
	if d == nil {
		return ""
	}

	return d.AsDuration().Round(time.Second).String()
}
