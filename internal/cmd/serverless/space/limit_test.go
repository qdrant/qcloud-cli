package space_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	spacev1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/v1"

	"github.com/qdrant/qcloud-cli/internal/resource"
	"github.com/qdrant/qcloud-cli/internal/testutil"
)

// spaceWithLimits returns existingSpace with a hard 10 GiB storage limit and a
// hard $50 monthly spend limit, both alerting at 80%.
func spaceWithLimits() *spacev1.Space {
	sp := existingSpace()
	sp.Configuration.StorageLimit = &spacev1.StorageLimit{
		MaxSize: uint64(10 * resource.GiB), Type: spacev1.LimitType_LIMIT_TYPE_HARD, AlertThresholdPercent: new(uint32(80)),
	}
	sp.Configuration.MonthlySpendLimit = &spacev1.SpendLimit{
		MaxAmountMillicents: 5_000_000, Type: spacev1.LimitType_LIMIT_TYPE_HARD, AlertThresholdPercent: new(uint32(80)), Currency: "USD",
	}
	return sp
}

func setupUpdateWith(env *testutil.TestEnv, sp *spacev1.Space) {
	env.ServerlessSpaceServer.GetSpaceCalls.Returns(&spacev1.GetSpaceResponse{Space: sp}, nil)
	env.ServerlessSpaceServer.UpdateSpaceCalls.Always(func(_ context.Context, req *spacev1.UpdateSpaceRequest) (*spacev1.UpdateSpaceResponse, error) {
		return &spacev1.UpdateSpaceResponse{Space: req.GetSpace()}, nil
	})
}

func TestSpaceCreate_Limits(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessSpaceServer.CreateSpaceCalls.Always(func(_ context.Context, req *spacev1.CreateSpaceRequest) (*spacev1.CreateSpaceResponse, error) {
		return &spacev1.CreateSpaceResponse{Space: req.GetSpace()}, nil
	})

	_, _, err := testutil.Exec(t, env, "serverless", "space", "create", "--cloud-region", "eu-central-1", "--name", "my-space",
		"--storage-limit", "100GiB",
		"--spend-limit", "49.99", "--spend-limit-type", "soft", "--spend-limit-alert-percent", "90",
	)
	require.NoError(t, err)

	req, ok := env.ServerlessSpaceServer.CreateSpaceCalls.Last()
	require.True(t, ok)
	cfg := req.GetSpace().GetConfiguration()

	storage := cfg.GetStorageLimit()
	require.NotNil(t, storage)
	assert.Equal(t, uint64(100*resource.GiB), storage.GetMaxSize())
	assert.Equal(t, spacev1.LimitType_LIMIT_TYPE_UNSPECIFIED, storage.GetType(), "the server defaults to hard")
	assert.Nil(t, storage.AlertThresholdPercent, "the server defaults to 80%")

	spend := cfg.GetMonthlySpendLimit()
	require.NotNil(t, spend)
	assert.Equal(t, uint64(4_999_000), spend.GetMaxAmountMillicents(), "$49.99 in millicents (1/1000 of a cent)")
	assert.Equal(t, spacev1.LimitType_LIMIT_TYPE_SOFT, spend.GetType())
	assert.Equal(t, uint32(90), spend.GetAlertThresholdPercent())
}

func TestSpaceCreate_NoLimitsByDefault(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessSpaceServer.CreateSpaceCalls.Always(func(_ context.Context, req *spacev1.CreateSpaceRequest) (*spacev1.CreateSpaceResponse, error) {
		return &spacev1.CreateSpaceResponse{Space: req.GetSpace()}, nil
	})

	_, _, err := testutil.Exec(t, env, "serverless", "space", "create", "--cloud-region", "eu-central-1", "--name", "my-space")
	require.NoError(t, err)

	req, ok := env.ServerlessSpaceServer.CreateSpaceCalls.Last()
	require.True(t, ok)
	assert.Nil(t, req.GetSpace().GetConfiguration().GetStorageLimit())
	assert.Nil(t, req.GetSpace().GetConfiguration().GetMonthlySpendLimit())
}

