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

func newDatasetDescribeCommand(s *state.State) *cobra.Command {
	return base.DescribeCmd[*spacedatasetv1.SampleDataset]{
		Use:   "describe <space-id> <dataset-id>",
		Short: "Describe a sample dataset",
		Long: `Describe a sample dataset that can be loaded into a serverless space.

Shows the size of the dataset and the configuration of the collection a load
creates: its dense and sparse vectors, and the embedding model that made each
of them. Use the same model to make query vectors.`,
		Example: `# Describe a sample dataset
qcloud serverless space dataset describe 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 midjourney-styles

# Output as JSON
qcloud serverless space dataset describe 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 midjourney-styles --json`,
		Args:              util.ExactArgs(2, "a space ID and a sample dataset ID"),
		ValidArgsFunction: completion.SpaceSampleDatasetIDCompletion(s),
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*spacedatasetv1.SampleDataset, error) {
			// There is no Get RPC: the catalog holds only a few datasets and
			// is always listed in full.
			resp, err := listSampleDatasets(s, cmd, args[0])
			if err != nil {
				return nil, err
			}

			for _, d := range resp.GetItems() {
				if d.GetId() == args[1] {
					return d, nil
				}
			}

			return nil, fmt.Errorf("sample dataset %q is not available in space %s; see \"qcloud serverless space dataset list %s\"",
				args[1], args[0], args[0])
		},
		PrintText: func(_ *cobra.Command, w io.Writer, d *spacedatasetv1.SampleDataset) error {
			fmt.Fprintf(w, "ID:            %s\n", d.GetId())
			fmt.Fprintf(w, "Name:          %s\n", d.GetName())
			fmt.Fprintf(w, "Version:       %s\n", d.GetVersion())
			fmt.Fprintf(w, "Points:        %d\n", d.GetPointCount())
			fmt.Fprintf(w, "Size:          %s\n", output.Bytes(d.GetSizeBytes()))
			fmt.Fprintf(w, "Indexed Size:  %s\n", output.Bytes(d.GetIndexedSizeBytes()))
			if d.GetDescription() != "" {
				fmt.Fprintf(w, "Description:   %s\n", d.GetDescription())
			}

			cfg := d.GetCollectionConfig()
			if dense := cfg.GetDenseVectors(); len(dense) > 0 {
				fmt.Fprintln(w)
				fmt.Fprintln(w, "Dense Vectors:")
				t := output.NewTable[string](w)
				t.AddField("NAME", sampleDatasetVectorLabel)
				t.AddField("SIZE", func(name string) string {
					return strconv.FormatUint(dense[name].GetSize(), 10)
				})
				t.AddField("DISTANCE", func(name string) string {
					return output.SampleDatasetDistance(dense[name].GetDistance())
				})
				t.AddField("MULTIVECTOR", func(name string) string {
					return output.BoolYesNo(dense[name].GetMultivector())
				})
				t.AddField("MODEL", func(name string) string {
					return dense[name].GetModel()
				})
				t.SetItems(sortedKeys(dense))
				t.Render()
			}

			if sparse := cfg.GetSparseVectors(); len(sparse) > 0 {
				fmt.Fprintln(w)
				fmt.Fprintln(w, "Sparse Vectors:")
				t := output.NewTable[string](w)
				t.AddField("NAME", sampleDatasetVectorLabel)
				t.AddField("IDF", func(name string) string {
					return output.BoolYesNo(sparse[name].GetUseIdf())
				})
				t.AddField("MODEL", func(name string) string {
					return sparse[name].GetModel()
				})
				t.SetItems(sortedKeys(sparse))
				t.Render()
			}

			return nil
		},
	}.CobraCommand(s)
}
