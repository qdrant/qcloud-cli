package cluster

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"
	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newCreateCommand(s *state.State) *cobra.Command {
	cmd := base.CreateCmd[*clusterv1.Cluster]{
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "create",
				Short: "Create a new cluster",
				Args:  cobra.NoArgs,
			}
			cmd.Flags().String("name", "", "Cluster name (auto-generated if not provided)")
			cmd.Flags().String("cloud-provider", "", "Cloud provider ID (required, see 'cluster cloud-provider list)")
			cmd.Flags().String("cloud-region", "", "Cloud provider region ID (required, see 'cluster cloud-region list --cloud-provider <provider_id>)")
			cmd.Flags().String("version", "", "Qdrant version (default latest)")
			cmd.Flags().Uint32("nodes", 1, "Number of nodes (default 1)")
			cmd.Flags().String("package", "", "Booking package name or ID (see 'cluster package list')")
			cmd.Flags().String("cpu", "", "CPU to select a package (e.g. \"1\", \"0.5\", or \"1000m\")")
			cmd.Flags().String("ram", "", "RAM in GiB to select a package (e.g. \"8\", \"8G\", \"8Gi\", or \"8GiB\")")
			cmd.Flags().String("disk", "", "Total disk size (e.g. \"200GiB\"); if larger than the package's included disk, the difference is provisioned as additional storage")
			cmd.Flags().String("gpu", "", "Number of GPUs to select a package (e.g. \"1\", \"2\", or \"1000m\")")
			cmd.Flags().Bool("multi-az", false, "Require a multi-AZ package")
			cmd.Flags().StringToString("label", nil, "Label to apply to the cluster ('key=value'), can be specified multiple times")
			cmd.Flags().Bool("wait", false, "Wait for the cluster to become healthy")
			cmd.Flags().Duration("wait-timeout", 10*time.Minute, "Maximum time to wait for cluster health")
			cmd.Flags().Duration("wait-poll-interval", 5*time.Second, "How often to poll for cluster health")
			_ = cmd.Flags().MarkHidden("wait-poll-interval")
			_ = cmd.MarkFlagRequired("cloud-provider")
			_ = cmd.MarkFlagRequired("cloud-region")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) (*clusterv1.Cluster, error) {
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
			if name == "" {
				suggested, err := client.Cluster().SuggestClusterName(ctx, &clusterv1.SuggestClusterNameRequest{
					AccountId: accountID,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to suggest cluster name: %w", err)
				}
				name = suggested.GetName()
			}
			cloudProvider, _ := cmd.Flags().GetString("cloud-provider")
			cloudRegion, _ := cmd.Flags().GetString("cloud-region")
			version, _ := cmd.Flags().GetString("version")
			nodes, _ := cmd.Flags().GetUint32("nodes")
			packageValue, _ := cmd.Flags().GetString("package")
			cpu, _ := cmd.Flags().GetString("cpu")
			ram, _ := cmd.Flags().GetString("ram")
			disk, _ := cmd.Flags().GetString("disk")
			gpu, _ := cmd.Flags().GetString("gpu")
			multiAz, _ := cmd.Flags().GetBool("multi-az")
			labelMap, _ := cmd.Flags().GetStringToString("label")

			if cpu != "" {
				cpu, err = normalizeMillicores(cpu)
				if err != nil {
					return nil, fmt.Errorf("invalid --cpu value: %w", err)
				}
			}
			if ram != "" {
				ram, err = normalizeRAM(ram)
				if err != nil {
					return nil, fmt.Errorf("invalid --ram value: %w", err)
				}
			}
			if gpu != "" {
				gpu, err = normalizeMillicores(gpu)
				if err != nil {
					return nil, fmt.Errorf("invalid --gpu value: %w", err)
				}
			}

			if packageValue == "" && cpu == "" && ram == "" {
				return nil, fmt.Errorf("either --package or --cpu and --ram are required")
			}

			var pkg *bookingv1.Package
			var packageID string

			if packageValue != "" {
				if isUUID(packageValue) {
					packageID = packageValue
					if disk != "" {
						pkg, err = resolvePackageByID(ctx, client.Booking(), accountID, cloudProvider, cloudRegion, packageValue)
						if err != nil {
							return nil, err
						}
					}
				} else {
					pkg, err = resolvePackageByName(ctx, client.Booking(), accountID, cloudProvider, cloudRegion, packageValue)
					if err != nil {
						return nil, err
					}
					packageID = pkg.GetId()
				}
			} else {
				pkg, err = resolvePackageByResources(ctx, client.Booking(), accountID, cloudProvider, cloudRegion, cpu, ram, gpu, multiAz)
				if err != nil {
					return nil, err
				}
				packageID = pkg.GetId()
			}

			cluster := &clusterv1.Cluster{
				AccountId:             accountID,
				Name:                  name,
				CloudProviderId:       cloudProvider,
				CloudProviderRegionId: cloudRegion,
				Configuration: &clusterv1.ClusterConfiguration{
					NumberOfNodes: nodes,
				},
			}
			if version != "" {
				cluster.Configuration.Version = &version
			}
			if packageID != "" {
				cluster.Configuration.PackageId = packageID
			}
			for k, v := range labelMap {
				cluster.Labels = append(cluster.Labels, &commonv1.KeyValue{Key: k, Value: v})
			}

			if disk != "" && pkg != nil {
				requestedGiB, err := parseDiskGiB(disk)
				if err != nil {
					return nil, fmt.Errorf("invalid --disk value: %w", err)
				}
				if pkgDiskStr := pkg.GetResourceConfiguration().GetDisk(); pkgDiskStr != "" {
					if pkgGiB, err := parseDiskGiB(pkgDiskStr); err == nil && requestedGiB > pkgGiB {
						cluster.Configuration.AdditionalResources = &clusterv1.AdditionalResources{
							Disk: requestedGiB - pkgGiB,
						}
					}
				}
			}

			resp, err := client.Cluster().CreateCluster(ctx, &clusterv1.CreateClusterRequest{
				Cluster: cluster,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create cluster: %w", err)
			}

			created := resp.GetCluster()

			wait, _ := cmd.Flags().GetBool("wait")
			if !wait {
				return created, nil
			}

			waitTimeout, _ := cmd.Flags().GetDuration("wait-timeout")
			pollInterval, _ := cmd.Flags().GetDuration("wait-poll-interval")
			fmt.Fprintf(cmd.ErrOrStderr(), "Cluster %s created, waiting for it to become healthy...\n", created.GetId())
			return waitForHealthyWithInterval(ctx, client.Cluster(), cmd.ErrOrStderr(), accountID, created.GetId(), waitTimeout, pollInterval)
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, created *clusterv1.Cluster) {
			if ep := created.GetState().GetEndpoint(); ep != nil && ep.GetUrl() != "" {
				fmt.Fprintf(out, "Cluster %s (%s) is ready. Endpoint: %s\n", created.GetId(), created.GetName(), ep.GetUrl())
			} else {
				fmt.Fprintf(out, "Cluster %s (%s) created successfully.\n", created.GetId(), created.GetName())
			}
		},
	}.CobraCommand(s)
	_ = cmd.RegisterFlagCompletionFunc("cloud-provider", cloudProviderCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("cloud-region", cloudRegionCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("package", packageCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("version", versionCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("cpu", cpuCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("ram", ramCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("disk", diskCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("gpu", gpuCompletion(s))
	return cmd
}
