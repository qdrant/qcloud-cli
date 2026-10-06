package util_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/qdrant/qcloud-cli/internal/cmd/util"
)

func TestNextCronRun(t *testing.T) {
	next, ok := util.NextCronRun("0 2 * * *")
	assert.True(t, ok)
	assert.True(t, next.After(time.Now()))
	assert.Equal(t, 2, next.Hour())
	assert.Equal(t, 0, next.Minute())

	_, ok = util.NextCronRun("not a cron")
	assert.False(t, ok)
}
