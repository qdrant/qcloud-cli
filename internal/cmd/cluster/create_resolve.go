package cluster

import (
	"context"
	"fmt"
	"strings"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"
)

// resolvePackageByID fetches a single package by its UUID.
func resolvePackageByID(ctx context.Context, booking bookingv1.BookingServiceClient,
	accountID, cloudProvider, cloudRegion, id string) (*bookingv1.Package, error) {
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
	accountID, cloudProvider, cloudRegion, cpu, ram string, multiAz bool) (*bookingv1.Package, error) {
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
		matches = append(matches, p)
	}
	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return nil, fmt.Errorf("no package found matching cpu=%q ram=%q", cpu, ram)
	default:
		names := make([]string, len(matches))
		for i, p := range matches {
			names[i] = p.GetName()
		}
		return nil, fmt.Errorf("multiple packages match cpu=%q ram=%q: %s", cpu, ram, strings.Join(names, ", "))
	}
}
