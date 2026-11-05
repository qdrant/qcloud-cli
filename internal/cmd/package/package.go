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
		Long: `List the packages available for creating a cluster in a cloud provider and region.

A package is a configuration (CPU, RAM, disk, and optionally GPU) for a cluster
node, with an hourly price. When you are authenticated, the packages
available to that account are listed, including multi-AZ packages and those of
the account's tier. Without credentials, the public standard-tier packages are
listed.`,
		Example: `# List packages for a cloud provider and region
qcloud package list --cloud-provider aws --cloud-region eu-central-1

# List packages for a hybrid cloud provider (no region required)
qcloud package list --cloud-provider hybrid`,
		Fetch: func(s *state.State, cmd *cobra.Command) (*bookingv1.ListPackagesResponse, error) {
			ctx := cmd.Context()
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

			// The public list only has standard-tier packages and misses
			// multi-AZ ones, so prefer the account's list when credentials exist.
			hasCredentials, err := s.HasCredentials()
			if err != nil {
				return nil, err
			}

			if hasCredentials && s.Config.AccountID() != "" {
				client, err := s.Client(ctx)
				if err != nil {
					return nil, err
				}

				resp, err := client.Booking().ListPackages(ctx, &bookingv1.ListPackagesRequest{
					AccountId:             s.Config.AccountID(),
					CloudProviderId:       cloudProvider,
					CloudProviderRegionId: cloudRegionPtr,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list packages: %w", err)
				}

				return resp, nil
			}

			if hasCredentials {
				fmt.Fprintln(cmd.ErrOrStderr(), "Hint: no account ID configured, listing the public standard-tier packages. Set QDRANT_CLOUD_ACCOUNT_ID or use --account-id to list the packages and prices of your account.")
			}

			client, err := s.UnAuthenticatedClient(ctx)
			if err != nil {
				return nil, err
			}

			resp, err := client.Booking().ListGlobalPackages(ctx, &bookingv1.ListGlobalPackagesRequest{
				CloudProviderId:       cloudProvider,
				CloudProviderRegionId: cloudRegionPtr,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list packages: %w", err)
			}

			return &bookingv1.ListPackagesResponse{Items: resp.GetItems()}, nil
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
				return output.PackageRAM(p.GetResourceConfiguration())
			})
			t.AddField("CPU", func(p *bookingv1.Package) string {
				return output.PackageCPU(p.GetResourceConfiguration())
			})
			t.AddField("DISK", func(p *bookingv1.Package) string {
				return output.PackageDisk(p.GetResourceConfiguration())
			})
			t.AddField("GPU", func(p *bookingv1.Package) string {
				return output.PackageGPU(p.GetResourceConfiguration())
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
