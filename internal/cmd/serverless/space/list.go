package space

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	spacev1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newListCommand(s *state.State) *cobra.Command {
	cmd := base.ListCmd[*spacev1.ListSpacesResponse]{
		Use:   "list",
		Short: "List all spaces",
		Long: `List all serverless spaces in the current account.

By default, all spaces are fetched automatically across multiple pages. Use
--page-size and --page-token for manual pagination; the next page token is
included in the JSON output when more pages exist.

Use --cloud-region to only list the spaces hosted in a specific region.`,
		Example: `# List all spaces
qcloud serverless space list

# List spaces in a specific region
qcloud serverless space list --cloud-region aws-eu-central-1

# List spaces in JSON format
qcloud serverless space list --json

# Manual pagination
qcloud serverless space list --page-size 10`,
		Fetch: func(s *state.State, cmd *cobra.Command) (*spacev1.ListSpacesResponse, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			newRequest := func() *spacev1.ListSpacesRequest {
				req := &spacev1.ListSpacesRequest{AccountId: accountID}
				if cmd.Flags().Changed("cloud-region") {
					region, _ := cmd.Flags().GetString("cloud-region")
					req.CloudRegionId = &region
				}

				return req
			}

			pageSizeChanged := cmd.Flags().Changed("page-size")
			pageTokenChanged := cmd.Flags().Changed("page-token")

			if !pageSizeChanged && !pageTokenChanged {
				// Auto-paginate: fetch all pages and return combined results.
				var allItems []*spacev1.Space
				var nextToken *string
				for {
					req := newRequest()
					req.PageToken = nextToken

					resp, err := client.ServerlessSpace().ListSpaces(ctx, req)
					if err != nil {
						return nil, fmt.Errorf("failed to list spaces: %w", err)
					}

					allItems = append(allItems, resp.GetItems()...)
					if resp.GetNextPageToken() == "" {
						break
					}

					nextToken = resp.NextPageToken
				}

				return &spacev1.ListSpacesResponse{Items: allItems}, nil
			}

			// Manual mode: single request with provided flags.
			req := newRequest()
			if pageSizeChanged {
				ps, _ := cmd.Flags().GetInt32("page-size")
				req.PageSize = &ps
			}

			if pageTokenChanged {
				pt, _ := cmd.Flags().GetString("page-token")
				req.PageToken = &pt
			}

			resp, err := client.ServerlessSpace().ListSpaces(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("failed to list spaces: %w", err)
			}

			return resp, nil
		},
		OutputTable: func(_ *cobra.Command, w io.Writer, resp *spacev1.ListSpacesResponse) (output.TableRenderer, error) {
			t := output.NewTable[*spacev1.Space](w)
			t.AddField("ID", func(v *spacev1.Space) string {
				return v.GetId()
			})
			t.AddField("NAME", func(v *spacev1.Space) string {
				return v.GetName()
			})
			t.AddField("PHASE", func(v *spacev1.Space) string {
				if v.GetState() != nil {
					return output.SpacePhase(v.GetState().GetPhase())
				}

				return ""
			})
			t.AddField("REGION", func(v *spacev1.Space) string {
				return v.GetCloudRegionId()
			})
			t.AddField("ENDPOINT", func(v *spacev1.Space) string {
				return v.GetState().GetEndpoint().GetUrl()
			})
			t.AddField("CREATED", func(v *spacev1.Space) string {
				if v.GetCreatedAt() != nil {
					return output.HumanTime(v.GetCreatedAt().AsTime())
				}

				return ""
			})
			t.SetItems(resp.GetItems())
			return t, nil
		},
	}.CobraCommand(s)

	cmd.Flags().Int32("page-size", 0, "Maximum number of spaces to return per page (manual pagination mode)")
	cmd.Flags().String("page-token", "", "Page token from a previous response to resume from (manual pagination mode)")
	cmd.Flags().String("cloud-region", "", "Filter by cloud region ID")

	return cmd
}
