package util

import (
	"time"

	"github.com/robfig/cron/v3"
)

// cronParser accepts standard 5-field expressions, 6-field expressions with a
// leading seconds field, and descriptors such as "@daily" or "@every 1h".
var cronParser = cron.NewParser(
	cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
)

// NextCronRun returns the next time after now (UTC) at which the cron
// expression fires. It returns false if the expression cannot be parsed.
func NextCronRun(cronExpr string) (time.Time, bool) {
	s, err := cronParser.Parse(cronExpr)
	if err != nil {
		return time.Time{}, false
	}

	return s.Next(time.Now().UTC()), true
}
