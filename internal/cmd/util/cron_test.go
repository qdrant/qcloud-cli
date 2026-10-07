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

func TestNextCronRun_SecondsField(t *testing.T) {
	next, ok := util.NextCronRun("30 0 2 * * *")
	assert.True(t, ok)
	assert.Equal(t, 2, next.Hour())
	assert.Equal(t, 0, next.Minute())
	assert.Equal(t, 30, next.Second())
}

func TestNextCronRun_Descriptors(t *testing.T) {
	next, ok := util.NextCronRun("@daily")
	assert.True(t, ok)
	assert.Equal(t, 0, next.Hour())

	next, ok = util.NextCronRun("@every 1h")
	assert.True(t, ok)
	assert.True(t, next.After(time.Now()))
}
