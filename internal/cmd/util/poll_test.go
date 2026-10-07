package util_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qdrant/qcloud-cli/internal/cmd/util"
)

func TestPollUntilDone_DoneAfterRetries(t *testing.T) {
	calls := 0
	v, err := util.PollUntilDone(t.Context(), time.Second, time.Millisecond, "thing",
		func(context.Context) (string, bool, error) {
			calls++
			return "ok", calls == 3, nil
		})
	require.NoError(t, err)
	assert.Equal(t, "ok", v)
	assert.Equal(t, 3, calls)
}

func TestPollUntilDone_Error(t *testing.T) {
	boom := errors.New("boom")
	_, err := util.PollUntilDone(t.Context(), time.Second, time.Millisecond, "thing",
		func(context.Context) (string, bool, error) {
			return "", false, boom
		})
	require.ErrorIs(t, err, boom)
}

func TestPollUntilDone_Timeout(t *testing.T) {
	_, err := util.PollUntilDone(t.Context(), 20*time.Millisecond, time.Millisecond, "thing",
		func(context.Context) (string, bool, error) {
			return "", false, nil
		})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out waiting for thing")
}
