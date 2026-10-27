package clusterutil

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"

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
	}
	if multiAz {
		req.MultiAz = new(true)
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
	}
	if q.MultiAz {
		req.MultiAz = new(true)
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
		suggestions := closestPackages(q, resp.GetItems(), maxPackageSuggestions)
		if len(suggestions) == 0 {
			return nil, fmt.Errorf("no package found matching %s; use 'package list' to see available packages", desc)
		}

		return nil, noPackageFoundError(desc, suggestions, q.GPU != 0)
	default:
		names := make([]string, len(matches))
		for i, p := range matches {
			names[i] = p.GetName()
		}

		return nil, fmt.Errorf("multiple packages match %s: %s; use 'package list' to see available packages", desc, strings.Join(names, ", "))
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
// The next biggest package, the closest one that has at least the requested
// amount of every resource, always comes first.
func closestPackages(q PackageResourceQuery, pkgs []*bookingv1.Package, limit int) []packageSuggestion {
	var suggestions []packageSuggestion
	for _, p := range pkgs {
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
		return cmp.Or(
			cmp.Compare(a.distance, b.distance),
			cmp.Compare(a.cpu, b.cpu),
			cmp.Compare(a.ram, b.ram),
			cmp.Compare(a.gpu, b.gpu),
		)
	})

	if i := slices.IndexFunc(suggestions, func(sg packageSuggestion) bool { return sg.fits }); i > 0 {
		next := suggestions[i]
		copy(suggestions[1:i+1], suggestions[:i])
		suggestions[0] = next
	}

	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	return suggestions
}

// noPackageFoundError builds the error for a resource request that no package
// matches. It names the next biggest package, if there is one, and lists the
// suggestions as an indented table, one package per line.
func noPackageFoundError(desc string, suggestions []packageSuggestion, includeGPU bool) error {
	var sb strings.Builder
	fmt.Fprintf(&sb, "no package found matching %s\n", desc)

	if next := suggestions[0]; next.fits {
		fmt.Fprintf(&sb, "next biggest package: %s (cpu=%s ram=%s", next.pkg.GetName(), next.cpu.String(), next.ram.String())
		if includeGPU {
			fmt.Fprintf(&sb, " gpu=%s", next.gpu.String())
		}

		sb.WriteString(")\n")
	}

	sb.WriteString("closest available packages:\n")

	w := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', 0)

	header := "  NAME\tCPU\tRAM"
	if includeGPU {
		header += "\tGPU"
	}

	fmt.Fprintln(w, header+"\tDISK")

	for _, sg := range suggestions {
		line := fmt.Sprintf("  %s\t%s\t%s", sg.pkg.GetName(), sg.cpu.String(), sg.ram.String())
		if includeGPU {
			line += "\t" + sg.gpu.String()
		}

		disk := sg.pkg.GetResourceConfiguration().GetDisk()
		if parsed, err := resource.ParseByteQuantity(disk); err == nil {
			disk = parsed.String()
		}

		fmt.Fprintln(w, line+"\t"+disk)
	}

	_ = w.Flush()

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
