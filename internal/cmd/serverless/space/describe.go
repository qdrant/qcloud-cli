package space

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	spacev1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/resource"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newDescribeCommand(s *state.State) *cobra.Command {
	return base.DescribeCmd[*spacev1.Space]{
		Use:   "describe <space-id>",
		Short: "Describe a space",
		Long: `Describe a serverless space.

Shows the space's phase, region and endpoint together with its configuration:
network restrictions, per-collection size limits, search-worker settings, and
the storage and monthly spend limits of the space. Limits prefixed with
"platform" are derived from the account's quota and cannot be changed directly.

When usage has been measured, the Usage section shows storage and this month's
spend against each limit, and which operations the space accepts: FULL,
WRITES_BLOCKED (a hard storage limit is reached: upserts are rejected, searches
and deletes work) or BLOCKED (a hard spend limit is reached: all requests are
rejected until the next month). Usage is refreshed about every 5 minutes.`,
		Example: `# Describe a space
qcloud serverless space describe 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Output as JSON
qcloud serverless space describe 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --json`,
		Args: util.ExactArgs(1, "a space ID"),
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*spacev1.Space, error) {
			return getSpace(s, cmd, args[0])
		},
		PrintText: func(_ *cobra.Command, w io.Writer, space *spacev1.Space) error {
			fmt.Fprintf(w, "ID:        %s\n", space.GetId())
			fmt.Fprintf(w, "Name:      %s\n", space.GetName())
			if st := space.GetState(); st != nil {
				fmt.Fprintf(w, "Phase:     %s\n", output.SpacePhase(st.GetPhase()))
				if st.GetReason() != "" {
					fmt.Fprintf(w, "Reason:    %s\n", st.GetReason())
				}
			}

			fmt.Fprintf(w, "Region:    %s\n", space.GetCloudRegionId())

			if ep := space.GetState().GetEndpoint(); ep != nil && ep.GetUrl() != "" {
				fmt.Fprintf(w, "Endpoint:  %s\n", ep.GetUrl())
			}

			if space.GetCreatedAt() != nil {
				t := space.GetCreatedAt().AsTime()
				fmt.Fprintf(w, "Created:   %s  (%s)\n", output.HumanTime(t), output.FullDateTime(t))
			}

			if space.CostAllocationLabel != nil {
				fmt.Fprintf(w, "Cost Allocation Label: %s\n", space.GetCostAllocationLabel())
			}

			if labels := space.GetLabels(); len(labels) > 0 {
				fmt.Fprintf(w, "Labels:    ")
				for i, kv := range labels {
					if i > 0 {
						fmt.Fprintf(w, "           ")
					}

					fmt.Fprintf(w, "%s=%s\n", kv.GetKey(), kv.GetValue())
				}
			}

			cfg := space.GetConfiguration()
			if cfg == nil {
				return nil
			}

			fmt.Fprintln(w)
			fmt.Fprintln(w, "Configuration:")
			printSpaceConfiguration(w, cfg)

			if limits := space.GetState().GetLimits(); limits != nil {
				fmt.Fprintln(w)
				fmt.Fprintln(w, "Usage:")
				printSpaceLimitsStatus(w, cfg, limits)
			}

			return nil
		},
		ValidArgsFunction: completion.SpaceIDCompletion(s),
	}.CobraCommand(s)
}

