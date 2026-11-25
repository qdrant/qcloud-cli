package space

import (
	"fmt"
	"io"
	"strconv"

	"github.com/spf13/cobra"

	spacedatasetv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/dataset/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newDatasetListCommand(s *state.State) *cobra.Command {
	return base.ListCmd[*spacedatasetv1.ListSampleDatasetsResponse]{
		Use:   "list <space-id>",
		Short: "List the sample datasets that can be loaded into a space",
		Long: `List the sample datasets that can be loaded into a serverless space.

Only datasets available in the region of the space are listed. SIZE and
INDEXED SIZE are the storage the loaded collection uses; they count against the
collection size limits of the space. The VECTORS column lists the vector names
of the collection a load creates; "default" is the unnamed vector. Use
"dataset describe" to see the vector sizes, distances and embedding models.`,
		Example: `# List the sample datasets that can be loaded into a space
qcloud serverless space dataset list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Output as JSON
qcloud serverless space dataset list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --json`,
		Args: util.ExactArgs(1, "a space ID"),
		Fetch: func(s *state.State, cmd *cobra.Command) (*spacedatasetv1.ListSampleDatasetsResponse, error) {
			return listSampleDatasets(s, cmd, cmd.Flags().Arg(0))
		},
		OutputTable: func(_ *cobra.Command, w io.Writer, resp *spacedatasetv1.ListSampleDatasetsResponse) (output.TableRenderer, error) {
			t := output.NewTable[*spacedatasetv1.SampleDataset](w)
			t.AddField("ID", func(v *spacedatasetv1.SampleDataset) string {
				return v.GetId()
			})
			t.AddField("NAME", func(v *spacedatasetv1.SampleDataset) string {
				return v.GetName()
			})
			t.AddField("POINTS", func(v *spacedatasetv1.SampleDataset) string {
				return strconv.FormatUint(v.GetPointCount(), 10)
			})
			t.AddField("SIZE", func(v *spacedatasetv1.SampleDataset) string {
				return output.Bytes(v.GetSizeBytes())
			})
			t.AddField("INDEXED SIZE", func(v *spacedatasetv1.SampleDataset) string {
				return output.Bytes(v.GetIndexedSizeBytes())
			})
			t.AddField("VECTORS", func(v *spacedatasetv1.SampleDataset) string {
				return sampleDatasetVectorNames(v.GetCollectionConfig())
			})
			t.AddField("VERSION", func(v *spacedatasetv1.SampleDataset) string {
				return v.GetVersion()
			})
			t.SetItems(resp.GetItems())
			return t, nil
		},
		ValidArgsFunction: completion.SpaceIDCompletion(s),
	}.CobraCommand(s)
}

// listSampleDatasets returns the sample datasets that can be loaded into the space.
func listSampleDatasets(s *state.State, cmd *cobra.Command, spaceID string) (*spacedatasetv1.ListSampleDatasetsResponse, error) {
	ctx := cmd.Context()
	client, err := s.Client(ctx)
	if err != nil {
		return nil, err
	}

	accountID, err := s.AccountID()
	if err != nil {
		return nil, err
	}

	resp, err := client.ServerlessSampleDataset().ListSampleDatasets(ctx, &spacedatasetv1.ListSampleDatasetsRequest{
		AccountId: accountID,
		SpaceId:   spaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list sample datasets: %w", err)
	}

	return resp, nil
}
