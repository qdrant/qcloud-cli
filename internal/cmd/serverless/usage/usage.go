package usage

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"

	meteringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/metering/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

// monthLayout is the format of the --month flag and of the MONTH column.
const monthLayout = "2006-01"

// NewCommand creates the serverless usage command.
func NewCommand(s *state.State) *cobra.Command {
	// The period is resolved in Fetch and shown by PrintText.
	var start, end time.Time

	cmd := base.DescribeCmd[*meteringv1.GetUsageBreakdownResponse]{
		Use:   "usage",
		Short: "Show the billed serverless usage per space",
		Long: `Show the billed serverless usage of the current account per space and billable item.

Each row is one billable item of one space in one calendar month, as named on
the invoice (for example "Serverless Updated Data (GiB)"), with the quantity
consumed, the charge, the discount and the net charge. The net total of all
rows is shown below the table. Prepaid credits drawn for the period are not
subtracted.

By default, the period is the current calendar month (UTC) up to now. Use
--month for a whole calendar month, or --since and --until for any other
period. When only --until is given, the period starts on the first day of its
month. Usage is reported per calendar month, so a period that starts or ends
within a month only includes the usage metered in that part of the month.

For a parent billing account, the usage of all child accounts is included and
an ACCOUNT column is shown. Cluster usage and the prepaid credits drawn are
left out, also from --json. The API key needs the read:payment_information
permission.`,
		Example: `# Show the serverless usage of the current month so far
qcloud serverless usage

# Show the usage of October 2026
qcloud serverless usage --month 2026-10

# Show the usage of one space during the third quarter of 2026
qcloud serverless usage --space 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --since 2026-07-01 --until 2026-09-30

# Export the usage as JSON
qcloud serverless usage --month 2026-10 --json`,
		Args: cobra.NoArgs,
		Fetch: func(s *state.State, cmd *cobra.Command, _ []string) (*meteringv1.GetUsageBreakdownResponse, error) {
			var err error
			start, end, err = readPeriod(cmd, time.Now().UTC())
			if err != nil {
				return nil, err
			}

			var spaceID string
			if cmd.Flags().Changed("space") {
				spaceID, _ = cmd.Flags().GetString("space")
				if spaceID == "" {
					return nil, errors.New("--space must not be empty")
				}
			}

			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			resp, err := client.Metering().GetUsageBreakdown(ctx, &meteringv1.GetUsageBreakdownRequest{
				AccountId: accountID,
				StartTime: timestamppb.New(start),
				EndTime:   timestamppb.New(end),
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get usage breakdown: %w", err)
			}

			return spacesOnly(resp, spaceID), nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, resp *meteringv1.GetUsageBreakdownResponse) error {
			fmt.Fprintf(w, "Period: %s to %s\n", output.FullDateTime(start), output.FullDateTime(end))

			rows := flatten(resp)
			if len(rows) == 0 {
				fmt.Fprintln(w, "No serverless usage in this period.")
				return nil
			}

			fmt.Fprintln(w)

			multiAccount := slices.ContainsFunc(rows, func(r row) bool { return r.accountID != rows[0].accountID })

			t := output.NewTable[row](w)
			t.AddField("MONTH", func(r row) string { return r.month })
			if multiAccount {
				t.AddField("ACCOUNT", func(r row) string { return r.accountName })
			}

			t.AddField("SPACE", func(r row) string { return cmp.Or(r.space.GetSpaceName(), "-") })
			t.AddField("SPACE ID", func(r row) string { return r.space.GetSpaceId() })
			t.AddField("ITEM", func(r row) string { return r.space.GetBillableEntityType() })
			t.AddField("QUANTITY", func(r row) string {
				return strings.TrimSpace(quantity(r.space.GetQuantity()) + " " + r.space.GetUnit())
			})
			t.AddField("AMOUNT", func(r row) string {
				return output.Millicents(r.space.GetAmountMillicents(), r.space.GetCurrency())
			})
			t.AddField("DISCOUNT", func(r row) string { return discount(r.space) })
			t.AddField("NET", func(r row) string { return output.Millicents(net(r.space), r.space.GetCurrency()) })
			t.SetItems(rows)
			t.Render()

			fmt.Fprintf(w, "\nTotal: %s\n", total(rows))
			return nil
		},
	}.CobraCommand(s)

	util.AddTimeRangeFlags(cmd, 0)
	cmd.Flags().Lookup("since").Usage = "Start of the period (RFC3339, YYYY-MM-DD, or a duration ago such as 24h or 7d; default: first day of the month of --until)"
	cmd.Flags().Lookup("until").Usage = "End of the period, exclusive (RFC3339, YYYY-MM-DD for the end of that day, or a duration ago such as 1h; default: now)"
	cmd.Flags().String("month", "", "Calendar month in YYYY-MM format (UTC); cannot be combined with --since or --until")
	cmd.Flags().String("space", "", "Only include the space with this ID")
	_ = cmd.RegisterFlagCompletionFunc("space", completion.SpaceIDCompletion(s))
	cmd.MarkFlagsMutuallyExclusive("month", "since")
	cmd.MarkFlagsMutuallyExclusive("month", "until")

	return cmd
}

// readPeriod resolves the period from --month, or from --since and --until.
// The end is cut to now, because usage is only metered up to now.
func readPeriod(cmd *cobra.Command, now time.Time) (time.Time, time.Time, error) {
	if cmd.Flags().Changed("month") {
		v, _ := cmd.Flags().GetString("month")
		start, err := time.Parse(monthLayout, v)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid --month %q: must be in YYYY-MM format", v)
		}

		if !start.Before(now) {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid --month %q: must not be in the future", v)
		}

		end := start.AddDate(0, 1, 0)
		if end.After(now) {
			end = now
		}

		return start, end, nil
	}

	since, until, err := util.ReadTimeRange(cmd, now)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	end := now
	if until != nil && until.AsTime().Before(now) {
		end = until.AsTime()
	}

	var start time.Time
	if since != nil {
		start = since.AsTime()
	} else {
		start = monthStart(end)
		if !start.Before(end) {
			// --until is the first instant of a month: cover the month before it.
			start = start.AddDate(0, -1, 0)
		}
	}

	if !start.Before(now) {
		return time.Time{}, time.Time{}, errors.New("--since must be in the past")
	}

	if !start.Before(end) {
		return time.Time{}, time.Time{}, errors.New("--since must be before --until")
	}

	return start, end, nil
}

