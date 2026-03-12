package cluster

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newVersionCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Manage Qdrant versions",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newVersionListCommand(s))
	return cmd
}

func newVersionListCommand(s *state.State) *cobra.Command {
	return base.ListCmd[*clusterv1.ListQdrantReleasesResponse]{
		Use:   "list",
		Short: "List available Qdrant versions",
		Fetch: func(s *state.State, cmd *cobra.Command) (*clusterv1.ListQdrantReleasesResponse, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}
			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}
			resp, err := client.Cluster().ListQdrantReleases(ctx, &clusterv1.ListQdrantReleasesRequest{
				AccountId: accountID,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list versions: %w", err)
			}
			return resp, nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, resp *clusterv1.ListQdrantReleasesResponse) error {
			t := output.NewTable[*clusterv1.QdrantRelease](w)
			t.AddField("VERSION", func(r *clusterv1.QdrantRelease) string { return r.GetVersion() })
			t.AddField("DEFAULT", func(r *clusterv1.QdrantRelease) string { return boolToMark(r.GetDefault()) })
			t.AddField("END OF LIFE", func(r *clusterv1.QdrantRelease) string { return boolToMark(r.GetEndOfLife()) })
			t.AddField("UNAVAILABLE", func(r *clusterv1.QdrantRelease) string { return boolToMark(r.GetUnavailable()) })
			t.AddField("REMARKS", func(r *clusterv1.QdrantRelease) string { return r.GetRemarks() })
			t.Write(resp.GetItems())
			return nil
		},
	}.CobraCommand(s)
}
