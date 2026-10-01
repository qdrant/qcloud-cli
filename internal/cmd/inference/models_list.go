package inference

import (
	"cmp"
	"fmt"
	"io"
	"slices"

	"github.com/spf13/cobra"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newModelsListCommand(s *state.State) *cobra.Command {
	cmd := base.ListCmd[*bookingv1.ListGlobalInferenceModelsResponse]{
		Use:   "list",
		Short: "List available inference models",
		Long: `List the inference models globally available for a cloud provider and region.

The listing is global, not account-specific: it shows every model Qdrant Cloud
offers in that region, together with the vector type and modality it produces,
its output dimensionality and per-request token limit, and its price per one
million processed tokens. External models are served by a third-party vendor and
require that vendor's API key to be configured on the cluster.

Inference is only offered on managed cloud, so unlike "qcloud package
list" this command cannot be used with hybrid cloud.`,
		Example: `# List inference models available on AWS in eu-central-1
qcloud inference models list --cloud-provider aws --cloud-region eu-central-1

# List inference models as JSON
qcloud inference models list --cloud-provider gcp --cloud-region us-east4 --json`,
		Fetch: func(s *state.State, cmd *cobra.Command) (*bookingv1.ListGlobalInferenceModelsResponse, error) {
			ctx := cmd.Context()
			client, err := s.UnAuthenticatedClient(ctx)
			if err != nil {
				return nil, err
			}

			cloudProvider, _ := cmd.Flags().GetString("cloud-provider")
			cloudRegion, _ := cmd.Flags().GetString("cloud-region")

			resp, err := client.Booking().ListGlobalInferenceModels(ctx, &bookingv1.ListGlobalInferenceModelsRequest{
				CloudProviderId:       cloudProvider,
				CloudProviderRegionId: cloudRegion,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list inference models: %w", err)
			}

			// The API returns models in no particular order.
			slices.SortFunc(resp.GetItems(), func(a, b *bookingv1.InferenceModel) int {
				return cmp.Compare(a.GetName(), b.GetName())
			})

			return resp, nil
		},
		OutputTable: func(_ *cobra.Command, w io.Writer, resp *bookingv1.ListGlobalInferenceModelsResponse) (output.TableRenderer, error) {
			t := output.NewTable[*bookingv1.InferenceModel](w)
			t.AddField("NAME", func(m *bookingv1.InferenceModel) string {
				return m.GetName()
			})
			t.AddField("TITLE", func(m *bookingv1.InferenceModel) string {
				return m.GetTitle()
			})
			t.AddField("VENDOR", func(m *bookingv1.InferenceModel) string {
				return m.GetVendor()
			})
			t.AddField("VECTOR-TYPE", func(m *bookingv1.InferenceModel) string {
				return output.VectorType(m.GetVectorType())
			})
			t.AddField("MODALITY", func(m *bookingv1.InferenceModel) string {
				return output.ModelModality(m.GetModality())
			})
			t.AddField("DIMENSIONS", func(m *bookingv1.InferenceModel) string {
				return output.OptionalValue(m.Dimensionality, "n/a")
			})
			t.AddField("MAX-TOKENS", func(m *bookingv1.InferenceModel) string {
				return output.OptionalValue(m.MaxTokensPerRequest, "n/a")
			})
			t.AddField("EXTERNAL", func(m *bookingv1.InferenceModel) string {
				return output.BoolYesNo(m.GetIsExternal())
			})
			t.AddField("PRICE/1M-TOKENS", func(m *bookingv1.InferenceModel) string {
				return output.FormatMillicents(int32(m.GetUnitIntPrice()), "USD")
			})
			t.SetItems(resp.GetItems())
			return t, nil
		},
	}.CobraCommand(s)

	cmd.Flags().String("cloud-provider", "", "Cloud provider ID (required)")
	cmd.Flags().String("cloud-region", "", "Cloud provider region ID (required)")
	_ = cmd.MarkFlagRequired("cloud-provider")
	_ = cmd.MarkFlagRequired("cloud-region")

	_ = cmd.RegisterFlagCompletionFunc("cloud-provider", completion.CloudProviderCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("cloud-region", completion.CloudRegionCompletion(s))
	return cmd
}
