package usage_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	meteringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/metering/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

const (
	spaceA = "0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60"
	spaceB = "7f1e2d3c-4b5a-4968-8776-655443322110"
)

func spaceUsage(id, name, item string, qty float64, unit string, amount int64) *meteringv1.UsageBreakdownSpace {
	return &meteringv1.UsageBreakdownSpace{
		SpaceId:            id,
		SpaceName:          name,
		StartTime:          timestamppb.New(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)),
		EndTime:            timestamppb.New(time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)),
		AmountMillicents:   amount,
		Currency:           "USD",
		BillableEntityType: item,
		Quantity:           qty,
		Unit:               unit,
	}
}

func sampleBreakdown() *meteringv1.GetUsageBreakdownResponse {
	discounted := spaceUsage(spaceB, "", "Serverless Search Requests", 1234567, "requests", 2_000_000)
	discounted.DiscountAmountMillicents = new(int64(500_000))
	discounted.DiscountAmountPercent = new(25.0)

	return &meteringv1.GetUsageBreakdownResponse{
		Items: []*meteringv1.UsageBreakdownMonth{
			{
				Year:  2026,
				Month: 10,
				Accounts: []*meteringv1.UsageBreakdownAccount{{
					AccountId:   "test-account-id",
					AccountName: "Acme",
					Clusters: []*meteringv1.UsageBreakdownCluster{{
						ClusterId:          "c0c0c0c0-0000-4000-8000-000000000000",
						ClusterName:        "prod-cluster",
						AmountMillicents:   99_900_000,
						Currency:           "USD",
						BillableEntityType: "Cluster",
					}},
					Spaces: []*meteringv1.UsageBreakdownSpace{
						spaceUsage(spaceA, "products", "Serverless Stored Data (GiB-month)", 12.34567, "GiB-month", 1_234_000),
						discounted,
					},
				}},
				Funding: &meteringv1.UsageBreakdownFunding{
					CreditsDrawn: []*meteringv1.CreditDrawdown{{CreditContractId: "contract-1", AmountMillicents: 1_000}},
				},
			},
			{
				// A month with cluster usage only is left out.
				Year:  2026,
				Month: 9,
				Accounts: []*meteringv1.UsageBreakdownAccount{{
					AccountId:   "test-account-id",
					AccountName: "Acme",
					Clusters: []*meteringv1.UsageBreakdownCluster{{
						ClusterId: "c0c0c0c0-0000-4000-8000-000000000000", ClusterName: "prod-cluster", Currency: "USD",
					}},
				}},
			},
		},
	}
}

func TestUsage_TextOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.MeteringServer.GetUsageBreakdownCalls.Returns(sampleBreakdown(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "usage", "--month", "2026-10")
	require.NoError(t, err)

	assert.Contains(t, stdout, "Period: 2026-10-01 00:00:00 UTC to 2026-11-01 00:00:00 UTC")
	for _, h := range []string{"MONTH", "SPACE", "SPACE ID", "ITEM", "QUANTITY", "AMOUNT", "DISCOUNT", "NET"} {
		assert.Contains(t, stdout, h)
	}

	assert.NotContains(t, stdout, "ACCOUNT", "a single account has no ACCOUNT column")
	assert.Contains(t, stdout, "2026-10")
	assert.Contains(t, stdout, "products")
	assert.Contains(t, stdout, spaceA)
	assert.Contains(t, stdout, "Serverless Stored Data (GiB-month)")
	assert.Contains(t, stdout, "12.3457 GiB-month")
	assert.Contains(t, stdout, "$12.34")
	assert.Contains(t, stdout, "1234567 requests")
	assert.Contains(t, stdout, "$20.00")
	assert.Contains(t, stdout, "$5.00 (25%)", "discount with percentage")
	assert.Contains(t, stdout, "$15.00", "net charge after discount")
	assert.Contains(t, stdout, "Total: $27.34")
	assert.NotContains(t, stdout, "prod-cluster", "cluster usage is left out")
	assert.NotContains(t, stdout, "2026-09", "months without space usage are left out")
}

func TestUsage_TextOutput_SortsRowsAndMarksUnknownSpaceName(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.MeteringServer.GetUsageBreakdownCalls.Returns(sampleBreakdown(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "usage", "--month", "2026-10")
	require.NoError(t, err)

	// The space without a name sorts first and shows "-".
	assert.Regexp(t, `(?m)^2026-10\s+-\s+`+spaceB, stdout)
	assert.Less(t, strings.Index(stdout, spaceB), strings.Index(stdout, spaceA))
}

func TestUsage_DefaultPeriodIsCurrentMonth(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.MeteringServer.GetUsageBreakdownCalls.Returns(&meteringv1.GetUsageBreakdownResponse{}, nil)

	before := time.Now().UTC()
	stdout, _, err := testutil.Exec(t, env, "serverless", "usage")
	require.NoError(t, err)
	assert.Contains(t, stdout, "No serverless usage in this period.")

	req, ok := env.MeteringServer.GetUsageBreakdownCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, time.Date(before.Year(), before.Month(), 1, 0, 0, 0, 0, time.UTC), req.GetStartTime().AsTime())
	assert.WithinRange(t, req.GetEndTime().AsTime(), before.Add(-time.Second), time.Now())
}

