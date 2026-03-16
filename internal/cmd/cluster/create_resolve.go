package cluster

import (
	"context"
	"fmt"
	"strings"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"
)

// resolvePackageByName lists active packages and returns the first matching by name.
func resolvePackageByName(ctx context.Context, booking bookingv1.BookingServiceClient,
	accountID, cloudProvider, cloudRegion, name string) (*bookingv1.Package, error) {
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
func resolvePackageByResources(ctx context.Context, booking bookingv1.BookingServiceClient,
	accountID, cloudProvider, cloudRegion, cpu, ram, disk, gpu string, multiAz bool) (*bookingv1.Package, error) {
	req := &bookingv1.ListPackagesRequest{
		AccountId:             accountID,
		CloudProviderId:       cloudProvider,
		CloudProviderRegionId: &cloudRegion,
		Statuses:              []bookingv1.PackageStatus{bookingv1.PackageStatus_PACKAGE_STATUS_ACTIVE},
	}
	if multiAz {
		req.MultiAz = new(true)
	}
	resp, err := booking.ListPackages(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}
	var matches []*bookingv1.Package
	for _, p := range resp.GetItems() {
		rc := p.GetResourceConfiguration()
		if cpu != "" && rc.GetCpu() != cpu {
			continue
		}
		if ram != "" && rc.GetRam() != ram {
			continue
		}
		if disk != "" && rc.GetDisk() != disk {
			continue
		}
		if gpu != "" && rc.GetGpu() != gpu {
			continue
		}
		matches = append(matches, p)
	}
	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return nil, fmt.Errorf("no package found matching cpu=%q ram=%q disk=%q gpu=%q", cpu, ram, disk, gpu)
	default:
		names := make([]string, len(matches))
		for i, p := range matches {
			names[i] = p.GetName()
		}
		return nil, fmt.Errorf("multiple packages match cpu=%q ram=%q disk=%q gpu=%q: %s", cpu, ram, disk, gpu, strings.Join(names, ", "))
	}
}
