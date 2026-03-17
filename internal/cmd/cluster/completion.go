package cluster

import (
	"fmt"

	"github.com/spf13/cobra"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"
	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	platformv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/platform/v1"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// cloudProviderCompletion returns a completion function for the --cloud-provider flag.
func cloudProviderCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
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

		resp, err := client.Platform().ListCloudProviders(ctx, &platformv1.ListCloudProvidersRequest{
			AccountId: accountID,
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		completions := make([]string, 0, len(resp.GetItems()))
		for _, p := range resp.GetItems() {
			completions = append(completions, p.GetId()+"\t"+p.GetName())
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// cloudRegionCompletion returns a completion function for the --cloud-region flag.
func cloudRegionCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
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

		resp, err := client.Platform().ListCloudProviderRegions(ctx, &platformv1.ListCloudProviderRegionsRequest{
			AccountId:       accountID,
			CloudProviderId: provider,
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		completions := make([]string, 0, len(resp.GetItems()))
		for _, r := range resp.GetItems() {
			completions = append(completions, r.GetId()+"\t"+r.GetName())
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

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

// filteredPackages fetches active packages filtered by non-empty cpu/ram values and the multiAz flag.
// Returns nil (no completions) if --cloud-provider is not set.
func filteredPackages(cmd *cobra.Command, s *state.State, cpu, ram string, multiAz bool) ([]*bookingv1.Package, error) {
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
	}
	if region != "" {
		req.CloudProviderRegionId = &region
	}
	if multiAz {
		req.MultiAz = new(true)
	}

	resp, err := client.Booking().ListPackages(ctx, req)
	if err != nil {
		return nil, err
	}

	var result []*bookingv1.Package
	for _, p := range resp.GetItems() {
		rc := p.GetResourceConfiguration()
		if cpu != "" && rc.GetCpu() != cpu {
			continue
		}
		if ram != "" && rc.GetRam() != ram {
			continue
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
		multiAz, _ := cmd.Flags().GetBool("multi-az")
		pkgs, err := filteredPackages(cmd, s, "", ram, multiAz)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		seen := make(map[string]struct{})
		var completions []string
		for _, p := range pkgs {
			v := p.GetResourceConfiguration().GetCpu()
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
		multiAz, _ := cmd.Flags().GetBool("multi-az")
		pkgs, err := filteredPackages(cmd, s, cpu, "", multiAz)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		seen := make(map[string]struct{})
		var completions []string
		for _, p := range pkgs {
			v := p.GetResourceConfiguration().GetRam()
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
		multiAz, _ := cmd.Flags().GetBool("multi-az")
		pkgs, err := filteredPackages(cmd, s, cpu, ram, multiAz)
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
