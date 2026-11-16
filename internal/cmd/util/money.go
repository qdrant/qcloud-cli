package util

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// millicentsPerUSD is the number of millicents in one US dollar: a millicent
// is 1/1000 of a cent, so 1 USD = 100 cents = 100,000 millicents.
const millicentsPerUSD = 100_000

// ParseUSDMillicents parses a dollar amount such as "50", "49.99" or "$50"
// and returns it in millicents (1/1000 of a US cent). At most two decimals
// (whole cents) are accepted, and the amount must be greater than zero.
func ParseUSDMillicents(s string) (uint64, error) {
	raw := strings.TrimPrefix(strings.TrimSpace(s), "$")
	whole, frac, hasFrac := strings.Cut(raw, ".")
	if whole == "" || (hasFrac && (frac == "" || len(frac) > 2)) {
		return 0, fmt.Errorf("cannot parse %q as a dollar amount (e.g. 50 or 49.99)", s)
	}

	dollars, err := strconv.ParseUint(whole, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot parse %q as a dollar amount (e.g. 50 or 49.99)", s)
	}

	var cents uint64
	if hasFrac {
		frac += strings.Repeat("0", 2-len(frac))
		cents, err = strconv.ParseUint(frac, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot parse %q as a dollar amount (e.g. 50 or 49.99)", s)
		}
	}

	if dollars > (math.MaxUint64-cents*1000)/millicentsPerUSD {
		return 0, fmt.Errorf("dollar amount %q is too large", s)
	}

	v := dollars*millicentsPerUSD + cents*1000
	if v == 0 {
		return 0, fmt.Errorf("dollar amount %q must be greater than zero", s)
	}

	return v, nil
}
