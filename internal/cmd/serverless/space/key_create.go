package space

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"

	spaceauthv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/auth/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newKeyCreateCommand(s *state.State) *cobra.Command {
	return base.CreateCmd[*spaceauthv1.SpaceApiKey]{
		Long: `Create an API key for a serverless space.

A key grants either global access to the whole space (--access-type) or access to
individual collections (--collection, repeatable). The two kinds of rules cannot be
combined in one key. When neither flag is given, the server assigns global manage
access.

An expiration date given with --expires is inclusive: the key stays valid until
the end of that day (23:59:59 UTC).

The secret key value is printed only once. Store it securely; it cannot be
retrieved later. If --wait fails after the key was created, the secret is still
printed before the error is returned.`,
		Example: `# Create an API key with manage access
qcloud serverless space key create 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --name my-key

# Create a read-only key that expires at the end of the year
qcloud serverless space key create 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --name read-key --access-type read-only --expires 2026-12-31

# Create a key scoped to two collections
qcloud serverless space key create 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --name app-key --collection products=read-write --collection reviews=read-only

# Create a key and wait for it to become ready
qcloud serverless space key create 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --name my-key --wait`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "create <space-id>",
				Short: "Create an API key for a space",
				Args:  util.ExactArgs(1, "a space ID"),
			}
			cmd.Flags().String("name", "", "Name of the API key (required)")
			cmd.Flags().String("access-type", "", "Global access type: manage, read-only or metrics-read-only (default: server assigns manage)")
			cmd.Flags().StringArray("collection", nil, "Collection access rule as 'name=read-only|read-write'; can be specified multiple times")
			cmd.Flags().String("expires", "", "Expiration date in YYYY-MM-DD format; the key is valid until the end of that day (UTC)")
			util.AddWaitFlags(cmd, "the API key to become ready", time.Minute, time.Second)
			_ = cmd.MarkFlagRequired("name")
			cmd.MarkFlagsMutuallyExclusive("access-type", "collection")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) (*spaceauthv1.SpaceApiKey, error) {
			spaceID := args[0]

			name, _ := cmd.Flags().GetString("name")
			accessType, _ := cmd.Flags().GetString("access-type")
			collections, _ := cmd.Flags().GetStringArray("collection")
			expiresStr, _ := cmd.Flags().GetString("expires")

			rules, err := parseKeyAccessRules(accessType, collections)
			if err != nil {
				return nil, err
			}

			var expiresAt *timestamppb.Timestamp
			if expiresStr != "" {
				t, err := util.ParseDateEndOfDay(expiresStr)
				if err != nil {
					return nil, fmt.Errorf("invalid --expires %q: must be in YYYY-MM-DD format", expiresStr)
				}

				expiresAt = timestamppb.New(t)
			}

			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			resp, err := client.ServerlessSpaceApiKey().CreateSpaceApiKey(ctx, &spaceauthv1.CreateSpaceApiKeyRequest{
				SpaceApiKey: &spaceauthv1.SpaceApiKey{
					AccountId:   accountID,
					SpaceId:     spaceID,
					Name:        name,
					AccessRules: rules,
					ExpiresAt:   expiresAt,
				},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create API key: %w", err)
			}

			created := resp.GetSpaceApiKey()

			wait, _ := cmd.Flags().GetBool("wait")
			if !wait {
				return created, nil
			}

			waitTimeout, _ := cmd.Flags().GetDuration("wait-timeout")
			pollInterval, _ := cmd.Flags().GetDuration("wait-poll-interval")

			fmt.Fprintf(cmd.ErrOrStderr(), "API key %s created, waiting for it to become ready...\n", created.GetId())
			ready, err := waitForSpaceApiKeyReady(ctx, client.ServerlessSpaceApiKey(), cmd.ErrOrStderr(),
				accountID, spaceID, created.GetId(), waitTimeout, pollInterval)
			if err != nil {
				// The secret is only returned by the create call, so print it
				// before failing; otherwise it would be lost.
				if s.Config.JSONOutput() {
					_ = output.PrintJSON(cmd.OutOrStdout(), created)
				} else {
					output.CreatedAPIKey(cmd.OutOrStdout(), created)
				}

				return nil, fmt.Errorf("API key %s was created but did not become ready: %w", created.GetId(), err)
			}

			// The secret is only returned by the create call, never by list.
			ready.Key = created.GetKey()
			return ready, nil
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, key *spaceauthv1.SpaceApiKey) {
			output.CreatedAPIKey(out, key)
		},
		ValidArgsFunction: completion.SpaceIDCompletion(s),
	}.CobraCommand(s)
}

// parseKeyAccessRules builds the access rules from the --access-type and
// --collection flag values. It returns nil when neither is set.
func parseKeyAccessRules(accessType string, collections []string) ([]*spaceauthv1.AccessRule, error) {
	if accessType != "" {
		var at spaceauthv1.GlobalAccessRuleAccessType
		switch accessType {
		case "manage":
			at = spaceauthv1.GlobalAccessRuleAccessType_GLOBAL_ACCESS_RULE_ACCESS_TYPE_MANAGE
		case "read-only":
			at = spaceauthv1.GlobalAccessRuleAccessType_GLOBAL_ACCESS_RULE_ACCESS_TYPE_READ_ONLY
		case "metrics-read-only":
			at = spaceauthv1.GlobalAccessRuleAccessType_GLOBAL_ACCESS_RULE_ACCESS_TYPE_METRICS_READ_ONLY
		default:
			return nil, fmt.Errorf("invalid --access-type %q: must be manage, read-only or metrics-read-only", accessType)
		}

		return []*spaceauthv1.AccessRule{{
			Scope: &spaceauthv1.AccessRule_GlobalAccess{
				GlobalAccess: &spaceauthv1.GlobalAccessRule{AccessType: at},
			},
		}}, nil
	}

	rules := make([]*spaceauthv1.AccessRule, 0, len(collections))
	for _, raw := range collections {
		idx := strings.LastIndex(raw, "=")
		if idx <= 0 {
			return nil, fmt.Errorf("invalid --collection %q: must be in 'name=read-only|read-write' format", raw)
		}

		name, access := raw[:idx], raw[idx+1:]

		var at spaceauthv1.CollectionAccessRuleAccessType
		switch access {
		case "read-only":
			at = spaceauthv1.CollectionAccessRuleAccessType_COLLECTION_ACCESS_RULE_ACCESS_TYPE_READ_ONLY
		case "read-write":
			at = spaceauthv1.CollectionAccessRuleAccessType_COLLECTION_ACCESS_RULE_ACCESS_TYPE_READ_WRITE
		default:
			return nil, fmt.Errorf("invalid --collection %q: access must be read-only or read-write", raw)
		}

		rules = append(rules, &spaceauthv1.AccessRule{
			Scope: &spaceauthv1.AccessRule_CollectionAccess{
				CollectionAccess: &spaceauthv1.CollectionAccessRule{CollectionName: name, AccessType: at},
			},
		})
	}

	return rules, nil
}
