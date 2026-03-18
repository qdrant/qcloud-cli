package cluster

import (
	"context"
	"fmt"
	"strings"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"
)

// resolvePackageByID fetches a single package by its UUID.
func resolvePackageByID(
	ctx context.Context,
	booking bookingv1.BookingServiceClient,
	accountID, cloudProvider, cloudRegion, id string,
) (*bookingv1.Package, error) {
	resp, err := booking.GetPackage(ctx, &bookingv1.GetPackageRequest{
		AccountId:             accountID,
		CloudProviderId:       cloudProvider,
		CloudProviderRegionId: &cloudRegion,
		Id:                    id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get package %s: %w", id, err)
	}

	return resp.GetPackage(), nil
}

// resolvePackageByName lists active packages and returns the first matching by name.
func resolvePackageByName(
	ctx context.Context,
	booking bookingv1.BookingServiceClient,
	accountID, cloudProvider, cloudRegion, name string,
) (*bookingv1.Package, error) {
	resp, err := booking.ListPackages(ctx, &bookingv1.ListPackagesRequest{
		AccountId:             accountID,
		CloudProviderId:       cloudProvider,
		CloudProviderRegionId: &cloudRegion,
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
	return nil, fmt.Errorf("package %q not found for provider=%s region=%s", name, cloudProvider, cloudRegion)
}

// resolvePackageByResources lists active packages and returns the unique one matching
// all non-empty resource dimensions (cpu, ram, disk, gpu) and the multiAz flag.
// Returns an error if zero or more than one package matches.
func resolvePackageByResources(
	ctx context.Context,
	booking bookingv1.BookingServiceClient,
	accountID, cloudProvider, cloudRegion, cpu, ram, gpu string,
	multiAz bool,
) (*bookingv1.Package, error) {
	req := &bookingv1.ListPackagesRequest{
		AccountId:             accountID,
		CloudProviderId:       cloudProvider,
		CloudProviderRegionId: &cloudRegion,
		Statuses:              []bookingv1.PackageStatus{bookingv1.PackageStatus_PACKAGE_STATUS_ACTIVE},
	}
	if multiAz {
		req.MultiAz = new(true)
	}

	if gpu != "" {
		req.Gpu = new(true)
	} else {
		req.Gpu = new(false)
	}

	resp, err := booking.ListPackages(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	var wantCPU, wantRAM, wantGPU int64
	if cpu != "" {
		wantCPU, _ = parseCPUMillicores(cpu)
	}
	if ram != "" {
		wantRAM, _ = parseRAMGiB(ram)
	}
	if gpu != "" {
		wantGPU, _ = parseGPUMillicores(gpu)
	}

	var matches []*bookingv1.Package
	for _, p := range resp.GetItems() {
		rc := p.GetResourceConfiguration()
		if cpu != "" {
			pkgCPU, _ := parseCPUMillicores(rc.GetCpu())
			if pkgCPU != wantCPU {
				continue
			}
		}
		if ram != "" {
			pkgRAM, _ := parseRAMGiB(rc.GetRam())
			if pkgRAM != wantRAM {
				continue
			}
		}
		if gpu != "" {
			pkgGPU, _ := parseGPUMillicores(rc.GetGpu())
			if pkgGPU != wantGPU {
				continue
			}
		}
		matches = append(matches, p)
	}

	var filterDesc []string
	if cpu != "" {
		filterDesc = append(filterDesc, fmt.Sprintf("cpu=%q", cpu))
	}
	if ram != "" {
		filterDesc = append(filterDesc, fmt.Sprintf("ram=%q", ram))
	}
	if gpu != "" {
		filterDesc = append(filterDesc, fmt.Sprintf("gpu=%q", gpu))
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