// getSpace fetches a single space by ID.
func getSpace(s *state.State, cmd *cobra.Command, spaceID string) (*spacev1.Space, error) {
	ctx := cmd.Context()
	client, err := s.Client(ctx)
	if err != nil {
		return nil, err
	}

	accountID, err := s.AccountID()
	if err != nil {
		return nil, err
	}

	resp, err := client.ServerlessSpace().GetSpace(ctx, &spacev1.GetSpaceRequest{
		AccountId: accountID,
		SpaceId:   spaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get space: %w", err)
	}

	return resp.GetSpace(), nil
}

// printSpaceConfiguration prints the configuration of a space as indented
// "key: value" lines.
func printSpaceConfiguration(w io.Writer, cfg *spacev1.SpaceConfiguration) {
	notSet := "(not set)"

	fmt.Fprintf(w, "  Allowed IPs:                     %s\n", output.JoinOrDefault(cfg.GetAllowedIpSourceRanges(), "(any)"))
	fmt.Fprintf(w, "  Allowed Origins:                 %s\n", output.JoinOrDefault(cfg.GetAllowedOrigins(), notSet))
	fmt.Fprintf(w, "  Max Collections (platform):      %s\n", output.UnlimitedIfZero(fmt.Sprintf("%d", cfg.GetMaxCollectionsPerSpace()), cfg.GetMaxCollectionsPerSpace()))

	col := cfg.GetCollectionSettings()
	maxSize := notSet
	if col != nil && col.MaxSize != nil {
		maxSize = resource.ByteQuantity(col.GetMaxSize()).String()
	}

	fmt.Fprintf(w, "  Max Collection Size:             %s\n", maxSize)
	platformMaxSize := notSet
	if col != nil {
		platformMaxSize = output.UnlimitedIfZero(resource.ByteQuantity(col.GetPlatformMaxSize()).String(), col.GetPlatformMaxSize())
	}

	fmt.Fprintf(w, "  Max Collection Size (platform):  %s\n", platformMaxSize)

	maxIndexedSize := notSet
	if col != nil && col.MaxIndexedSize != nil {
		maxIndexedSize = resource.ByteQuantity(col.GetMaxIndexedSize()).String()
	}

	fmt.Fprintf(w, "  Max Indexed Size:                %s\n", maxIndexedSize)
	platformMaxIndexedSize := notSet
	if col != nil {
		platformMaxIndexedSize = output.UnlimitedIfZero(resource.ByteQuantity(col.GetPlatformMaxIndexedSize()).String(), col.GetPlatformMaxIndexedSize())
	}

	fmt.Fprintf(w, "  Max Indexed Size (platform):     %s\n", platformMaxIndexedSize)
	sizeWarning := notSet
	if col != nil && col.SizeWarningThresholdPercent != nil {
		sizeWarning = fmt.Sprintf("%d%%", col.GetSizeWarningThresholdPercent())
	}

	fmt.Fprintf(w, "  Size Warning Threshold:          %s\n", sizeWarning)

	searcher := cfg.GetSearcherSettings()
	minIdleTimeout := notSet
	if searcher != nil && searcher.MinIdleTimeout != nil {
		minIdleTimeout = searcher.GetMinIdleTimeout().AsDuration().String()
	}

	maxWorkers := notSet
	if searcher != nil && searcher.MaxWorkers != nil {
		maxWorkers = fmt.Sprintf("%d", searcher.GetMaxWorkers())
	}

	fmt.Fprintf(w, "  Searcher Min Idle Timeout:       %s\n", minIdleTimeout)
	fmt.Fprintf(w, "  Searcher Max Workers:            %s\n", maxWorkers)
	platformMaxWorkers := notSet
	if searcher != nil {
		platformMaxWorkers = output.UnlimitedIfZero(fmt.Sprintf("%d", searcher.GetPlatformMaxWorkers()), searcher.GetPlatformMaxWorkers())
	}

	fmt.Fprintf(w, "  Searcher Max Workers (platform): %s\n", platformMaxWorkers)
	fmt.Fprintf(w, "  Storage Limit:                   %s\n", storageLimitString(cfg.GetStorageLimit()))
	fmt.Fprintf(w, "  Monthly Spend Limit:             %s\n", spendLimitString(cfg.GetMonthlySpendLimit()))
}

// storageLimitString formats a storage limit, e.g. "10GiB (hard, alert at 80%)".
func storageLimitString(l *spacev1.StorageLimit) string {
	if l == nil {
		return "unlimited"
	}

	return fmt.Sprintf("%s (%s)", resource.ByteQuantity(l.GetMaxSize()).String(), limitTypeAndAlert(l.GetType(), l.AlertThresholdPercent)) //nolint:gosec // sizes fit in int64
}

// spendLimitString formats a monthly spend limit, e.g. "$50.00 (hard, alert at 80%)".
func spendLimitString(l *spacev1.SpendLimit) string {
	if l == nil {
		return "unlimited"
	}

	return fmt.Sprintf("%s (%s)", output.USDMillicents(l.GetMaxAmountMillicents()), limitTypeAndAlert(l.GetType(), l.AlertThresholdPercent))
}

// limitTypeAndAlert formats the type and alert threshold of a limit, e.g.
// "hard, alert at 80%".
func limitTypeAndAlert(t spacev1.LimitType, alertPercent *uint32) string {
	s := strings.ToLower(output.LimitType(t))
	if alertPercent != nil {
		s += fmt.Sprintf(", alert at %d%%", *alertPercent)
	}

	return s
}

// printSpaceLimitsStatus prints the usage of a space against its storage and
// monthly spend limits as indented "key: value" lines.
func printSpaceLimitsStatus(w io.Writer, cfg *spacev1.SpaceConfiguration, st *spacev1.SpaceLimitsStatus) {
	fmt.Fprintf(w, "  Access:                          %s\n", output.SpaceAccessMode(st.GetAccessMode()))

	if storage := st.GetStorage(); storage != nil {
		used := output.Bytes(storage.GetUsedSize())
		if l := cfg.GetStorageLimit(); l != nil && l.GetMaxSize() > 0 {
			used = fmt.Sprintf("%s of %s (%s)", used, resource.ByteQuantity(l.GetMaxSize()).String(), //nolint:gosec // sizes fit in int64
				output.Percent(output.PercentOf(float64(storage.GetUsedSize()), float64(l.GetMaxSize()))))
		}

		fmt.Fprintf(w, "  Storage:                         %s, %s\n", used, output.LimitUsageLevel(storage.GetLevel()))
	}

	if spend := st.GetMonthlySpend(); spend != nil {
		used := output.USDMillicents(spend.GetUsedAmountMillicents())
		if l := cfg.GetMonthlySpendLimit(); l != nil && l.GetMaxAmountMillicents() > 0 {
			used = fmt.Sprintf("%s of %s (%s)", used, output.USDMillicents(l.GetMaxAmountMillicents()),
				output.Percent(output.PercentOf(float64(spend.GetUsedAmountMillicents()), float64(l.GetMaxAmountMillicents()))))
		}

		fmt.Fprintf(w, "  Spend This Month:                %s, %s\n", used, output.LimitUsageLevel(spend.GetLevel()))
		if end := spend.GetPeriodEnd(); end != nil {
			fmt.Fprintf(w, "  Spend Period Ends:               %s\n", output.FullDateTime(end.AsTime()))
		}
	}
}
