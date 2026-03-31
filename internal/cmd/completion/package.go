package completion

import (
	"fmt"

	"github.com/spf13/cobra"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/clusterutil"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/resource"
	"github.com/qdrant/qcloud-cli/internal/state"
)

// CPUCompletion returns a completion function for a --cpu flag.
func CPUCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		provider, region, err := getCloudValuesFromFlags(cmd)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		ram := *cmd.Flags().Lookup("ram").Value.(*resource.ByteQuantity)
		gpu := *cmd.Flags().Lookup("gpu").Value.(*resource.Millicores)
		multiAz, _ := cmd.Flags().GetBool("multi-az")
		pkgs, err := clusterutil.FilteredPackages(cmd, s, provider, region, clusterutil.PackageFilter{
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

// RAMCompletion returns a completion function for a --ram flag.
func RAMCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		provider, region, err := getCloudValuesFromFlags(cmd)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		cpu := *cmd.Flags().Lookup("cpu").Value.(*resource.Millicores)
		gpu := *cmd.Flags().Lookup("gpu").Value.(*resource.Millicores)
		multiAz, _ := cmd.Flags().GetBool("multi-az")
		pkgs, err := clusterutil.FilteredPackages(cmd, s, provider, region, clusterutil.PackageFilter{
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

// DiskCompletion returns a completion function for a --disk flag.
func DiskCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		provider, region, err := getCloudValuesFromFlags(cmd)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		cpu := *cmd.Flags().Lookup("cpu").Value.(*resource.Millicores)
		ram := *cmd.Flags().Lookup("ram").Value.(*resource.ByteQuantity)
		gpu := *cmd.Flags().Lookup("gpu").Value.(*resource.Millicores)
		multiAz, _ := cmd.Flags().GetBool("multi-az")
		pkgs, err := clusterutil.FilteredPackages(cmd, s, provider, region, clusterutil.PackageFilter{
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

// GPUCompletion returns a completion function for a --gpu flag.
func GPUCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		provider, region, err := getCloudValuesFromFlags(cmd)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		cpu := *cmd.Flags().Lookup("cpu").Value.(*resource.Millicores)
		ram := *cmd.Flags().Lookup("ram").Value.(*resource.ByteQuantity)
		multiAz, _ := cmd.Flags().GetBool("multi-az")
		pkgs, err := clusterutil.FilteredPackages(cmd, s, provider, region, clusterutil.PackageFilter{
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

// PackageNameCompletion returns a completion function for a --package flag.
func PackageNameCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		provider, region, err := getCloudValuesFromFlags(cmd)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
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

		req := &bookingv1.ListPackagesRequest{
			AccountId:             accountID,
			CloudProviderId:       provider,
			CloudProviderRegionId: region,
			Statuses:              []bookingv1.PackageStatus{bookingv1.PackageStatus_PACKAGE_STATUS_ACTIVE},
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
			desc += " | " + output.FormatMillicents(p.GetUnitIntPricePerHour(), p.GetCurrency()) + "/hr"
			completions = append(completions, p.GetName()+"\t"+desc)
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// getCloudValuesFromFlags returns (cloud, region, error) for completion of package related values. 
// Region can be nil when the cloud is 'hybrid'.
func getCloudValuesFromFlags(cmd *cobra.Command) (string, *string, error) {
	cloud, err := cmd.Flags().GetString("cloud-provider")
	if err != nil {
		return "", nil, err
	}

	r, err := cmd.Flags().GetString("cloud-region")
	if err != nil {
		return "", nil, err
	}


	return cloud, &r, nil
}