func TestSpaceCreate_InvalidLimitFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "invalid storage limit", args: []string{"--storage-limit", "lots"}},
		{name: "zero storage limit", args: []string{"--storage-limit", "0"}},
		{name: "invalid limit type", args: []string{"--storage-limit", "10GiB", "--storage-limit-type", "medium"}},
		{name: "alert percent too low", args: []string{"--storage-limit", "10GiB", "--storage-limit-alert-percent", "0"}},
		{name: "alert percent too high", args: []string{"--spend-limit", "50", "--spend-limit-alert-percent", "100"}},
		{name: "invalid spend limit", args: []string{"--spend-limit", "fifty"}},
		{name: "spend limit sub-cent", args: []string{"--spend-limit", "1.234"}},
		{name: "zero spend limit", args: []string{"--spend-limit", "0"}},
		{name: "type without limit", args: []string{"--spend-limit-type", "soft"}},
		{name: "alert percent without limit", args: []string{"--storage-limit-alert-percent", "90"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := testutil.NewTestEnv(t)

			args := append([]string{"serverless", "space", "create", "--cloud-region", "r", "--name", "my-space"}, tt.args...)
			_, _, err := testutil.Exec(t, env, args...)
			require.Error(t, err)
			assert.Equal(t, 0, env.ServerlessSpaceServer.CreateSpaceCalls.Count())
		})
	}
}

func TestSpaceUpdate_SetLimit(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdate(env)

	_, _, err := testutil.Exec(t, env, "serverless", "space", "update", "space-abc",
		"--storage-limit", "512MiB", "--storage-limit-type", "hard", "--storage-limit-alert-percent", "75",
	)
	require.NoError(t, err)

	req, ok := env.ServerlessSpaceServer.UpdateSpaceCalls.Last()
	require.True(t, ok)
	assert.Equal(t, []string{"configuration.storage_limit"}, req.GetUpdateMask().GetPaths())
	storage := req.GetSpace().GetConfiguration().GetStorageLimit()
	require.NotNil(t, storage)
	assert.Equal(t, uint64(512*resource.MiB), storage.GetMaxSize())
	assert.Equal(t, spacev1.LimitType_LIMIT_TYPE_HARD, storage.GetType())
	assert.Equal(t, uint32(75), storage.GetAlertThresholdPercent())
	assert.Nil(t, req.GetSpace().GetConfiguration().GetMonthlySpendLimit())
}

func TestSpaceUpdate_ChangeLimitTypeKeepsTheRest(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateWith(env, spaceWithLimits())

	_, _, err := testutil.Exec(t, env, "serverless", "space", "update", "space-abc", "--storage-limit-type", "soft")
	require.NoError(t, err)

	req, ok := env.ServerlessSpaceServer.UpdateSpaceCalls.Last()
	require.True(t, ok)
	assert.Equal(t, []string{"configuration.storage_limit"}, req.GetUpdateMask().GetPaths())
	storage := req.GetSpace().GetConfiguration().GetStorageLimit()
	assert.Equal(t, uint64(10*resource.GiB), storage.GetMaxSize(), "size is kept")
	assert.Equal(t, uint32(80), storage.GetAlertThresholdPercent(), "alert threshold is kept")
	assert.Equal(t, spacev1.LimitType_LIMIT_TYPE_SOFT, storage.GetType())
	assert.Equal(t, uint64(5_000_000), req.GetSpace().GetConfiguration().GetMonthlySpendLimit().GetMaxAmountMillicents(),
		"the spend limit is untouched")
}

func TestSpaceUpdate_RemoveLimit(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateWith(env, spaceWithLimits())

	_, _, err := testutil.Exec(t, env, "serverless", "space", "update", "space-abc", "--spend-limit", "unlimited")
	require.NoError(t, err)

	req, ok := env.ServerlessSpaceServer.UpdateSpaceCalls.Last()
	require.True(t, ok)
	assert.Equal(t, []string{"configuration.monthly_spend_limit"}, req.GetUpdateMask().GetPaths())
	assert.Nil(t, req.GetSpace().GetConfiguration().GetMonthlySpendLimit(), "an unset limit in the mask removes it")
	assert.NotNil(t, req.GetSpace().GetConfiguration().GetStorageLimit())
}

