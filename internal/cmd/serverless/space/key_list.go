package space

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	spaceauthv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/auth/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newKeyListCommand(s *state.State) *cobra.Command {
	return base.ListCmd[*spaceauthv1.ListSpaceApiKeysResponse]{
		Use:   "list <space-id>",
		Short: "List API keys for a space",
		Long: `List the API keys of a serverless space.

The ACCESS column shows the global access type of a key, or the per-collection
access rules as "collection:ACCESS" pairs. Secret key values are never listed;
use the POSTFIX column to identify a key.`,
		Example: `# List API keys for a space
qcloud serverless space key list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Output as JSON
qcloud serverless space key list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --json`,
		Args: util.ExactArgs(1, "a space ID"),
		Fetch: func(s *state.State, cmd *cobra.Command) (*spaceauthv1.ListSpaceApiKeysResponse, error) {
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

			resp, err := client.ServerlessSpaceApiKey().ListSpaceApiKeys(ctx, &spaceauthv1.ListSpaceApiKeysRequest{
				AccountId: accountID,
				SpaceId:   spaceID,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list API keys: %w", err)
			}

			return resp, nil
		},
		OutputTable: func(_ *cobra.Command, w io.Writer, resp *spaceauthv1.ListSpaceApiKeysResponse) (output.TableRenderer, error) {
			t := output.NewTable[*spaceauthv1.SpaceApiKey](w)
			t.AddField("ID", func(v *spaceauthv1.SpaceApiKey) string {
				return v.GetId()
			})
			t.AddField("NAME", func(v *spaceauthv1.SpaceApiKey) string {
				return v.GetName()
			})
			t.AddField("PHASE", func(v *spaceauthv1.SpaceApiKey) string {
				return output.SpaceApiKeyPhase(v.GetState().GetPhase())
			})
			t.AddField("ACCESS", func(v *spaceauthv1.SpaceApiKey) string {
				return keyAccessSummary(v.GetAccessRules())
			})
			t.AddField("POSTFIX", func(v *spaceauthv1.SpaceApiKey) string {
				return v.GetPostfix()
			})
			t.AddField("EXPIRES", func(v *spaceauthv1.SpaceApiKey) string {
				if v.GetExpiresAt() != nil {
					return output.FullDateTime(v.GetExpiresAt().AsTime())
				}

				return ""
			})
			t.AddField("CREATED", func(v *spaceauthv1.SpaceApiKey) string {
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
}

// keyAccessSummary renders the access rules of a key as a short string, e.g.
// "MANAGE" or "products:READ_ONLY, orders:READ_WRITE".
func keyAccessSummary(rules []*spaceauthv1.AccessRule) string {
	parts := make([]string, 0, len(rules))
	for _, r := range rules {
		if g := r.GetGlobalAccess(); g != nil {
			parts = append(parts, output.SpaceGlobalAccessType(g.GetAccessType()))
			continue
		}

		if c := r.GetCollectionAccess(); c != nil {
			parts = append(parts, c.GetCollectionName()+":"+output.SpaceCollectionAccessType(c.GetAccessType()))
		}
	}

	return strings.Join(parts, ", ")
}
