package cluster

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newPackageCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "package",
		Short: "Manage cluster packages",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newPackageListCommand(s))
	return cmd
}

func newPackageListCommand(s *state.State) *cobra.Command {
	cmd := base.ListCmd[*bookingv1.ListPackagesResponse]{
		Use:   "list",
		Short: "List available packages for cluster creation",
		Fetch: func(s *state.State, cmd *cobra.Command) (*bookingv1.ListPackagesResponse, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			cloudProvider, _ := cmd.Flags().GetString("cloud-provider")
			cloudRegion, _ := cmd.Flags().GetString("cloud-region")

			resp, err := client.Booking().ListPackages(ctx, &bookingv1.ListPackagesRequest{
				AccountId:             accountID,
				CloudProviderId:       cloudProvider,
				CloudProviderRegionId: &cloudRegion,
				Statuses:              []bookingv1.PackageStatus{bookingv1.PackageStatus_PACKAGE_STATUS_ACTIVE},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list packages: %w", err)
			}

			return resp, nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, resp *bookingv1.ListPackagesResponse) error {
			t := output.NewTable[*bookingv1.Package](w)
			t.AddField("NAME", func(p *bookingv1.Package) string {
				return p.GetName()
			})
			t.AddField("ID", func(p *bookingv1.Package) string {
				return p.GetId()
			})
			t.AddField("TIER", func(p *bookingv1.Package) string {
				return strings.TrimPrefix(p.GetTier().String(), "PACKAGE_TIER_")
			})
			t.AddField("RAM", func(p *bookingv1.Package) string {
				if rc := p.GetResourceConfiguration(); rc != nil {
					return rc.GetRam()
				}
				return ""
			})
			t.AddField("CPU", func(p *bookingv1.Package) string {
				if rc := p.GetResourceConfiguration(); rc != nil {
					return rc.GetCpu()
				}
				return ""
			})
			t.AddField("DISK", func(p *bookingv1.Package) string {
				if rc := p.GetResourceConfiguration(); rc != nil {
					return rc.GetDisk()
				}
				return ""
			})
			t.AddField("PRICE/HR", func(p *bookingv1.Package) string {
				return formatMillicents(p.GetUnitIntPricePerHour(), p.GetCurrency())
			})
			t.Write(resp.GetItems())
			return nil
		},
	}.CobraCommand(s)

	cmd.Flags().String("cloud-provider", "", "Cloud provider ID (required)")
	cmd.Flags().String("cloud-region", "", "Cloud provider region ID (required)")
	_ = cmd.MarkFlagRequired("cloud-provider")
	_ = cmd.MarkFlagRequired("cloud-region")
	return cmd
}
