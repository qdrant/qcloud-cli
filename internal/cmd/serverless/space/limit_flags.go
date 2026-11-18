package space

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	spacev1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/resource"
)

// limitUnlimited is the value of --storage-limit and --spend-limit that
// removes the limit.
const limitUnlimited = "unlimited"

// Bounds of the limit alert thresholds.
const (
	minLimitAlertPercent = 1
	maxLimitAlertPercent = 99
)

// addLimitFlags registers the storage and monthly spend limit flags shared by
// "space create" and "space update".
func addLimitFlags(cmd *cobra.Command) {
	cmd.Flags().String("storage-limit", "", "Maximum storage of the space (e.g. 10GiB, 512MiB), or 'unlimited' to remove the limit")
	cmd.Flags().String("storage-limit-type", "", "What happens at the storage limit: 'hard' rejects upserts, 'soft' only alerts (default: hard)")
	cmd.Flags().Uint32("storage-limit-alert-percent", 0, "Percentage of the storage limit at which an alert is sent (1 to 99, default: 80)")
	cmd.Flags().String("spend-limit", "", "Maximum monthly spend on searches and writes, in USD (e.g. 50 or 49.99), or 'unlimited' to remove the limit")
	cmd.Flags().String("spend-limit-type", "", "What happens at the spend limit: 'hard' rejects all requests until next month, 'soft' only alerts (default: hard)")
	cmd.Flags().Uint32("spend-limit-alert-percent", 0, "Percentage of the spend limit at which an alert is sent (1 to 99, default: 80)")
}

// applyLimitFlags applies the limit flags to cfg. It works for both create
// (empty configuration) and update (pre-cloned configuration): a type or
// alert threshold changes the existing limit, and needs --storage-limit or
// --spend-limit when the space has no such limit yet. Unset type and threshold
// of a new limit are left to the server defaults (hard, 80%).
func applyLimitFlags(cmd *cobra.Command, cfg *spacev1.SpaceConfiguration) error {
	storage, err := applyLimit(cmd, limitFlags{
		value: "storage-limit", limitType: "storage-limit-type", alertPercent: "storage-limit-alert-percent",
	}, cfg.GetStorageLimit() != nil)
	if err != nil {
		return err
	}

	switch {
	case storage.remove:
		cfg.StorageLimit = nil
	case storage.changed:
		if cfg.StorageLimit == nil {
			cfg.StorageLimit = &spacev1.StorageLimit{}
		}

		l := cfg.StorageLimit
		if storage.raw != "" {
			size, err := resource.ParseByteQuantity(storage.raw)
			if err != nil {
				return fmt.Errorf("invalid --storage-limit: %w", err)
			}

			if size <= 0 {
				return fmt.Errorf("invalid --storage-limit %q: must be greater than zero", storage.raw)
			}

			l.MaxSize = uint64(size)
		}

		if storage.limitType != nil {
			l.Type = *storage.limitType
		}

		if storage.alertPercent != nil {
			l.AlertThresholdPercent = storage.alertPercent
		}
	}

	spend, err := applyLimit(cmd, limitFlags{
		value: "spend-limit", limitType: "spend-limit-type", alertPercent: "spend-limit-alert-percent",
	}, cfg.GetMonthlySpendLimit() != nil)
	if err != nil {
		return err
	}

	switch {
	case spend.remove:
		cfg.MonthlySpendLimit = nil
	case spend.changed:
		if cfg.MonthlySpendLimit == nil {
			cfg.MonthlySpendLimit = &spacev1.SpendLimit{}
		}

		l := cfg.MonthlySpendLimit
		if spend.raw != "" {
			amount, err := util.ParseUSDMillicents(spend.raw)
			if err != nil {
				return fmt.Errorf("invalid --spend-limit: %w", err)
			}

			l.MaxAmountMillicents = amount
		}

		if spend.limitType != nil {
			l.Type = *spend.limitType
		}

		if spend.alertPercent != nil {
			l.AlertThresholdPercent = spend.alertPercent
		}
	}

	return nil
}

// limitFlags names the three flags of one limit.
type limitFlags struct {
	value, limitType, alertPercent string
}

// limitChange is what the flags of one limit ask for.
type limitChange struct {
	// changed is true when any of the limit's flags was set.
	changed bool
	// remove is true for the value 'unlimited'.
	remove bool
	// raw is the new limit value, empty when it is not changed.
	raw          string
	limitType    *spacev1.LimitType
	alertPercent *uint32
}

// applyLimit reads and validates the flags of one limit. exists reports
// whether the space already has the limit.
func applyLimit(cmd *cobra.Command, f limitFlags, exists bool) (limitChange, error) {
	var c limitChange
	flags := cmd.Flags()
	c.changed = flags.Changed(f.value) || flags.Changed(f.limitType) || flags.Changed(f.alertPercent)
	if !c.changed {
		return c, nil
	}

	if flags.Changed(f.value) {
		c.raw, _ = flags.GetString(f.value)
		c.raw = strings.TrimSpace(c.raw)
		if c.raw == "" {
			return c, fmt.Errorf("invalid --%s: must not be empty", f.value)
		}

		if strings.EqualFold(c.raw, limitUnlimited) {
			if flags.Changed(f.limitType) || flags.Changed(f.alertPercent) {
				return c, fmt.Errorf("--%s and --%s cannot be combined with --%s=%s", f.limitType, f.alertPercent, f.value, limitUnlimited)
			}

			c.remove = true
			return c, nil
		}
	} else if !exists {
		return c, fmt.Errorf("--%s and --%s need --%s when the space has no such limit", f.limitType, f.alertPercent, f.value)
	}

	if flags.Changed(f.limitType) {
		raw, _ := flags.GetString(f.limitType)
		t, err := parseLimitType(raw)
		if err != nil {
			return c, fmt.Errorf("invalid --%s: %w", f.limitType, err)
		}

		c.limitType = &t
	}

	if flags.Changed(f.alertPercent) {
		v, _ := flags.GetUint32(f.alertPercent)
		if v < minLimitAlertPercent || v > maxLimitAlertPercent {
			return c, fmt.Errorf("invalid --%s %d: must be between %d and %d", f.alertPercent, v, minLimitAlertPercent, maxLimitAlertPercent)
		}

		c.alertPercent = &v
	}

	return c, nil
}

// parseLimitType parses a limit type flag value.
func parseLimitType(s string) (spacev1.LimitType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "hard":
		return spacev1.LimitType_LIMIT_TYPE_HARD, nil
	case "soft":
		return spacev1.LimitType_LIMIT_TYPE_SOFT, nil
	default:
		return spacev1.LimitType_LIMIT_TYPE_UNSPECIFIED, fmt.Errorf("%q must be 'hard' or 'soft'", s)
	}
}
