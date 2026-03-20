package cluster

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"
	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/resource"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newScaleCommand(s *state.State) *cobra.Command {
	cmd := base.UpdateCmd[*clusterv1.Cluster]{
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "scale <cluster-id>",
				Short: "Scales the resources of a cluster",
				Long:  "Scales the resources of a cluster",
				Args:  util.ExactArgs(1, "a cluster ID"),
			}
			cmd.Flags().Uint32("nodes", 0, "Number of nodes")
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompts")
			cmd.Flags().Var(new(resource.Millicores), "cpu", "CPU to select a package (e.g. \"1\", \"0.5\", or \"1000m\")")
			cmd.Flags().Var(new(resource.ByteQuantity), "ram", "RAM to select a package (e.g. \"8\", \"8G\", \"8Gi\", or \"8GiB\")")
			cmd.Flags().Var(new(resource.ByteQuantity), "disk", "Total disk size (e.g. \"200GiB\"); if larger than the package's included disk, the difference is provisioned as additional storage")
			cmd.Flags().Var(new(resource.Millicores), "gpu", "Number of GPUs to select a package (e.g. \"1\", \"2\", or \"1000m\")")
			cmd.Flags().Bool("wait", false, "Wait for the cluster to become healthy")
			cmd.Flags().Duration("wait-timeout", 10*time.Minute, "Maximum time to wait for cluster health")
			cmd.Flags().Duration("wait-poll-interval", 5*time.Second, "How often to poll for cluster health")
			_ = cmd.Flags().MarkHidden("wait-poll-interval")
			cmd.Flags().String("disk-performance", "", `Disk performance tier ("balanced", "cost-optimised", "performance")`)
			return cmd
		},
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
				ClusterId: args[0],
				AccountId: accountID,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get cluster: %w", err)
			}

			return resp.GetCluster(), nil
		},
		Update: func(s *state.State, cmd *cobra.Command, cluster *clusterv1.Cluster) (*clusterv1.Cluster, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			currentPkg, err := client.Booking().GetPackage(ctx, &bookingv1.GetPackageRequest{
				AccountId:             accountID,
				Id:                    cluster.Configuration.PackageId,
				CloudProviderId:       cluster.CloudProviderId,
				CloudProviderRegionId: &cluster.CloudProviderRegionId,
			})
			if err != nil {
				return nil, err
			}

			// the new package has to match current values when they are not changed by the user
			// so if the flags are present set them, else grab the values from the current package
			var cpu resource.Millicores
			var ram resource.ByteQuantity
			var gpu resource.Millicores
			if cmd.Flags().Changed("cpu") {
				cpu = *cmd.Flags().Lookup("cpu").Value.(*resource.Millicores)
			} else {
				cpu, err = resource.ParseMillicores(currentPkg.GetPackage().GetResourceConfiguration().GetCpu())
				if err != nil {
					return nil, err
				}
			}

			if cmd.Flags().Changed("ram") {
				ram = *cmd.Flags().Lookup("ram").Value.(*resource.ByteQuantity)
			} else {
				ram, err = resource.ParseByteQuantity(currentPkg.GetPackage().GetResourceConfiguration().GetRam())
				if err != nil {
					return nil, err
				}
			}

			if cmd.Flags().Changed("gpu") {
				gpu = *cmd.Flags().Lookup("gpu").Value.(*resource.Millicores)
			} else {
				if currentPkg.GetPackage().GetResourceConfiguration().GetGpu() != "" {
					gpu, err = resource.ParseMillicores(currentPkg.GetPackage().GetResourceConfiguration().GetGpu())
					if err != nil {
						return nil, err
					}
				}
			}

			currentPkgDisk, err := resource.ParseByteQuantity(currentPkg.GetPackage().GetResourceConfiguration().GetDisk())
			if err != nil {
				return nil, err
			}

			// If no resource flags changed, keep the current package — avoids a
			// ListPackages round-trip and prevents spurious failures when the current
			// package is deprecated or shares specs with another active package.
			var newPkg *bookingv1.Package
			if cmd.Flags().Changed("cpu") || cmd.Flags().Changed("ram") || cmd.Flags().Changed("gpu") {
				// scale doesn't allow changing multi-az so any new package selected needs to
				// use the same multi-az value
				newPkg, err = resolvePackageByResources(
					ctx,
					client.Booking(),
					accountID,
					cluster.CloudProviderId,
					cluster.CloudProviderRegionId,
					cpu,
					gpu,
					ram,
					currentPkg.GetPackage().GetMultiAz(),
				)
				if err != nil {
					return nil, err
				}
				cluster.Configuration.PackageId = newPkg.Id
			} else {
				newPkg = currentPkg.GetPackage()
			}

			// disk handling
			newPkgDisk, err := resource.ParseByteQuantity(newPkg.GetResourceConfiguration().GetDisk())
			if err != nil {
				return nil, err
			}

			currentAdditionalDisk := resource.ByteQuantity(int64(cluster.Configuration.AdditionalResources.GetDisk()) * int64(resource.GiB))
			currentTotalDisk := currentPkgDisk + currentAdditionalDisk

			// If a new package is selected from user changes and the user changes the disk too
			// but it's smaller than the new packages' minimum disk value, it will be overriden.
			// This is used to notify the user about it.
			diskWillBeOverridden := false
			var newEffectiveDisk resource.ByteQuantity
			var requestedDisk resource.ByteQuantity
			if cmd.Flags().Changed("disk") {
				requestedDisk = *cmd.Flags().Lookup("disk").Value.(*resource.ByteQuantity)
				newEffectiveDisk = max(requestedDisk, newPkgDisk)
				if newEffectiveDisk < currentTotalDisk {
					return nil, fmt.Errorf("disk cannot be downscaled from %s to %s", currentTotalDisk, requestedDisk)
				}

				// only apply additional disk calculation if requested disk is bigger than the disk package
				// if disk is less than the package disk, let the api fail
				cluster.Configuration.AdditionalResources = &clusterv1.AdditionalResources{
					Disk: uint32(newEffectiveDisk.GiB() - newPkgDisk.GiB()),
				}

				if requestedDisk < newEffectiveDisk {
					diskWillBeOverridden = true
				}
			} else {
				// at this point the user didn't request a disk value but a package change could
				// make the disk change.
				newEffectiveDisk = max(currentTotalDisk, newPkgDisk)

				cluster.Configuration.AdditionalResources = &clusterv1.AdditionalResources{
					Disk: uint32(newEffectiveDisk.GiB() - newPkgDisk.GiB()),
				}
			}

			oldNodes := cluster.Configuration.NumberOfNodes
			if cmd.Flags().Changed("nodes") {
				nodes, err := cmd.Flags().GetUint32("nodes")
				if err != nil {
					return nil, err
				}

				if nodes == 0 {
					return nil, errors.New("nodes can't be downscaled to 0")
				}

				cluster.Configuration.NumberOfNodes = nodes
			}

			oldStorageTier := storageTierString(cluster.Configuration.GetClusterStorageConfiguration().GetStorageTierType())
			if cmd.Flags().Changed("disk-performance") {
				perfStr, _ := cmd.Flags().GetString("disk-performance")
				tierType, err := parseDiskPerformance(perfStr)
				if err != nil {
					return nil, err
				}

				clusterStorageConfig := cluster.GetConfiguration().GetClusterStorageConfiguration()
				if clusterStorageConfig != nil {
					clusterStorageConfig.StorageTierType = tierType
				} else {
					cluster.GetConfiguration().ClusterStorageConfiguration = &clusterv1.ClusterStorageConfiguration{
						StorageTierType: tierType,
					}
				}
			}
			newStorageTier := storageTierString(cluster.Configuration.GetClusterStorageConfiguration().GetStorageTierType())

			force, _ := cmd.Flags().GetBool("force")
			prompt := scaleConfirmPrompt(
				cluster,
				currentPkg.GetPackage(),
				newPkg,
				oldNodes,
				currentTotalDisk,
				newEffectiveDisk,
				requestedDisk,
				diskWillBeOverridden,
				oldStorageTier,
				newStorageTier,
			)
			if !util.ConfirmAction(force, prompt) {
				fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
				return nil, nil
			}

			resp, err := client.Cluster().UpdateCluster(ctx, &clusterv1.UpdateClusterRequest{
				Cluster: cluster,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to update cluster: %w", err)
			}

			wait, _ := cmd.Flags().GetBool("wait")
			if !wait {
				return resp.GetCluster(), nil
			}

			waitTimeout, _ := cmd.Flags().GetDuration("wait-timeout")
			pollInterval, _ := cmd.Flags().GetDuration("wait-poll-interval")
			fmt.Fprintf(cmd.ErrOrStderr(), "Scaling Cluster %s (%s)...\n", resp.GetCluster().GetId(), resp.GetCluster().GetName())
			return waitForHealthyWithInterval(
				ctx,
				client.Cluster(),
				cmd.ErrOrStderr(),
				accountID,
				resp.GetCluster().GetId(),
				waitTimeout,
				pollInterval,
			)
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, updated *clusterv1.Cluster) {
			if updated == nil {
				return
			}
			if updated.State.Phase != clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY {
				fmt.Fprintf(out, "Cluster %s (%s) is scaling, it will take some time to take effect. Use 'cluster wait %s' to wait for it to become healthy\n", updated.GetId(), updated.GetName(), updated.GetId())
				return
			}
			fmt.Fprintf(out, "Cluster %s (%s) scaled successfully.\n", updated.GetId(), updated.GetName())
		},
		ValidArgsFunction: completion.ClusterIDCompletion(s),
	}.CobraCommand(s)

	_ = cmd.RegisterFlagCompletionFunc("cpu", cpuCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("ram", ramCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("disk", diskCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("gpu", gpuCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("disk-performance", diskPerformanceCompletion())
	return cmd
}

// scaleDiff formats a field value as "old => new" when the value changes, or just "val" when unchanged.
func scaleDiff(oldVal, newVal string) string {
	if oldVal == newVal {
		return newVal
	}
	return oldVal + " => " + newVal
}

// scaleConfirmPrompt builds the confirmation message shown before a scale operation,
// displaying old => new for fields that are changing.
func scaleConfirmPrompt(
	cluster *clusterv1.Cluster,
	oldPkg, newPkg *bookingv1.Package,
	oldNodes uint32,
	currentTotalDisk, newEffectiveDisk, requestedDisk resource.ByteQuantity,
	diskWillBeOverridden bool,
	oldStorageTier, newStorageTier string,
) string {
	oldRC := oldPkg.GetResourceConfiguration()
	newRC := newPkg.GetResourceConfiguration()

	diskLine := scaleDiff(currentTotalDisk.String(), newEffectiveDisk.String())
	if diskWillBeOverridden {
		diskLine = fmt.Sprintf("%s (requested: %s - package minimum disk is being applied)", scaleDiff(currentTotalDisk.String(), newEffectiveDisk.String()), requestedDisk)
	}

	prompt := fmt.Sprintf(
		"Cluster %s (%s) will be scaled to:\n  Nodes:   %s\n  CPU:     %s\n  RAM:     %s\n  Disk:    %s",
		cluster.GetId(), cluster.GetName(),
		scaleDiff(fmt.Sprintf("%d", oldNodes), fmt.Sprintf("%d", cluster.Configuration.NumberOfNodes)),
		scaleDiff(oldRC.GetCpu(), newRC.GetCpu()),
		scaleDiff(oldRC.GetRam(), newRC.GetRam()),
		diskLine,
	)
	if oldRC.GetGpu() != "" || newRC.GetGpu() != "" {
		prompt += fmt.Sprintf("\n  GPU:     %s", scaleDiff(oldRC.GetGpu(), newRC.GetGpu()))
	}
	if oldStorageTier != "" || newStorageTier != "" {
		prompt += fmt.Sprintf("\n  Storage tier: %s", scaleDiff(oldStorageTier, newStorageTier))
	}
	prompt += "\nProceed?"
	return prompt
}
