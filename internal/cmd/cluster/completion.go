package cluster

import (
	"fmt"

	"github.com/spf13/cobra"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"
	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// versionCompletion returns a completion function for the --version flag.
func versionCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		ctx := cmd.Context()
		client, err := s.Client(ctx)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		accountID, err := s.AccountID()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		resp, err := client.Cluster().ListQdrantReleases(ctx, &clusterv1.ListQdrantReleasesRequest{
			AccountId: accountID,
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		completions := make([]string, 0, len(resp.GetItems()))
		for _, r := range resp.GetItems() {
			if r.GetUnavailable() {
				continue
			}
			desc := ""
			if r.GetDefault() {
				desc += "(default)"
			}
			if r.GetEndOfLife() {
				if desc != "" {
					desc += " "
				}
				desc += "(end of life)"
			}
			if remarks := r.GetRemarks(); remarks != "" {
				if desc != "" {
					desc += " "
				}
				desc += remarks
			}
			entry := r.GetVersion()
			if desc != "" {
				entry += "\t" + desc
			}
			completions = append(completions, entry)
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// packageFilter holds the parameters for filtering packages.
type packageFilter struct {
	CPU        string
	RAM        string
	GPU        string
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

	var wantCPU, wantRAM, wantGPU int64
	if f.CPU != "" {
		wantCPU, _ = parseCPUMillicores(f.CPU)
	}
	if f.RAM != "" {
		wantRAM, _ = parseRAMGiB(f.RAM)
	}
	if f.GPU != "" {
		wantGPU, _ = parseGPUMillicores(f.GPU)
	}

	var result []*bookingv1.Package
	for _, p := range resp.GetItems() {
		rc := p.GetResourceConfiguration()
		if f.CPU != "" {
			pkgCPU, _ := parseCPUMillicores(rc.GetCpu())
			if pkgCPU != wantCPU {
				continue
			}
		}
		if f.RAM != "" {
			pkgRAM, _ := parseRAMGiB(rc.GetRam())
			if pkgRAM != wantRAM {
				continue
			}
		}
		if f.GPU != "" {
			pkgGPU, _ := parseGPUMillicores(rc.GetGpu())
			if pkgGPU != wantGPU {
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
		ram, _ := cmd.Flags().GetString("ram")
		gpu, err := gpuFlagToMillicores(cmd)
		if err != nil {
			cobra.CompErrorln(err.Error())
			return nil, cobra.ShellCompDirectiveError
		}
		multiAz, _ := cmd.Flags().GetBool("multi-az")
		filter := packageFilter{
			RAM:        ram,
			GPU:        gpu,
			IncludeGPU: gpu != "",
			MultiAz:    multiAz,
		}
		pkgs, err := filteredPackages(cmd, s, filter)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		seen := make(map[string]struct{})
		var completions []string
		for _, p := range pkgs {
			v, err := normalizeMillicores(p.GetResourceConfiguration().GetCpu())
			if err != nil {
				cobra.CompErrorln(fmt.Sprintf("package %s: %v", p.GetName(), err))
				continue
			}
			if _, ok := seen[v]; !ok {
				seen[v] = struct{}{}
				completions = append(completions, v)
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
		cpu, _ := cmd.Flags().GetString("cpu")
		gpu, err := gpuFlagToMillicores(cmd)
		if err != nil {
			cobra.CompErrorln(err.Error())
			return nil, cobra.ShellCompDirectiveError
		}
		multiAz, _ := cmd.Flags().GetBool("multi-az")
		filter := packageFilter{
			CPU:        cpu,
			GPU:        gpu,
			IncludeGPU: gpu != "",
			MultiAz:    multiAz,
		}
		pkgs, err := filteredPackages(cmd, s, filter)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		seen := make(map[string]struct{})
		var completions []string
		for _, p := range pkgs {
			v, err := normalizeRAM(p.GetResourceConfiguration().GetRam())
			if err != nil {
				cobra.CompErrorln(fmt.Sprintf("package %s: %v", p.GetName(), err))
				continue
			}
			if _, ok := seen[v]; !ok {
				seen[v] = struct{}{}
				completions = append(completions, v)
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
		cpu, _ := cmd.Flags().GetString("cpu")
		ram, _ := cmd.Flags().GetString("ram")
		gpu, err := gpuFlagToMillicores(cmd)
		if err != nil {
			cobra.CompErrorln(err.Error())
			return nil, cobra.ShellCompDirectiveError
		}
		multiAz, _ := cmd.Flags().GetBool("multi-az")
		filter := packageFilter{
			CPU:        cpu,
			RAM:        ram,
			GPU:        gpu,
			IncludeGPU: gpu != "",
			MultiAz:    multiAz,
		}
		pkgs, err := filteredPackages(cmd, s, filter)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		seen := make(map[string]struct{})
		var completions []string
		for _, p := range pkgs {
			v := p.GetResourceConfiguration().GetDisk()
			if _, ok := seen[v]; !ok {
				seen[v] = struct{}{}
				completions = append(completions, v)
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
		cpu, _ := cmd.Flags().GetString("cpu")
		ram, _ := cmd.Flags().GetString("ram")
		multiAz, _ := cmd.Flags().GetBool("multi-az")
		filter := packageFilter{
			CPU:        cpu,
			RAM:        ram,
			IncludeGPU: true,
			MultiAz:    multiAz,
		}
		pkgs, err := filteredPackages(cmd, s, filter)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		seen := make(map[string]struct{})
		var completions []string
		for _, p := range pkgs {
			v := p.GetResourceConfiguration().GetGpu()
			if v == "" {
				continue
			}
			// Convert millicores (e.g. "1000m") to integer GPUs (e.g. "1")
			// to match the --gpu flag's Int type.
			intVal := gpuMillicoresToCount(v)
			if intVal == "" {
				continue
			}
			if _, ok := seen[intVal]; !ok {
				seen[intVal] = struct{}{}
				completions = append(completions, intVal)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
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
			desc := packageTierString(p.GetTier())
			if rc := p.GetResourceConfiguration(); rc != nil {
				desc += fmt.Sprintf(" | %s RAM / %s CPU / %s disk", rc.GetRam(), rc.GetCpu(), rc.GetDisk())
			}
			desc += " | " + formatMillicents(p.GetUnitIntPricePerHour(), p.GetCurrency()) + "/hr"
			completions = append(completions, p.GetName()+"\t"+desc)
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}