func TestSpaceUpdate_InvalidLimitFlags(t *testing.T) {
	tests := []struct {
		name  string
		space *spacev1.Space
		args  []string
	}{
		{name: "type without an existing limit", space: existingSpace(), args: []string{"--storage-limit-type", "soft"}},
		{name: "unlimited with type", space: spaceWithLimits(), args: []string{"--spend-limit", "unlimited", "--spend-limit-type", "soft"}},
		{name: "empty limit", space: spaceWithLimits(), args: []string{"--storage-limit", ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := testutil.NewTestEnv(t)
			setupUpdateWith(env, tt.space)

			args := append([]string{"serverless", "space", "update", "space-abc"}, tt.args...)
			_, _, err := testutil.Exec(t, env, args...)
			require.Error(t, err)
			assert.Equal(t, 0, env.ServerlessSpaceServer.UpdateSpaceCalls.Count())
		})
	}
}

func TestSpaceDescribe_Limits(t *testing.T) {
	env := testutil.NewTestEnv(t)
	sp := spaceWithLimits()
	sp.Configuration.MonthlySpendLimit.Type = spacev1.LimitType_LIMIT_TYPE_SOFT
	sp.State = &spacev1.SpaceState{
		Phase: spacev1.SpaceStatePhase_SPACE_STATE_PHASE_READY,
		Limits: &spacev1.SpaceLimitsStatus{
			AccessMode: spacev1.SpaceAccessMode_SPACE_ACCESS_MODE_WRITES_BLOCKED,
			Storage: &spacev1.StorageLimitStatus{
				UsedSize: uint64(10 * resource.GiB),
				Level:    spacev1.LimitUsageLevel_LIMIT_USAGE_LEVEL_LIMIT_REACHED,
			},
			MonthlySpend: &spacev1.SpendLimitStatus{
				UsedAmountMillicents: 4_012_345,
				Level:                spacev1.LimitUsageLevel_LIMIT_USAGE_LEVEL_THRESHOLD_REACHED,
				PeriodStart:          timestamppb.New(time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)),
				PeriodEnd:            timestamppb.New(time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)),
				Currency:             "USD",
			},
		},
	}
	env.ServerlessSpaceServer.GetSpaceCalls.Returns(&spacev1.GetSpaceResponse{Space: sp}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "describe", "space-abc")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Storage Limit:                   10GiB (hard, alert at 80%)")
	assert.Contains(t, stdout, "Monthly Spend Limit:             $50.00 (soft, alert at 80%)")
	assert.Contains(t, stdout, "Usage:")
	assert.Contains(t, stdout, "Access:                          WRITES_BLOCKED")
	assert.Contains(t, stdout, "Storage:                         10 GiB of 10GiB (100%), LIMIT_REACHED")
	assert.Contains(t, stdout, "Spend This Month:                $40.12 of $50.00 (80.2%), THRESHOLD_REACHED")
	assert.Contains(t, stdout, "Spend Period Ends:")
	assert.Contains(t, stdout, "2026-12-01")
}

func TestSpaceDescribe_UnlimitedWithoutUsage(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessSpaceServer.GetSpaceCalls.Returns(&spacev1.GetSpaceResponse{Space: existingSpace()}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "describe", "space-abc")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Storage Limit:                   unlimited")
	assert.Contains(t, stdout, "Monthly Spend Limit:             unlimited")
	assert.NotContains(t, stdout, "Usage:", "no usage measured yet")
}

func TestSpaceDescribe_LimitsJSON(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.ServerlessSpaceServer.GetSpaceCalls.Returns(&spacev1.GetSpaceResponse{Space: spaceWithLimits()}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "space", "describe", "space-abc", "--json")
	require.NoError(t, err)

	var result struct {
		Configuration struct {
			StorageLimit struct {
				MaxSize string `json:"maxSize"`
				Type    string `json:"type"`
			} `json:"storageLimit"`
			MonthlySpendLimit struct {
				MaxAmountMillicents string `json:"maxAmountMillicents"`
				Currency            string `json:"currency"`
			} `json:"monthlySpendLimit"`
		} `json:"configuration"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "10737418240", result.Configuration.StorageLimit.MaxSize)
	assert.Equal(t, "LIMIT_TYPE_HARD", result.Configuration.StorageLimit.Type)
	assert.Equal(t, "5000000", result.Configuration.MonthlySpendLimit.MaxAmountMillicents)
	assert.Equal(t, "USD", result.Configuration.MonthlySpendLimit.Currency)
}
