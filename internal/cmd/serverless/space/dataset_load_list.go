package space

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	spacedatasetv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/dataset/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newDatasetLoadListCommand(s *state.State) *cobra.Command {
	cmd := base.ListCmd[*spacedatasetv1.ListSampleDatasetLoadsResponse]{
		Use:   "list <space-id>",
		Short: "List the sample dataset loads of a space",
		Long: `List the sample dataset loads of a serverless space.

Only loads created in the last 30 days are listed. PROGRESS shows how much data
was copied while a load is running. Use "dataset load describe" to see why a
load failed.

All pages are fetched by default. Use --page-size and --page-token for manual
pagination.`,
		Example: `# List the sample dataset loads of a space
qcloud serverless space dataset load list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Output as JSON
qcloud serverless space dataset load list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --json`,
		Args: util.ExactArgs(1, "a space ID"),
		Fetch: func(s *state.State, cmd *cobra.Command) (*spacedatasetv1.ListSampleDatasetLoadsResponse, error) {
			spaceID := cmd.Flags().Arg(0)

			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			items, next, total, err := util.FetchPages(cmd, func(pageSize *int32, pageToken *string) ([]*spacedatasetv1.SampleDatasetLoad, string, *int32, error) {
				resp, err := client.ServerlessSampleDataset().ListSampleDatasetLoads(ctx, &spacedatasetv1.ListSampleDatasetLoadsRequest{
					AccountId: accountID,
					SpaceId:   spaceID,
					PageSize:  pageSize,
					PageToken: pageToken,
				})
				if err != nil {
					return nil, "", nil, fmt.Errorf("failed to list sample dataset loads: %w", err)
				}

				return resp.GetItems(), resp.GetNextPageToken(), resp.TotalSize, nil
			})
			if err != nil {
				return nil, err
			}

			return &spacedatasetv1.ListSampleDatasetLoadsResponse{Items: items, NextPageToken: next, TotalSize: total}, nil
		},
		OutputTable: func(_ *cobra.Command, w io.Writer, resp *spacedatasetv1.ListSampleDatasetLoadsResponse) (output.TableRenderer, error) {
			t := output.NewTable[*spacedatasetv1.SampleDatasetLoad](w)
			t.AddField("ID", func(v *spacedatasetv1.SampleDatasetLoad) string {
				return v.GetId()
			})
			t.AddField("DATASET", func(v *spacedatasetv1.SampleDatasetLoad) string {
				return v.GetSampleDatasetId()
			})
			t.AddField("VERSION", func(v *spacedatasetv1.SampleDatasetLoad) string {
				return v.GetSampleDatasetVersion()
			})
			t.AddField("COLLECTION", func(v *spacedatasetv1.SampleDatasetLoad) string {
				return v.GetCollectionName()
			})
			t.AddField("STATUS", func(v *spacedatasetv1.SampleDatasetLoad) string {
				return output.SampleDatasetLoadStatus(v.GetStatus())
			})
			t.AddField("PROGRESS", func(v *spacedatasetv1.SampleDatasetLoad) string {
				return v.GetStats().GetProgress()
			})
			t.AddField("DURATION", func(v *spacedatasetv1.SampleDatasetLoad) string {
				return formatOptionalDuration(v.GetStats().GetDuration())
			})
			t.AddField("CREATED", func(v *spacedatasetv1.SampleDatasetLoad) string {
				if v.GetCreatedAt() != nil {
					return output.HumanTime(v.GetCreatedAt().AsTime())
				}

				return ""
			})
			t.SetItems(resp.GetItems())
			return t, nil
		},
		ValidArgsFunction: completion.SpaceIDCompletion(s),
	}.CobraCommand(s)
	util.AddPaginationFlags(cmd, "loads")
	return cmd
}
