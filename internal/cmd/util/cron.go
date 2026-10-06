package util

import (
	"time"

	"github.com/robfig/cron/v3"
)

// NextCronRun returns the next time after now (UTC) at which the standard cron
// expression fires. It returns false if the expression cannot be parsed.
func NextCronRun(cronExpr string) (time.Time, bool) {
	s, err := cron.ParseStandard(cronExpr)
	if err != nil {
		return time.Time{}, false
	}

	return s.Next(time.Now().UTC()), true
}
