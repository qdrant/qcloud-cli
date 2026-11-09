package completion

import (
	"cmp"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// ClusterNodeCompletion completes a cluster ID, then the numbers of its nodes.
func ClusterNodeCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	clusterIDs := ClusterIDCompletion(s)
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		switch len(args) {
		case 0:
			return clusterIDs(cmd, args, toComplete)
		case 1:
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}

			resp, err := client.Cluster().GetCluster(ctx, &clusterv1.GetClusterRequest{AccountId: accountID, ClusterId: args[0]})
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}

			nodes := make([]string, 0, len(resp.GetCluster().GetState().GetNodes()))
			for _, n := range resp.GetCluster().GetState().GetNodes() {
				nodes = append(nodes, NodeNumber(args[0], n.GetName()))
			}

			slices.SortFunc(nodes, CompareNodes)

			return nodes, cobra.ShellCompDirectiveNoFileComp
		default:
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	}
}

// NodeNumber returns the short name of a node, such as "0", from its full ID.
func NodeNumber(clusterID, nodeID string) string {
	return strings.TrimPrefix(nodeID, "qdrant-"+clusterID+"-")
}

// CompareNodes orders node numbers numerically, falling back to text.
func CompareNodes(a, b string) int {
	an, aErr := strconv.Atoi(a)
	bn, bErr := strconv.Atoi(b)
	if aErr == nil && bErr == nil {
		return cmp.Compare(an, bn)
	}

	return cmp.Compare(a, b)
}
