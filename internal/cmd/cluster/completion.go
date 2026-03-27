package cluster

import (
	"fmt"

	"github.com/spf13/cobra"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/resource"
	"github.com/qdrant/qcloud-cli/internal/state"
)

// packageFilter holds the parameters for filtering packages.
// Zero values for CPU/RAM/GPU mean "do not filter on that dimension".
type packageFilter struct {
	CPU        resource.Millicores
	RAM        resource.ByteQuantity
	GPU        resource.Millicores
	IncludeGPU bool
	MultiAz    bool
}

// filteredPackages fetches active packages matching the given filter.
// Returns nil (no completions) if --cloud-provider is not set.
func filteredPackages(cmd *cobra.Command, s *state.State, f packageFilter) ([]*bookingv1.Package, error) {
	provider, _ := cmd.Flags().GetString("cloud-provider")
	if provider == "" {
		return nil, nil
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

	region, _ := cmd.Flags().GetString("cloud-region")
	req := &bookingv1.ListPackagesRequest{
		AccountId:       accountID,
		CloudProviderId: provider,
		Statuses:        []bookingv1.PackageStatus{bookingv1.PackageStatus_PACKAGE_STATUS_ACTIVE},
		Gpu:             &f.IncludeGPU,
	}
	if region != "" {
		req.CloudProviderRegionId = &region
	}
	if f.MultiAz {
		req.MultiAz = new(true)
	}

	resp, err := client.Booking().ListPackages(ctx, req)
	if err != nil {
		return nil, err
	}

	var result []*bookingv1.Package
	for _, p := range resp.GetItems() {
		rc := p.GetResourceConfiguration()
		if f.CPU != 0 {
			pkgCPU, _ := resource.ParseMillicores(rc.GetCpu())
			if pkgCPU != f.CPU {
				continue
			}
		}
		if f.RAM != 0 {
			pkgRAM, _ := resource.ParseByteQuantity(rc.GetRam())
			if pkgRAM != f.RAM {
				continue
			}
		}
		if f.GPU != 0 {
			pkgGPU, _ := resource.ParseMillicores(rc.GetGpu())
			if pkgGPU != f.GPU {
				continue
			}
		}
		result = append(result, p)
	}
	return result, nil
}

// cpuCompletion returns a completion function for the --cpu flag.
func cpuCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		provider, _ := cmd.Flags().GetString("cloud-provider")
		if provider == "" {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		ram := *cmd.Flags().Lookup("ram").Value.(*resource.ByteQuantity)
		gpu := *cmd.Flags().Lookup("gpu").Value.(*resource.Millicores)
		multiAz, _ := cmd.Flags().GetBool("multi-az")
		pkgs, err := filteredPackages(cmd, s, packageFilter{
			RAM:        ram,
			GPU:        gpu,
			IncludeGPU: gpu != 0,
			MultiAz:    multiAz,
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		seen := make(map[resource.Millicores]struct{})
		var completions []string
		for _, p := range pkgs {
			v, err := resource.ParseMillicores(p.GetResourceConfiguration().GetCpu())
			if err != nil {
				cobra.CompErrorln(fmt.Sprintf("package %s: %v", p.GetName(), err))
				continue
			}
			if _, ok := seen[v]; !ok {
				seen[v] = struct{}{}
				completions = append(completions, v.String())
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// ramCompletion returns a completion function for the --ram flag.
func ramCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		provider, _ := cmd.Flags().GetString("cloud-provider")
		if provider == "" {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		cpu := *cmd.Flags().Lookup("cpu").Value.(*resource.Millicores)
		gpu := *cmd.Flags().Lookup("gpu").Value.(*resource.Millicores)
		multiAz, _ := cmd.Flags().GetBool("multi-az")
		pkgs, err := filteredPackages(cmd, s, packageFilter{
			CPU:        cpu,
			GPU:        gpu,
			IncludeGPU: gpu != 0,
			MultiAz:    multiAz,
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		seen := make(map[resource.ByteQuantity]struct{})
		var completions []string
		for _, p := range pkgs {
			v, err := resource.ParseByteQuantity(p.GetResourceConfiguration().GetRam())
			if err != nil {
				cobra.CompErrorln(fmt.Sprintf("package %s: %v", p.GetName(), err))
				continue
			}
			if _, ok := seen[v]; !ok {
				seen[v] = struct{}{}
				completions = append(completions, resource.FormatByteQuantity(v, resource.UnitGiB))
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// diskCompletion returns a completion function for the --disk flag.
func diskCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		provider, _ := cmd.Flags().GetString("cloud-provider")
		if provider == "" {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		cpu := *cmd.Flags().Lookup("cpu").Value.(*resource.Millicores)
		ram := *cmd.Flags().Lookup("ram").Value.(*resource.ByteQuantity)
		gpu := *cmd.Flags().Lookup("gpu").Value.(*resource.Millicores)
		multiAz, _ := cmd.Flags().GetBool("multi-az")
		pkgs, err := filteredPackages(cmd, s, packageFilter{
			CPU:        cpu,
			RAM:        ram,
			GPU:        gpu,
			IncludeGPU: gpu != 0,
			MultiAz:    multiAz,
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		seen := make(map[resource.ByteQuantity]struct{})
		var completions []string
		for _, p := range pkgs {
			v, err := resource.ParseByteQuantity(p.GetResourceConfiguration().GetDisk())
			if err != nil {
				cobra.CompErrorln(fmt.Sprintf("package %s: %v", p.GetName(), err))
				continue
			}
			if _, ok := seen[v]; !ok {
				seen[v] = struct{}{}
				completions = append(completions, v.String())
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// gpuCompletion returns a completion function for the --gpu flag.
func gpuCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		provider, _ := cmd.Flags().GetString("cloud-provider")
		if provider == "" {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		cpu := *cmd.Flags().Lookup("cpu").Value.(*resource.Millicores)
		ram := *cmd.Flags().Lookup("ram").Value.(*resource.ByteQuantity)
		multiAz, _ := cmd.Flags().GetBool("multi-az")
		pkgs, err := filteredPackages(cmd, s, packageFilter{
			CPU:        cpu,
			RAM:        ram,
			IncludeGPU: true,
			MultiAz:    multiAz,
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		seen := make(map[resource.Millicores]struct{})
		var completions []string
		for _, p := range pkgs {
			m, err := resource.ParseMillicores(p.GetResourceConfiguration().GetGpu())
			if err != nil || m <= 0 || int64(m)%1000 != 0 {
				continue
			}
			if _, ok := seen[m]; !ok {
				seen[m] = struct{}{}
				completions = append(completions, m.String())
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// diskPerformanceCompletion returns a static completion function for the --disk-performance flag.
func diskPerformanceCompletion() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{diskPerfBalanced, diskPerfCostOptimised, diskPerfPerformance}, cobra.ShellCompDirectiveNoFileComp
	}
}

// restartModeCompletion returns a static completion function for the --restart-mode flag.
func restartModeCompletion() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{restartModeRolling, restartModeParallel, restartModeAutomatic}, cobra.ShellCompDirectiveNoFileComp
	}
}

// rebalanceStrategyCompletion returns a static completion function for the --rebalance-strategy flag.
func rebalanceStrategyCompletion() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{rebalanceByCount, rebalanceBySize, rebalanceByCountAndSize}, cobra.ShellCompDirectiveNoFileComp
	}
}

// packageCompletion returns a completion function for the --package flag.
func packageCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		provider, _ := cmd.Flags().GetString("cloud-provider")
		if provider == "" {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		ctx := cmd.Context()
		client, err := s.Client(ctx)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		accountID, err := s.AccountID()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		region, _ := cmd.Flags().GetString("cloud-region")
		req := &bookingv1.ListPackagesRequest{
			AccountId:       accountID,
			CloudProviderId: provider,
			Statuses:        []bookingv1.PackageStatus{bookingv1.PackageStatus_PACKAGE_STATUS_ACTIVE},
		}
		if region != "" {
			req.CloudProviderRegionId = &region
		}

		resp, err := client.Booking().ListPackages(ctx, req)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		completions := make([]string, 0, len(resp.GetItems()))
		for _, p := range resp.GetItems() {
			desc := output.PackageTier(p.GetTier())
			if rc := p.GetResourceConfiguration(); rc != nil {
				desc += fmt.Sprintf(" | %s RAM / %s CPU / %s disk", rc.GetRam(), rc.GetCpu(), rc.GetDisk())
			}
			desc += " | " + formatMillicents(p.GetUnitIntPricePerHour(), p.GetCurrency()) + "/hr"
			completions = append(completions, p.GetName()+"\t"+desc)
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}
