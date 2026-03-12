package cluster

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newDescribeCommand(s *state.State) *cobra.Command {
	return base.DescribeCmd[*clusterv1.Cluster]{
		Use:   "describe <cluster-id>",
		Short: "Describe a cluster",
		Args:  util.ExactArgs(1, "a cluster ID"),
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*clusterv1.Cluster, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			resp, err := client.Cluster().GetCluster(ctx, &clusterv1.GetClusterRequest{
				AccountId: accountID,
				ClusterId: args[0],
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get cluster: %w", err)
			}

			return resp.GetCluster(), nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, cluster *clusterv1.Cluster) error {
			fmt.Fprintf(w, "ID:       %s\n", cluster.GetId())
			fmt.Fprintf(w, "Name:     %s\n", cluster.GetName())
			if cluster.GetState() != nil {
				fmt.Fprintf(w, "Status:   %s\n", phaseString(cluster.GetState().GetPhase()))
			}
			if cluster.GetConfiguration() != nil {
				cfg := cluster.GetConfiguration()
				fmt.Fprintf(w, "Version:  %s\n", cfg.GetVersion())
				fmt.Fprintf(w, "Nodes:    %d\n", cfg.GetNumberOfNodes())
				fmt.Fprintf(w, "Package:  %s\n", cfg.GetPackageId())
			}
			fmt.Fprintf(w, "Cloud:    %s\n", cluster.GetCloudProviderId())
			fmt.Fprintf(w, "Region:   %s\n", cluster.GetCloudProviderRegionId())
			if cluster.GetCreatedAt() != nil {
				t := cluster.GetCreatedAt().AsTime()
				fmt.Fprintf(w, "Created:  %s  (%s)\n", output.HumanTime(t), output.FullDateTime(t))
			}
			if labels := cluster.GetLabels(); len(labels) > 0 {
				fmt.Fprintf(w, "Labels:   ")
				for i, kv := range labels {
					if i > 0 {
						fmt.Fprintf(w, "          ")
					}
					fmt.Fprintf(w, "%s=%s\n", kv.GetKey(), kv.GetValue())
				}
			}

			if st := cluster.GetState(); st != nil {
				if ep := st.GetEndpoint(); ep != nil {
					fmt.Fprintln(w)
					fmt.Fprintf(w, "Endpoint:   %s\n", ep.GetUrl())
					fmt.Fprintf(w, "REST Port:  %d\n", ep.GetRestPort())
					fmt.Fprintf(w, "gRPC Port:  %d\n", ep.GetGrpcPort())
				}

				if res := st.GetResources(); res != nil {
					fmt.Fprintln(w)
					fmt.Fprintln(w, "Resources (per node):")
					if disk := res.GetDisk(); disk != nil {
						fmt.Fprintf(w, "  Disk:  %s base, %s available\n",
							formatGiB(disk.GetBase()), formatGiB(disk.GetAvailable()))
					}
					if ram := res.GetRam(); ram != nil {
						fmt.Fprintf(w, "  RAM:   %s base, %s reserved, %s available\n",
							formatGiB(ram.GetBase()), formatGiB(ram.GetReserved()), formatGiB(ram.GetAvailable()))
					}
					if cpu := res.GetCpu(); cpu != nil {
						fmt.Fprintf(w, "  CPU:   %s base, %s reserved, %s available\n",
							formatMillicores(cpu.GetBase()), formatMillicores(cpu.GetReserved()), formatMillicores(cpu.GetAvailable()))
					}
				}

				if nodes := st.GetNodes(); len(nodes) > 0 {
					fmt.Fprintln(w)
					fmt.Fprintln(w, "Nodes:")
					for _, n := range nodes {
						started := ""
						if n.GetStartedAt() != nil {
							started = "started " + output.HumanTime(n.GetStartedAt().AsTime())
						}
						fmt.Fprintf(w, "  %-40s  %-12s  %-10s  %s\n",
							n.GetName(), nodeStateString(n.GetState()), n.GetVersion(), started)
					}
				}
			}

			return nil
		},
		ValidArgsFunction: clusterIDCompletion(s),
	}.CobraCommand(s)
}
