package cluster

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	clusterauthv2 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/auth/v2"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newKeyListCommand(s *state.State) *cobra.Command {
	return base.DescribeCmd[*clusterauthv2.ListDatabaseApiKeysResponse]{
		Use:   "list <cluster-id>",
		Short: "List API keys for a cluster",
		Example: `# List API keys for a cluster
qcloud cluster key list 7b2ea926-724b-4de2-b73a-8675c42a6ebe`,
		Args: util.ExactArgs(1, "a cluster ID"),
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*clusterauthv2.ListDatabaseApiKeysResponse, error) {
			clusterID := args[0]

			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			resp, err := client.DatabaseApiKey().ListDatabaseApiKeys(ctx, &clusterauthv2.ListDatabaseApiKeysRequest{
				AccountId: accountID,
				ClusterId: &clusterID,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list API keys: %w", err)
			}

			return resp, nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, resp *clusterauthv2.ListDatabaseApiKeysResponse) error {
			t := output.NewTable[*clusterauthv2.DatabaseApiKey](w)
			t.AddField("ID", func(v *clusterauthv2.DatabaseApiKey) string {
				return v.GetId()
			})
			t.AddField("NAME", func(v *clusterauthv2.DatabaseApiKey) string {
				return v.GetName()
			})
			t.AddField("POSTFIX", func(v *clusterauthv2.DatabaseApiKey) string {
				return v.GetPostfix()
			})
			t.AddField("CREATED", func(v *clusterauthv2.DatabaseApiKey) string {
				if v.GetCreatedAt() != nil {
					return output.HumanTime(v.GetCreatedAt().AsTime())
				}
				return ""
			})
			t.AddField("EXPIRES", func(v *clusterauthv2.DatabaseApiKey) string {
				if v.GetExpiresAt() != nil {
					return output.FullDateTime(v.GetExpiresAt().AsTime())
				}
				return ""
			})
			t.Write(resp.GetItems())
			return nil
		},
		ValidArgsFunction: completion.ClusterIDCompletion(s),
	}.CobraCommand(s)
}
