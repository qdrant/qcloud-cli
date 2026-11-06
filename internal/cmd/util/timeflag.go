package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
)

// errInvalidTimeFlag describes the formats accepted by ParseTimeFlag.
var errInvalidTimeFlag = errors.New("must be RFC3339, YYYY-MM-DD, or a duration ago such as 30m, 24h or 7d")

// AddTimeRangeFlags registers the --since and --until flags read by
// ReadPastTimeRange. defaultRange is the length of the range when --since is
// omitted, and is shown in the help text.
func AddTimeRangeFlags(cmd *cobra.Command, defaultRange time.Duration) {
	cmd.Flags().StringP("since", "s", "", "Start of the time range (RFC3339, YYYY-MM-DD, or a duration ago such as 24h or 7d; default: "+describeRange(defaultRange)+" before --until)")
	cmd.Flags().StringP("until", "u", "", "End of the time range (RFC3339, YYYY-MM-DD, or a duration ago such as 1h; default: now)")
}

// describeRange formats d in whole days when it is a multiple of a day, such
// as "7 days", and as a compact duration otherwise, such as "1h".
func describeRange(d time.Duration) string {
	const day = 24 * time.Hour
	switch {
	case d == day:
		return "1 day"
	case d > 0 && d%day == 0:
		return strconv.Itoa(int(d/day)) + " days"
	default:
		return output.CompactDuration(d)
	}
}

// ParseTimeFlag parses a time flag value used as a lower bound. It accepts an
// RFC3339 timestamp, a date in YYYY-MM-DD format (start of the day in UTC), or a
// positive duration that is subtracted from now: any Go duration (e.g. "90m",
// "24h") or a number of days with a "d" suffix (e.g. "7d").
func ParseTimeFlag(s string, now time.Time) (time.Time, error) {
	return parseTimeFlag(s, now, false)
}

// ParseUntilFlag is like ParseTimeFlag but for upper bounds: a YYYY-MM-DD date
// resolves to the end of that day, so that the whole day is included.
func ParseUntilFlag(s string, now time.Time) (time.Time, error) {
	return parseTimeFlag(s, now, true)
}

// ReadTimeRange parses the --since and --until flags of cmd with ParseTimeFlag
// and ParseUntilFlag. Unset flags yield nil so that the server defaults apply.
// It returns an error when both are set and --since is not before --until.
func ReadTimeRange(cmd *cobra.Command, now time.Time) (since, until *timestamppb.Timestamp, err error) {
	var sinceT, untilT time.Time
	if cmd.Flags().Changed("since") {
		v, _ := cmd.Flags().GetString("since")
		sinceT, err = ParseTimeFlag(v, now)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid --since %q: %w", v, err)
		}

		since = timestamppb.New(sinceT)
	}

	if cmd.Flags().Changed("until") {
		v, _ := cmd.Flags().GetString("until")
		untilT, err = ParseUntilFlag(v, now)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid --until %q: %w", v, err)
		}

		until = timestamppb.New(untilT)
	}

	if since != nil && until != nil && !sinceT.Before(untilT) {
		return nil, nil, errors.New("--since must be before --until")
	}

	return since, until, nil
}

// ReadPastTimeRange reads --since and --until like ReadTimeRange, for a range
// that lies in the past. --until defaults to now and is cut to now when it is in
// the future, and --since defaults to defaultRange before --until. It returns an
// error when --since is not in the past or not before --until.
func ReadPastTimeRange(cmd *cobra.Command, now time.Time, defaultRange time.Duration) (since, until time.Time, err error) {
	sinceTS, untilTS, err := ReadTimeRange(cmd, now)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	if sinceTS != nil && !sinceTS.AsTime().Before(now) {
		return time.Time{}, time.Time{}, errors.New("--since must be in the past")
	}

	until = now
	if untilTS != nil && untilTS.AsTime().Before(now) {
		until = untilTS.AsTime()
	}

	since = until.Add(-defaultRange)
	if sinceTS != nil {
		since = sinceTS.AsTime()
	}

	if !since.Before(until) {
		return time.Time{}, time.Time{}, errors.New("--since must be before --until")
	}

	return since, until, nil
}

func parseTimeFlag(s string, now time.Time, endOfDay bool) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}

	if endOfDay {
		if t, err := ParseDateEndOfDay(s); err == nil {
			return t, nil
		}
	} else if t, err := time.Parse(time.DateOnly, s); err == nil {
		return t, nil
	}

	d, err := parseAgo(s)
	if err != nil {
		return time.Time{}, errInvalidTimeFlag
	}

	return now.Add(-d), nil
}

// parseAgo parses a positive Go duration or a whole number of days ("7d").
func parseAgo(s string) (time.Duration, error) {
	if days, ok := strings.CutSuffix(s, "d"); ok {
		n, err := strconv.ParseUint(days, 10, 32)
		if err != nil || n == 0 {
			return 0, errInvalidTimeFlag
		}

		return time.Duration(n) * 24 * time.Hour, nil
	}

	d, err := time.ParseDuration(s)
	if err != nil || d <= 0 {
		return 0, errInvalidTimeFlag
	}

	return d, nil
}
