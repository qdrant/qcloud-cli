package space

import (
	"fmt"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/durationpb"

	spacev1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/resource"
)

// addSharedSpaceFlags registers the flags shared by "space create" and "space update".
func addSharedSpaceFlags(cmd *cobra.Command) {
	cmd.Flags().StringArray("label", nil, "Label ('key=value') to add/overwrite; append '-' to remove ('key-'), can be specified multiple times")
	cmd.Flags().String("cost-allocation-label", "", "Label for billing reports")
	cmd.Flags().StringArray("allowed-ip", nil, "Allowed client IP CIDR range (e.g. \"10.0.0.0/8\"); append '-' to remove; IPv4 only")
	cmd.Flags().StringArray("allowed-origin", nil, "Allowed browser origin for CORS (e.g. \"https://app.example.com\"); append '-' to remove; max 10")
	cmd.Flags().String("max-collection-size", "", "Maximum size per collection (e.g. 10GiB, 512MiB); must not exceed the platform limit")
	cmd.Flags().Duration("searcher-idle-timeout", 0, "Idle timeout after which search workers may be scaled down (1m to 15m)")
	cmd.Flags().Uint64("searcher-max-workers", 0, "Maximum number of search workers per collection")
}

// applySharedSpaceFlags applies all shared flags to space. It works for both
// create (empty space) and update (pre-cloned space with existing values).
// Labels, allowed IPs and allowed origins are merged with existing values so
// that removals ('-' suffix) work in both commands.
func applySharedSpaceFlags(cmd *cobra.Command, space *spacev1.Space) error {
	if cmd.Flags().Changed("label") {
		raw, _ := cmd.Flags().GetStringArray("label")
		changes, err := util.ParseLabels(raw)
		if err != nil {
			return err
		}

		space.Labels = util.ApplyLabels(space.Labels, changes)
	}

	if cmd.Flags().Changed("cost-allocation-label") {
		v, _ := cmd.Flags().GetString("cost-allocation-label")
		space.CostAllocationLabel = &v
	}

	if space.Configuration == nil {
		space.Configuration = &spacev1.SpaceConfiguration{}
	}

	cfg := space.Configuration

	if cmd.Flags().Changed("allowed-ip") {
		raw, _ := cmd.Flags().GetStringArray("allowed-ip")
		changes, err := util.ParseIPs(raw)
		if err != nil {
			return err
		}

		cfg.AllowedIpSourceRanges = util.ApplyIPs(cfg.AllowedIpSourceRanges, changes)
	}

	if cmd.Flags().Changed("allowed-origin") {
		raw, _ := cmd.Flags().GetStringArray("allowed-origin")
		changes, err := util.ParseListChanges("--allowed-origin", "origin", raw)
		if err != nil {
			return err
		}

		cfg.AllowedOrigins = util.ApplyListChanges(cfg.AllowedOrigins, changes)
	}

	if cmd.Flags().Changed("max-collection-size") {
		raw, _ := cmd.Flags().GetString("max-collection-size")
		size, err := resource.ParseByteQuantity(raw)
		if err != nil {
			return fmt.Errorf("invalid --max-collection-size: %w", err)
		}

		if size <= 0 {
			return fmt.Errorf("invalid --max-collection-size %q: must be greater than zero", raw)
		}

		if cfg.CollectionSettings == nil {
			cfg.CollectionSettings = &spacev1.CollectionSettings{}
		}

		cfg.CollectionSettings.MaxSize = new(uint64(size))
	}

	if cmd.Flags().Changed("searcher-idle-timeout") {
		v, _ := cmd.Flags().GetDuration("searcher-idle-timeout")
		if cfg.SearcherSettings == nil {
			cfg.SearcherSettings = &spacev1.SearcherSettings{}
		}

		cfg.SearcherSettings.IdleTimeout = durationpb.New(v)
	}

	if cmd.Flags().Changed("searcher-max-workers") {
		v, _ := cmd.Flags().GetUint64("searcher-max-workers")
		if cfg.SearcherSettings == nil {
			cfg.SearcherSettings = &spacev1.SearcherSettings{}
		}

		cfg.SearcherSettings.MaxWorkers = &v
	}

	return nil
}

// spaceResultMessage returns the result line of a create command: the endpoint
// when the space is already ready (after --wait), or a plain creation message.
func spaceResultMessage(space *spacev1.Space, action string) string {
	if ep := space.GetState().GetEndpoint(); ep != nil && ep.GetUrl() != "" &&
		space.GetState().GetPhase() == spacev1.SpaceStatePhase_SPACE_STATE_PHASE_READY {
		return fmt.Sprintf("Space %s (%s) is ready. Endpoint: %s\n", space.GetId(), space.GetName(), ep.GetUrl())
	}

	return fmt.Sprintf("Space %s (%s) %s.\n", space.GetId(), space.GetName(), action)
}
