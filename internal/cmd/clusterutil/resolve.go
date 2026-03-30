package clusterutil

import (
	"context"
	"fmt"
	"strings"

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
) (*bookingv1.Package, error) {
	resp, err := booking.ListPackages(ctx, &bookingv1.ListPackagesRequest{
		AccountId:             accountID,
		CloudProviderId:       cloudProvider,
		CloudProviderRegionId: cloudRegion,
		Statuses:              []bookingv1.PackageStatus{bookingv1.PackageStatus_PACKAGE_STATUS_ACTIVE},
	})
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

// ResolvePackageByResources lists active packages and returns the unique one matching
// all non-zero resource dimensions (cpu, ram, gpu) and the multiAz flag.
// Returns an error if zero or more than one package matches.
func ResolvePackageByResources(
	ctx context.Context,
	booking bookingv1.BookingServiceClient,
	accountID, cloudProvider string,
	cloudRegion *string,
	cpu, gpu resource.Millicores,
	ram resource.ByteQuantity,
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

	if gpu != 0 {
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
		if cpu != 0 {
			pkgCPU, _ := resource.ParseMillicores(rc.GetCpu())
			if pkgCPU != cpu {
				continue
			}
		}
		if ram != 0 {
			pkgRAM, _ := resource.ParseByteQuantity(rc.GetRam())
			if pkgRAM != ram {
				continue
			}
		}
		if gpu != 0 {
			pkgGPU, _ := resource.ParseMillicores(rc.GetGpu())
			if pkgGPU != gpu {
				continue
			}
		}
		matches = append(matches, p)
	}

	var filterDesc []string
	if cpu != 0 {
		filterDesc = append(filterDesc, fmt.Sprintf("cpu=%q", cpu.String()))
	}
	if ram != 0 {
		filterDesc = append(filterDesc, fmt.Sprintf("ram=%q", ram.String()))
	}
	if gpu != 0 {
		filterDesc = append(filterDesc, fmt.Sprintf("gpu=%q", gpu.String()))
	}
	desc := strings.Join(filterDesc, " ")

	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return nil, fmt.Errorf("no package found matching %s; use 'cluster package list' to see available packages", desc)
	default:
		names := make([]string, len(matches))
		for i, p := range matches {
			names[i] = p.GetName()
		}
		return nil, fmt.Errorf("multiple packages match %s: %s; use 'cluster package list' to see available packages", desc, strings.Join(names, ", "))
	}
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
