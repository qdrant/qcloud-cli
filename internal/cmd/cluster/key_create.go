package cluster

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"

	clusterauthv2 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/auth/v2"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newKeyCreateCommand(s *state.State) *cobra.Command {
	return base.CreateCmd[*clusterauthv2.DatabaseApiKey]{
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "create <cluster-id>",
				Short: "Create an API key for a cluster",
				Args:  util.ExactArgs(1, "a cluster ID"),
			}
			cmd.Flags().String("name", "", "Name of the API key (required)")
			cmd.Flags().String("access-type", "", "Access type: manage or read-only (default: server assigns manage)")
			cmd.Flags().String("expires", "", "Expiration date in YYYY-MM-DD format")
			_ = cmd.MarkFlagRequired("name")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) (*clusterauthv2.DatabaseApiKey, error) {
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

			name, _ := cmd.Flags().GetString("name")
			accessType, _ := cmd.Flags().GetString("access-type")
			expiresStr, _ := cmd.Flags().GetString("expires")

			key := &clusterauthv2.DatabaseApiKey{
				AccountId: accountID,
				ClusterId: clusterID,
				Name:      name,
			}

			if accessType != "" {
				var at clusterauthv2.GlobalAccessRuleAccessType
				switch accessType {
				case "manage":
					at = clusterauthv2.GlobalAccessRuleAccessType_GLOBAL_ACCESS_RULE_ACCESS_TYPE_MANAGE
				case "read-only":
					at = clusterauthv2.GlobalAccessRuleAccessType_GLOBAL_ACCESS_RULE_ACCESS_TYPE_READ_ONLY
				default:
					return nil, fmt.Errorf("invalid --access-type %q: must be manage or read-only", accessType)
				}
				key.AccessRules = []*clusterauthv2.AccessRule{
					{
						Scope: &clusterauthv2.AccessRule_GlobalAccess{
							GlobalAccess: &clusterauthv2.GlobalAccessRule{
								AccessType: at,
							},
						},
					},
				}
			}

			if expiresStr != "" {
				t, err := time.Parse("2006-01-02", expiresStr)
				if err != nil {
					return nil, fmt.Errorf("invalid --expires %q: must be in YYYY-MM-DD format", expiresStr)
				}
				key.ExpiresAt = timestamppb.New(t)
			}

			resp, err := client.DatabaseApiKey().CreateDatabaseApiKey(ctx, &clusterauthv2.CreateDatabaseApiKeyRequest{
				DatabaseApiKey: key,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create API key: %w", err)
			}

			return resp.GetDatabaseApiKey(), nil
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, key *clusterauthv2.DatabaseApiKey) {
			fmt.Fprintf(out, "API key %s (%s) created.\n", key.GetId(), key.GetName())
			if k := key.GetKey(); k != "" {
				fmt.Fprintln(out, "")
				fmt.Fprintln(out, "Save this key now — it will not be shown again:")
				fmt.Fprintf(out, "  %s\n", k)
			}
		},
		ValidArgsFunction: completion.ClusterIDCompletion(s),
	}.CobraCommand(s)
}