func TestUsage_MonthRequestsWholeMonth(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.MeteringServer.GetUsageBreakdownCalls.Returns(&meteringv1.GetUsageBreakdownResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "serverless", "usage", "--month", "2026-02")
	require.NoError(t, err)

	req, ok := env.MeteringServer.GetUsageBreakdownCalls.Last()
	require.True(t, ok)
	assert.Equal(t, time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), req.GetStartTime().AsTime())
	assert.Equal(t, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), req.GetEndTime().AsTime())
}

func TestUsage_CurrentMonthIsCutToNow(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.MeteringServer.GetUsageBreakdownCalls.Returns(&meteringv1.GetUsageBreakdownResponse{}, nil)

	before := time.Now().UTC()
	_, _, err := testutil.Exec(t, env, "serverless", "usage", "--month", before.Format("2006-01"))
	require.NoError(t, err)

	req, ok := env.MeteringServer.GetUsageBreakdownCalls.Last()
	require.True(t, ok)
	assert.WithinRange(t, req.GetEndTime().AsTime(), before.Add(-time.Second), time.Now())
}

func TestUsage_SinceAndUntil(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.MeteringServer.GetUsageBreakdownCalls.Returns(&meteringv1.GetUsageBreakdownResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "serverless", "usage", "--since", "2026-07-01", "--until", "2026-09-30T00:00:00Z")
	require.NoError(t, err)

	req, ok := env.MeteringServer.GetUsageBreakdownCalls.Last()
	require.True(t, ok)
	assert.Equal(t, time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), req.GetStartTime().AsTime())
	assert.Equal(t, time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC), req.GetEndTime().AsTime())
}

func TestUsage_OnlyUntilStartsAtItsMonth(t *testing.T) {
	for _, tc := range []struct {
		until     string
		wantStart time.Time
	}{
		{"2026-09-15T12:00:00Z", time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)},
		// The first instant of a month covers the month before it.
		{"2026-09-01T00:00:00Z", time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
	} {
		t.Run(tc.until, func(t *testing.T) {
			env := testutil.NewTestEnv(t)
			env.MeteringServer.GetUsageBreakdownCalls.Returns(&meteringv1.GetUsageBreakdownResponse{}, nil)

			_, _, err := testutil.Exec(t, env, "serverless", "usage", "--until", tc.until)
			require.NoError(t, err)

			req, ok := env.MeteringServer.GetUsageBreakdownCalls.Last()
			require.True(t, ok)
			assert.Equal(t, tc.wantStart, req.GetStartTime().AsTime())
		})
	}
}

