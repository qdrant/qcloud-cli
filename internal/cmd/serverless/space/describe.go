package space

import (
	"fmt"
	"io"

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
network restrictions, per-collection size limits and search-worker settings.
Limits prefixed with "platform" are derived from the account's quota and cannot
be changed directly.`,
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
				fmt.Fprintf(w, "REST Port: %d\n", ep.GetRestPort())
				fmt.Fprintf(w, "gRPC Port: %d\n", ep.GetGrpcPort())
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

// unlimitedIfZero returns "unlimited" for platform limits where 0 means no limit.
func unlimitedIfZero(formatted string, v uint64) string {
	if v == 0 {
		return "unlimited"
	}

	return formatted
}

// printSpaceConfiguration prints the configuration of a space as indented
// "key: value" lines.
func printSpaceConfiguration(w io.Writer, cfg *spacev1.SpaceConfiguration) {
	notSet := "(not set)"

	fmt.Fprintf(w, "  Allowed IPs:                     %s\n", output.JoinOrDefault(cfg.GetAllowedIpSourceRanges(), "(any)"))
	fmt.Fprintf(w, "  Allowed Origins:                 %s\n", output.JoinOrDefault(cfg.GetAllowedOrigins(), notSet))
	fmt.Fprintf(w, "  Max Collections (platform):      %s\n", unlimitedIfZero(fmt.Sprintf("%d", cfg.GetMaxCollectionsPerSpace()), cfg.GetMaxCollectionsPerSpace()))

	col := cfg.GetCollectionSettings()
	maxSize := notSet
	if col != nil && col.MaxSize != nil {
		maxSize = resource.ByteQuantity(col.GetMaxSize()).String()
	}

	fmt.Fprintf(w, "  Max Collection Size:             %s\n", maxSize)
	fmt.Fprintf(w, "  Max Collection Size (platform):  %s\n", unlimitedIfZero(resource.ByteQuantity(col.GetPlatformMaxSize()).String(), col.GetPlatformMaxSize()))

	searcher := cfg.GetSearcherSettings()
	idleTimeout := notSet
	if searcher != nil && searcher.IdleTimeout != nil {
		idleTimeout = searcher.GetIdleTimeout().AsDuration().String()
	}

	maxWorkers := notSet
	if searcher != nil && searcher.MaxWorkers != nil {
		maxWorkers = fmt.Sprintf("%d", searcher.GetMaxWorkers())
	}

	fmt.Fprintf(w, "  Searcher Idle Timeout:           %s\n", idleTimeout)
	fmt.Fprintf(w, "  Searcher Max Workers:            %s\n", maxWorkers)
	fmt.Fprintf(w, "  Searcher Max Workers (platform): %s\n", unlimitedIfZero(fmt.Sprintf("%d", searcher.GetPlatformMaxWorkers()), searcher.GetPlatformMaxWorkers()))
}