// monthStart returns the first instant of the UTC calendar month of t.
func monthStart(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// spacesOnly returns a copy of resp with only the serverless space usage,
// optionally of a single space. Cluster usage and the credits drawn, which
// cover all usage of the account, are left out, as are the accounts and
// months without any space usage left.
func spacesOnly(resp *meteringv1.GetUsageBreakdownResponse, spaceID string) *meteringv1.GetUsageBreakdownResponse {
	out := &meteringv1.GetUsageBreakdownResponse{}
	for _, m := range resp.GetItems() {
		month := &meteringv1.UsageBreakdownMonth{Year: m.GetYear(), Month: m.GetMonth()}
		for _, a := range m.GetAccounts() {
			var spaces []*meteringv1.UsageBreakdownSpace
			for _, sp := range a.GetSpaces() {
				if spaceID == "" || sp.GetSpaceId() == spaceID {
					spaces = append(spaces, sp)
				}
			}

			if len(spaces) > 0 {
				month.Accounts = append(month.Accounts, &meteringv1.UsageBreakdownAccount{
					AccountId:   a.GetAccountId(),
					AccountName: a.GetAccountName(),
					Spaces:      spaces,
				})
			}
		}

		if len(month.GetAccounts()) > 0 {
			out.Items = append(out.Items, month)
		}
	}

	return out
}

// row is one table row: the usage of one billable item of one space in a month.
type row struct {
	month       string
	accountID   string
	accountName string
	space       *meteringv1.UsageBreakdownSpace
}

// flatten turns the response into rows sorted by month, account, space name,
// space ID and billable item.
func flatten(resp *meteringv1.GetUsageBreakdownResponse) []row {
	var rows []row
	for _, m := range resp.GetItems() {
		month := fmt.Sprintf("%04d-%02d", m.GetYear(), m.GetMonth())
		for _, a := range m.GetAccounts() {
			for _, sp := range a.GetSpaces() {
				rows = append(rows, row{month: month, accountID: a.GetAccountId(), accountName: a.GetAccountName(), space: sp})
			}
		}
	}

	slices.SortStableFunc(rows, func(a, b row) int {
		return cmp.Or(
			cmp.Compare(a.month, b.month),
			cmp.Compare(a.accountName, b.accountName),
			cmp.Compare(a.accountID, b.accountID),
			cmp.Compare(a.space.GetSpaceName(), b.space.GetSpaceName()),
			cmp.Compare(a.space.GetSpaceId(), b.space.GetSpaceId()),
			cmp.Compare(a.space.GetBillableEntityType(), b.space.GetBillableEntityType()),
		)
	})

	return rows
}

// net returns the charge of sp after its discount.
func net(sp *meteringv1.UsageBreakdownSpace) int64 {
	return sp.GetAmountMillicents() - sp.GetDiscountAmountMillicents()
}

// discount formats the discount of sp, with its percentage when known, or "-"
// when there is none.
func discount(sp *meteringv1.UsageBreakdownSpace) string {
	if sp.DiscountAmountMillicents == nil && sp.DiscountAmountPercent == nil {
		return "-"
	}

	s := output.Millicents(sp.GetDiscountAmountMillicents(), sp.GetCurrency())
	if sp.DiscountAmountPercent != nil {
		s += " (" + strconv.FormatFloat(sp.GetDiscountAmountPercent(), 'f', -1, 64) + "%)"
	}

	return s
}

// total formats the sum of the net charges of rows, one amount per currency.
func total(rows []row) string {
	sums := map[string]int64{}
	var currencies []string
	for _, r := range rows {
		c := r.space.GetCurrency()
		if _, ok := sums[c]; !ok {
			currencies = append(currencies, c)
		}

		sums[c] += net(r.space)
	}

	slices.Sort(currencies)
	parts := make([]string, 0, len(currencies))
	for _, c := range currencies {
		parts = append(parts, output.Millicents(sums[c], c))
	}

	return strings.Join(parts, ", ")
}

// quantity formats a metered quantity with at most four decimals and no
// trailing zeros (e.g. "12.5", "1234567", "0.0001").
func quantity(v float64) string {
	return strconv.FormatFloat(math.Round(v*1e4)/1e4, 'f', -1, 64)
}