func TestUsage_SpaceFilter(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.MeteringServer.GetUsageBreakdownCalls.Returns(sampleBreakdown(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "usage", "--month", "2026-10", "--space", spaceA)
	require.NoError(t, err)
	assert.Contains(t, stdout, spaceA)
	assert.NotContains(t, stdout, spaceB)
	assert.Contains(t, stdout, "Total: $12.34")
}

func TestUsage_SpaceFilterWithoutUsage(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.MeteringServer.GetUsageBreakdownCalls.Returns(sampleBreakdown(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "usage", "--month", "2026-10", "--space", "11111111-2222-4333-8444-555555555555")
	require.NoError(t, err)
	assert.Contains(t, stdout, "No serverless usage in this period.")
}

func TestUsage_ParentAccountShowsAccountColumn(t *testing.T) {
	env := testutil.NewTestEnv(t)
	resp := sampleBreakdown()
	resp.Items[0].Accounts = append(resp.Items[0].Accounts, &meteringv1.UsageBreakdownAccount{
		AccountId:   "child-account-id",
		AccountName: "Acme Child",
		Spaces: []*meteringv1.UsageBreakdownSpace{
			spaceUsage("22222222-3333-4444-8555-666666666666", "child-space", "Serverless Stored Data (GiB-month)", 1, "GiB-month", 100_000),
		},
	})
	env.MeteringServer.GetUsageBreakdownCalls.Returns(resp, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "usage", "--month", "2026-10")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ACCOUNT")
	assert.Contains(t, stdout, "Acme Child")
	assert.Contains(t, stdout, "child-space")
	assert.Contains(t, stdout, "Total: $28.34")
}

func TestUsage_TotalPerCurrency(t *testing.T) {
	env := testutil.NewTestEnv(t)
	eur := spaceUsage(spaceB, "eu-space", "Serverless Stored Data (GiB-month)", 1, "GiB-month", 300_000)
	eur.Currency = "EUR"
	env.MeteringServer.GetUsageBreakdownCalls.Returns(&meteringv1.GetUsageBreakdownResponse{
		Items: []*meteringv1.UsageBreakdownMonth{{
			Year: 2026, Month: 10,
			Accounts: []*meteringv1.UsageBreakdownAccount{{
				AccountId: "test-account-id", AccountName: "Acme",
				Spaces: []*meteringv1.UsageBreakdownSpace{
					spaceUsage(spaceA, "us-space", "Serverless Stored Data (GiB-month)", 1, "GiB-month", 100_000),
					eur,
				},
			}},
		}},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "usage", "--month", "2026-10")
	require.NoError(t, err)
	assert.Contains(t, stdout, "3.00 EUR")
	assert.Contains(t, stdout, "Total: 3.00 EUR, $1.00")
}

func TestUsage_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.MeteringServer.GetUsageBreakdownCalls.Returns(sampleBreakdown(), nil)

	stdout, _, err := testutil.Exec(t, env, "serverless", "usage", "--month", "2026-10", "--json")
	require.NoError(t, err)

	var result struct {
		Items []struct {
			Year     int             `json:"year"`
			Month    int             `json:"month"`
			Funding  json.RawMessage `json:"funding"`
			Accounts []struct {
				AccountID string            `json:"accountId"`
				Clusters  []json.RawMessage `json:"clusters"`
				Spaces    []struct {
					SpaceID                  string `json:"spaceId"`
					AmountMillicents         string `json:"amountMillicents"`
					DiscountAmountMillicents string `json:"discountAmountMillicents"`
					Unit                     string `json:"unit"`
				} `json:"spaces"`
			} `json:"accounts"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))

	require.Len(t, result.Items, 1, "months without space usage are left out")
	m := result.Items[0]
	assert.Equal(t, 2026, m.Year)
	assert.Equal(t, 10, m.Month)
	assert.Nil(t, m.Funding, "credits drawn are left out")
	require.Len(t, m.Accounts, 1)
	assert.Empty(t, m.Accounts[0].Clusters, "cluster usage is left out")
	require.Len(t, m.Accounts[0].Spaces, 2)
	assert.Equal(t, spaceA, m.Accounts[0].Spaces[0].SpaceID)
	assert.Equal(t, "1234000", m.Accounts[0].Spaces[0].AmountMillicents)
	assert.Equal(t, "GiB-month", m.Accounts[0].Spaces[0].Unit)
	assert.Equal(t, "500000", m.Accounts[0].Spaces[1].DiscountAmountMillicents)
}

func TestUsage_InputErrors(t *testing.T) {
	future := time.Now().UTC().AddDate(0, 2, 0).Format("2006-01")
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"invalid month", []string{"--month", "2026-13"}, "YYYY-MM"},
		{"month as date", []string{"--month", "2026-10-01"}, "YYYY-MM"},
		{"future month", []string{"--month", future}, "must not be in the future"},
		{"month with since", []string{"--month", "2026-10", "--since", "2026-10-01"}, "none of the others can be"},
		{"month with until", []string{"--month", "2026-10", "--until", "2026-10-31"}, "none of the others can be"},
		{"invalid since", []string{"--since", "yesterday"}, "invalid --since"},
		{"since after until", []string{"--since", "2026-10-02", "--until", "2026-10-01T00:00:00Z"}, "--since must be before --until"},
		{"since in the future", []string{"--since", time.Now().UTC().AddDate(0, 0, 2).Format(time.DateOnly)}, "--since must be in the past"},
		{"empty space", []string{"--space", ""}, "--space must not be empty"},
		{"positional arg", []string{"extra"}, "unknown command"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := testutil.NewTestEnv(t)

			_, _, err := testutil.Exec(t, env, append([]string{"serverless", "usage"}, tc.args...)...)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.want)
			assert.Equal(t, 0, env.MeteringServer.GetUsageBreakdownCalls.Count(), "no API call on invalid input")
		})
	}
}

func TestUsage_BackendError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.MeteringServer.GetUsageBreakdownCalls.Returns(nil, fmt.Errorf("permission denied"))

	_, _, err := testutil.Exec(t, env, "serverless", "usage")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get usage breakdown")
	assert.Contains(t, err.Error(), "permission denied")
}
