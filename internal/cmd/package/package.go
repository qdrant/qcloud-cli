package packagecmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/qcloudapi"
	"github.com/qdrant/qcloud-cli/internal/state"
)

// NewCommand creates the "package" parent command and registers all subcommands.
func NewCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "package",
		Short: "Manage packages",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newListCommand(s))
	return cmd
}

func newListCommand(s *state.State) *cobra.Command {
	cmd := base.ListCmd[*bookingv1.ListPackagesResponse]{
		Use:   "list",
		Short: "List available packages for cluster creation",
		Example: `# List packages for a cloud provider and region
qcloud package list --cloud-provider aws --cloud-region eu-central-1

# List packages for a hybrid cloud provider (no region required)
qcloud package list --cloud-provider hybrid`,
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

			var cloudRegionPtr *string
			if cloudProvider == qcloudapi.HybridCloudProviderID {
				cloudRegionPtr = nil
			} else {
				if cloudRegion == "" {
					return nil, fmt.Errorf("--cloud-region is required when --cloud-provider is not %q", qcloudapi.HybridCloudProviderID)
				}
				cloudRegionPtr = &cloudRegion
			}

			resp, err := client.Booking().ListPackages(ctx, &bookingv1.ListPackagesRequest{
				AccountId:             accountID,
				CloudProviderId:       cloudProvider,
				CloudProviderRegionId: cloudRegionPtr,
				Statuses:              []bookingv1.PackageStatus{bookingv1.PackageStatus_PACKAGE_STATUS_ACTIVE},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list packages: %w", err)
			}

			return resp, nil
		},
		OutputTable: func(_ *cobra.Command, w io.Writer, resp *bookingv1.ListPackagesResponse) (output.TableRenderer, error) {
			t := output.NewTable[*bookingv1.Package](w)
			t.AddField("NAME", func(p *bookingv1.Package) string {
				return p.GetName()
			})
			t.AddField("ID", func(p *bookingv1.Package) string {
				return p.GetId()
			})
			t.AddField("TIER", func(p *bookingv1.Package) string {
				return output.PackageTier(p.GetTier())
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
			t.AddField("GPU", func(p *bookingv1.Package) string {
				if rc := p.GetResourceConfiguration(); rc != nil {
					if v := rc.GetGpu(); v != "" {
						return v
					}
				}
				return "n/a"
			})
			t.AddField("MULTI-AZ", func(p *bookingv1.Package) string {
				return output.BoolYesNo(p.GetMultiAz())
			})
			t.AddField("PRICE/HR", func(p *bookingv1.Package) string {
				return output.FormatMillicents(p.GetUnitIntPricePerHour(), p.GetCurrency())
			})
			t.SetItems(resp.GetItems())
			return t, nil
		},
	}.CobraCommand(s)

	cmd.Flags().String("cloud-provider", "", "Cloud provider ID (required)")
	cmd.Flags().String("cloud-region", "", "Cloud provider region ID (required for non-hybrid providers)")
	_ = cmd.MarkFlagRequired("cloud-provider")

	_ = cmd.RegisterFlagCompletionFunc("cloud-provider", completion.CloudProviderCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("cloud-region", completion.CloudRegionCompletion(s))
	return cmd
}
