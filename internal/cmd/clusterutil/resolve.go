package clusterutil

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/resource"
	"github.com/qdrant/qcloud-cli/internal/state"
)

// PackageFilter holds the parameters for filtering packages.
// Zero values for CPU/RAM/GPU mean "do not filter on that dimension".
type PackageFilter struct {
	CPU        resource.Millicores
	RAM        resource.ByteQuantity
	GPU        resource.Millicores
	IncludeGPU bool
	MultiAz    bool
}

// FilteredPackages fetches active packages matching the given filter.
func FilteredPackages(
	cmd *cobra.Command,
	s *state.State,
	cloudProvider string,
	cloudRegion *string,
	f PackageFilter,
) ([]*bookingv1.Package, error) {
	ctx := cmd.Context()
	client, err := s.Client(ctx)
	if err != nil {
		return nil, err
	}

	accountID, err := s.AccountID()
	if err != nil {
		return nil, err
	}

	req := &bookingv1.ListPackagesRequest{
		AccountId:             accountID,
		CloudProviderId:       cloudProvider,
		CloudProviderRegionId: cloudRegion,
		Statuses:              []bookingv1.PackageStatus{bookingv1.PackageStatus_PACKAGE_STATUS_ACTIVE},
		Gpu:                   &f.IncludeGPU,
		// Leaving multi_az unset would also return multi-AZ packages.
		MultiAz: &f.MultiAz,
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

// ResolvePackageByID fetches a single package by its UUID.
func ResolvePackageByID(
	ctx context.Context,
	booking bookingv1.BookingServiceClient,
	accountID, cloudProvider string,
	cloudRegion *string,
	id string,
) (*bookingv1.Package, error) {
	resp, err := booking.GetPackage(ctx, &bookingv1.GetPackageRequest{
		AccountId:             accountID,
		CloudProviderId:       cloudProvider,
		CloudProviderRegionId: cloudRegion,
		Id:                    id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get package %s: %w", id, err)
	}

	return resp.GetPackage(), nil
}

// ResolvePackageByName lists active packages and returns the first matching by name.
func ResolvePackageByName(
	ctx context.Context,
	booking bookingv1.BookingServiceClient,
	accountID, cloudProvider string,
	cloudRegion *string,
	name string,
	multiAz bool,
) (*bookingv1.Package, error) {
	req := &bookingv1.ListPackagesRequest{
		AccountId:             accountID,
		CloudProviderId:       cloudProvider,
		CloudProviderRegionId: cloudRegion,
		Statuses:              []bookingv1.PackageStatus{bookingv1.PackageStatus_PACKAGE_STATUS_ACTIVE},
		// Leaving multi_az unset would also return multi-AZ packages, which
		// share their names with the single-AZ ones.
		MultiAz: &multiAz,
	}

	resp, err := booking.ListPackages(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	for _, p := range resp.GetItems() {
		if p.GetName() == name {
			return p, nil
		}
	}

	return nil, fmt.Errorf("package %q not found for provider=%s", name, cloudProvider)
}

// PackageResourceQuery holds the parameters for resolving a package by resources.
type PackageResourceQuery struct {
	AccountID     string
	CloudProvider string
	CloudRegion   *string
	CPU           resource.Millicores
	GPU           resource.Millicores
	RAM           resource.ByteQuantity
	MultiAz       bool
	// Current is the cluster's current package when scaling, nil when creating.
	// It is never suggested, and a request that lowers any of its resources is
	// not given a smallest package that fits.
	Current *bookingv1.Package
}

// lowersCurrent reports whether the query asks for less of any resource than
// the current package has.
func (q PackageResourceQuery) lowersCurrent() bool {
	if q.Current == nil {
		return false
	}

	rc := q.Current.GetResourceConfiguration()
	cpu, _ := resource.ParseMillicores(rc.GetCpu())
	ram, _ := resource.ParseByteQuantity(rc.GetRam())
	gpu, _ := resource.ParseMillicores(rc.GetGpu())

	return (q.CPU != 0 && q.CPU < cpu) || (q.RAM != 0 && q.RAM < ram) || (q.GPU != 0 && q.GPU < gpu)
}

// ResolvePackageByResources lists active packages and returns the unique one matching
// all non-zero resource dimensions (cpu, ram, gpu) and the multiAz flag.
// Returns an error if zero or more than one package matches.
func ResolvePackageByResources(
	ctx context.Context,
	booking bookingv1.BookingServiceClient,
	q PackageResourceQuery,
) (*bookingv1.Package, error) {
	req := &bookingv1.ListPackagesRequest{
		AccountId:             q.AccountID,
		CloudProviderId:       q.CloudProvider,
		CloudProviderRegionId: q.CloudRegion,
		Statuses:              []bookingv1.PackageStatus{bookingv1.PackageStatus_PACKAGE_STATUS_ACTIVE},
		// Leaving multi_az unset would also return multi-AZ packages.
		MultiAz: &q.MultiAz,
	}

	if q.GPU != 0 {
		req.Gpu = new(true)
	} else {
		req.Gpu = new(false)
	}

	resp, err := booking.ListPackages(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	var matches []*bookingv1.Package
	for _, p := range resp.GetItems() {
		rc := p.GetResourceConfiguration()
		if q.CPU != 0 {
			pkgCPU, _ := resource.ParseMillicores(rc.GetCpu())
			if pkgCPU != q.CPU {
				continue
			}
		}

		if q.RAM != 0 {
			pkgRAM, _ := resource.ParseByteQuantity(rc.GetRam())
			if pkgRAM != q.RAM {
				continue
			}
		}

		if q.GPU != 0 {
			pkgGPU, _ := resource.ParseMillicores(rc.GetGpu())
			if pkgGPU != q.GPU {
				continue
			}
		}

		matches = append(matches, p)
	}

	var filterDesc []string
	if q.CPU != 0 {
		filterDesc = append(filterDesc, fmt.Sprintf("cpu=%q", q.CPU.String()))
	}

	if q.RAM != 0 {
		filterDesc = append(filterDesc, fmt.Sprintf("ram=%q", q.RAM.String()))
	}

	if q.GPU != 0 {
		filterDesc = append(filterDesc, fmt.Sprintf("gpu=%q", q.GPU.String()))
	}

	desc := strings.Join(filterDesc, " ")

	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		suggestions, hasFit := closestPackages(q, resp.GetItems(), maxPackageSuggestions)
		if len(suggestions) == 0 {
			return nil, fmt.Errorf("no package found matching %s; use 'package list' to see available packages", desc)
		}

		return nil, noPackageFoundError(desc, suggestions, hasFit, q.GPU != 0)
	default:
		names := make([]string, len(matches))
		for i, p := range matches {
			names[i] = fmt.Sprintf("%s (id=%s tier=%s multi-az=%s)", p.GetName(), p.GetId(), output.PackageTier(p.GetTier()), output.BoolYesNo(p.GetMultiAz()))
		}

		return nil, fmt.Errorf("multiple packages match %s: %s; use --package with a package ID to pick one", desc, strings.Join(names, ", "))
	}
}

// maxPackageSuggestions is how many packages a "no package found" error suggests.
const maxPackageSuggestions = 5

// packageSuggestion is a package together with its parsed resources and its
// distance from the requested resources.
type packageSuggestion struct {
	pkg      *bookingv1.Package
	cpu      resource.Millicores
	ram      resource.ByteQuantity
	gpu      resource.Millicores
	distance float64
	// fits is true if the package has at least the requested amount of every resource.
	fits bool
}

// closestPackages returns up to limit packages ordered by how far their
// resources are from the requested ones. The distance sums, over every
// requested dimension, the absolute log2 ratio between package and request, so
// doubling or halving a dimension counts the same and a package that matches
// one dimension exactly ranks above one that is off in all of them.
// The smallest package that fits, the one with the least RAM, then CPU, then
// GPU among those that have at least the requested amount of every requested
// resource, always comes first. Unrequested resources do not decide whether a
// package fits, but they do decide which fitting package is the smallest.
// When scaling, the current package is left out, and a request that lowers a
// resource gets no smallest fit, since a package that fits it may be no smaller
// than the current one. hasFit reports whether the first suggestion is the
// smallest fit.
func closestPackages(q PackageResourceQuery, pkgs []*bookingv1.Package, limit int) (suggestions []packageSuggestion, hasFit bool) {
	for _, p := range pkgs {
		if q.Current != nil && p.GetId() == q.Current.GetId() {
			continue
		}

		rc := p.GetResourceConfiguration()
		cpu, _ := resource.ParseMillicores(rc.GetCpu())
		ram, _ := resource.ParseByteQuantity(rc.GetRam())
		gpu, _ := resource.ParseMillicores(rc.GetGpu())

		var distance float64
		missing := false
		fits := true
		for _, d := range []struct{ want, have int64 }{
			{int64(q.CPU), int64(cpu)},
			{int64(q.RAM), int64(ram)},
			{int64(q.GPU), int64(gpu)},
		} {
			if d.want == 0 {
				continue
			}

			if d.have <= 0 {
				missing = true
				break
			}

			distance += math.Abs(math.Log2(float64(d.have) / float64(d.want)))
			if d.have < d.want {
				fits = false
			}
		}

		if missing {
			continue
		}

		suggestions = append(suggestions, packageSuggestion{pkg: p, cpu: cpu, ram: ram, gpu: gpu, distance: distance, fits: fits})
	}

	slices.SortStableFunc(suggestions, func(a, b packageSuggestion) int {
		return cmp.Or(cmp.Compare(a.distance, b.distance), compareSize(a, b))
	})

	smallest := -1
	if !q.lowersCurrent() {
		for i, sg := range suggestions {
			if sg.fits && (smallest < 0 || compareSize(sg, suggestions[smallest]) < 0) {
				smallest = i
			}
		}
	}

	if smallest > 0 {
		fit := suggestions[smallest]
		copy(suggestions[1:smallest+1], suggestions[:smallest])
		suggestions[0] = fit
	}

	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	return suggestions, smallest >= 0
}

// compareSize orders packages by RAM, then CPU, then GPU.
func compareSize(a, b packageSuggestion) int {
	return cmp.Or(
		cmp.Compare(a.ram, b.ram),
		cmp.Compare(a.cpu, b.cpu),
		cmp.Compare(a.gpu, b.gpu),
	)
}

// noPackageFoundError builds the error for a resource request that no package
// matches. It names the smallest package that fits if hasFit is set, and lists the
// suggestions as an indented table, one package per line.
func noPackageFoundError(desc string, suggestions []packageSuggestion, hasFit, includeGPU bool) error {
	var sb strings.Builder
	fmt.Fprintf(&sb, "no package found matching %s\n", desc)

	if fit := suggestions[0]; hasFit {
		rc := fit.pkg.GetResourceConfiguration()
		fmt.Fprintf(&sb, "smallest package that fits: %s (cpu=%s ram=%s", fit.pkg.GetName(), output.PackageCPU(rc), output.PackageRAM(rc))
		if includeGPU {
			fmt.Fprintf(&sb, " gpu=%s", output.PackageGPU(rc))
		}

		sb.WriteString(")\n")
	}

	sb.WriteString("closest available packages:\n")

	var table strings.Builder
	t := output.NewTable[packageSuggestion](&table)
	t.AddField("NAME", func(sg packageSuggestion) string { return sg.pkg.GetName() })
	t.AddField("CPU", func(sg packageSuggestion) string { return output.PackageCPU(sg.pkg.GetResourceConfiguration()) })
	t.AddField("RAM", func(sg packageSuggestion) string { return output.PackageRAM(sg.pkg.GetResourceConfiguration()) })
	if includeGPU {
		t.AddField("GPU", func(sg packageSuggestion) string { return output.PackageGPU(sg.pkg.GetResourceConfiguration()) })
	}

	t.AddField("DISK", func(sg packageSuggestion) string { return output.PackageDisk(sg.pkg.GetResourceConfiguration()) })
	t.Write(suggestions)

	for line := range strings.Lines(table.String()) {
		sb.WriteString("  " + strings.TrimRight(line, " \n") + "\n")
	}

	sb.WriteString("use 'package list' to see all available packages")

	return errors.New(sb.String())
}

// CalculateAdditionalDisk returns the additional disk (in GiB) needed beyond
// what the package includes. Returns 0 if the requested disk is not larger
// than the package's included disk.
func CalculateAdditionalDisk(requestedDisk resource.ByteQuantity, pkg *bookingv1.Package) (uint32, error) {
	pkgDiskStr := pkg.GetResourceConfiguration().GetDisk()
	if pkgDiskStr == "" {
		return 0, nil
	}

	pkgDisk, err := resource.ParseByteQuantity(pkgDiskStr)
	if err != nil {
		return 0, err
	}

	if requestedDisk > pkgDisk {
		return uint32(requestedDisk.GiB() - pkgDisk.GiB()), nil
	}

	return 0, nil
}
