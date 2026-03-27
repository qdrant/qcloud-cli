package cluster

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"

	clusterauthv2 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/auth/v2"
	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newKeyCreateCommand(s *state.State) *cobra.Command {
	return base.CreateCmd[*clusterauthv2.DatabaseApiKey]{
		Example: `# Create an API key
qcloud cluster key create 7b2ea926-724b-4de2-b73a-8675c42a6ebe --name my-key

# Create a read-only key with expiration
qcloud cluster key create 7b2ea926-724b-4de2-b73a-8675c42a6ebe \
  --name read-key --access-type read-only --expires 2025-12-31

# Create a key and wait for it to become active on the cluster
qcloud cluster key create 7b2ea926-724b-4de2-b73a-8675c42a6ebe \
  --name my-key --wait`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "create <cluster-id>",
				Short: "Create an API key for a cluster",
				Args:  util.ExactArgs(1, "a cluster ID"),
			}
			cmd.Flags().String("name", "", "Name of the API key (required)")
			cmd.Flags().String("access-type", "", "Access type: manage or read-only (default: server assigns manage)")
			cmd.Flags().String("expires", "", "Expiration date in YYYY-MM-DD format")
			cmd.Flags().Bool("wait", false, "Wait for the API key to become active on the cluster")
			cmd.Flags().Duration("wait-timeout", time.Minute, "Maximum time to wait for the API key to become active")
			cmd.Flags().Duration("wait-poll-interval", time.Second, "How often to probe the cluster endpoint")
			_ = cmd.Flags().MarkHidden("wait-poll-interval")
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

			wait, _ := cmd.Flags().GetBool("wait")
			if !wait {
				return resp.GetDatabaseApiKey(), nil
			}

			waitTimeout, _ := cmd.Flags().GetDuration("wait-timeout")
			pollInterval, _ := cmd.Flags().GetDuration("wait-poll-interval")

			fmt.Fprintf(cmd.ErrOrStderr(), "API key created, waiting for it to become active on the cluster...\n")
			probe := newKeyProbe(client.Cluster(), accountID, clusterID, resp.GetDatabaseApiKey().GetKey())
			if err := waitForKeyReady(ctx, cmd.ErrOrStderr(), probe, waitTimeout, pollInterval); err != nil {
				return nil, err
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

const defaultQdrantRESTPort = 6333

func newKeyProbe(
	clusterSvc clusterv1.ClusterServiceClient,
	accountID, clusterID, apiKey string,
) func(ctx context.Context) error {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
		},
	}

	return func(ctx context.Context) error {
		clusterResp, err := clusterSvc.GetCluster(ctx, &clusterv1.GetClusterRequest{
			AccountId: accountID,
			ClusterId: clusterID,
		})
		if err != nil {
			return fmt.Errorf("failed to get cluster: %w", err)
		}

		ep := clusterResp.GetCluster().GetState().GetEndpoint()
		if ep == nil || ep.GetUrl() == "" {
			return fmt.Errorf("cluster %s has no endpoint yet", clusterID)
		}

		port := ep.GetRestPort()
		if port == 0 {
			port = defaultQdrantRESTPort
		}
		endpointURL := fmt.Sprintf("%s:%d", ep.GetUrl(), port)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpointURL, nil)
		if err != nil {
			return err
		}
		req.Header.Set("api-key", apiKey)
		resp, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close() //nolint:errcheck // best-effort close on a read-only probe
		_, _ = io.Copy(io.Discard, resp.Body)
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		return fmt.Errorf("cluster endpoint responded with HTTP %d at %s", resp.StatusCode, endpointURL)
	}
}
