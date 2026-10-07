package quota

import (
	"fmt"
	"io"
	"strconv"

	"github.com/spf13/cobra"

	serverlessquotav1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/quota/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/resource"
	"github.com/qdrant/qcloud-cli/internal/state"
)

// NewCommand creates the serverless quota command.
func NewCommand(s *state.State) *cobra.Command {
	return base.DescribeCmd[*serverlessquotav1.AccountQuota]{
		Use:   "quota",
		Short: "Show the serverless quota of the account",
		Long: `Show the serverless quota of the current account.

The quota caps how many spaces the account can create and sets the platform
limits that are copied into every space it owns: the maximum number of
collections per space, the maximum size of a collection and the maximum number
of search workers per collection. A limit of 0 is shown as "unlimited".`,
		Example: `# Show the serverless quota
qcloud serverless quota

# Output as JSON
qcloud serverless quota --json`,
		Args: cobra.NoArgs,
		Fetch: func(s *state.State, cmd *cobra.Command, _ []string) (*serverlessquotav1.AccountQuota, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			resp, err := client.ServerlessQuota().GetQuotas(ctx, &serverlessquotav1.GetQuotasRequest{
				AccountId: accountID,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get serverless quota: %w", err)
			}

			return resp.GetQuota(), nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, q *serverlessquotav1.AccountQuota) error {
			maxSpaces := output.UnlimitedIfZero(strconv.FormatUint(q.GetMaxSpaces(), 10), q.GetMaxSpaces())
			fmt.Fprintf(w, "Spaces:                            %d / %s\n", q.GetUsedSpaces(), maxSpaces)
			fmt.Fprintf(w, "Max Collections per Space:         %s\n",
				output.UnlimitedIfZero(strconv.FormatUint(q.GetMaxCollectionsPerSpace(), 10), q.GetMaxCollectionsPerSpace()))
			fmt.Fprintf(w, "Max Size per Collection:           %s\n",
				output.UnlimitedIfZero(resource.ByteQuantity(q.GetPlatformMaxSizePerCollection()).String(), q.GetPlatformMaxSizePerCollection()))
			fmt.Fprintf(w, "Max Search Workers per Collection: %s\n",
				output.UnlimitedIfZero(strconv.FormatUint(q.GetPlatformMaxWorkersPerCollection(), 10), q.GetPlatformMaxWorkersPerCollection()))
			return nil
		},
	}.CobraCommand(s)
}
