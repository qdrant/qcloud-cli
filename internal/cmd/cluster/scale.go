package cluster

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/spf13/cobra"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"
	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/clusterutil"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/resource"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newScaleCommand(s *state.State) *cobra.Command {
	cmd := base.UpdateCmd[*clusterv1.Cluster]{
		Example: `# Scale up CPU and RAM
qcloud cluster scale 7b2ea926-724b-4de2-b73a-8675c42a6ebe --cpu 4 --ram 16Gi

# Add more nodes
qcloud cluster scale 7b2ea926-724b-4de2-b73a-8675c42a6ebe --nodes 3

# Increase disk and wait for completion
qcloud cluster scale 7b2ea926-724b-4de2-b73a-8675c42a6ebe --disk 500Gi --wait`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "scale <cluster-id>",
				Short: "Scales the resources of a cluster",
				Long: `Scales the resources of a cluster.

Use this command to change the resource package for the nodes of a Qdrant cluster, change
the number of nodes in the cluster, and allocate more disk space per node. You can select
one of the pre-defined resource packages to apply for all nodes in the cluster. A package
defines CPU, RAM, GPU, and minimum disk size.

The --cpu, --ram, and --gpu flags specify per-node resources and are used to match a
package. If none of these flags are provided, the current package is preserved. Available
packages can be listed with 'package list' using the --cloud-provider and
--cloud-region flags.

Each package includes a baseline disk size. Requesting more disk than the baseline with
--disk provisions the difference as additional storage. Disk cannot be downscaled. If a
new package has a larger baseline disk than the current total, the disk size increases to
match.

Reducing RAM is always allowed, but a cluster whose memory usage does not fit the smaller
package can run out of memory and become unavailable. A warning is printed before the
scale when that is the case, or when the usage could not be determined; --force skips the
confirmation but not the warning.`,
				Args: util.ExactArgs(1, "a cluster ID"),
			}
			cmd.Flags().Uint32("nodes", 0, "Number of nodes")
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompts")
			cmd.Flags().Var(new(resource.Millicores), "cpu", "CPU per node (e.g. \"1\", \"0.5\", or \"1000m\")")
			cmd.Flags().Var(new(resource.ByteQuantity), "ram", "RAM per node (e.g. \"8\", \"8G\", \"8Gi\", or \"8GiB\")")
			cmd.Flags().Var(new(resource.ByteQuantity), "disk", "Total disk size per node (e.g. \"200GiB\"); if larger than the node's included disk, the difference is provisioned as additional storage")
			cmd.Flags().Var(new(resource.Millicores), "gpu", "Number of GPUs per node (e.g. \"1\", \"2\", or \"1000m\")")
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

			multiAz := currentPkg.GetPackage().GetMultiAz()

			// If no resource flags changed, keep the current package — avoids a
			// ListPackages round-trip and prevents spurious failures when the current
			// package is deprecated or shares specs with another active package.
			var newPkg *bookingv1.Package
			if util.AnyFlagChanged(cmd, []string{"cpu", "ram", "gpu"}) {
				newPkg, err = clusterutil.ResolvePackageByResources(ctx, client.Booking(), clusterutil.PackageResourceQuery{
					AccountID:     accountID,
					CloudProvider: cluster.CloudProviderId,
					CloudRegion:   &cluster.CloudProviderRegionId,
					CPU:           cpu,
					GPU:           gpu,
					RAM:           ram,
					MultiAz:       multiAz,
				})
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

			// Printed before the prompt and independently of it, so that --force keeps the
			// warning in the log of whatever ran the command.
			if warning := memoryDownscaleWarning(ctx, s.Logger, client.Cluster(), accountID, cluster.GetId(), currentPkg.GetPackage(), newPkg); warning != "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %s\n", warning)
			}

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
			if !util.ConfirmAction(force, cmd.ErrOrStderr(), prompt) {
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
			return clusterutil.WaitForClusterHealthy(
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

			if updated.GetState().GetPhase() != clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY {
				fmt.Fprintf(out, "Cluster %s (%s) is scaling, it will take some time to take effect. Use 'cluster wait %s' to wait for it to become healthy\n", updated.GetId(), updated.GetName(), updated.GetId())
				return
			}

			fmt.Fprintf(out, "Cluster %s (%s) scaled successfully.\n", updated.GetId(), updated.GetName())
		},
		ValidArgsFunction: completion.ClusterIDCompletion(s),
	}.CobraCommand(s)

	_ = cmd.RegisterFlagCompletionFunc("cpu", completion.CPUCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("ram", completion.RAMCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("disk", completion.DiskCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("gpu", completion.GPUCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("disk-performance", diskPerformanceCompletion())
	return cmd
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

	diskLine := output.DiffValue(currentTotalDisk.String(), newEffectiveDisk.String())
	if diskWillBeOverridden {
		diskLine = fmt.Sprintf("%s (requested: %s - package minimum disk is being applied)", output.DiffValue(currentTotalDisk.String(), newEffectiveDisk.String()), requestedDisk)
	}

	prompt := fmt.Sprintf(
		"Cluster %s (%s) will be scaled to:\n  Nodes:   %s\n  CPU:     %s\n  RAM:     %s\n  Disk:    %s",
		cluster.GetId(), cluster.GetName(),
		output.DiffValue(fmt.Sprintf("%d", oldNodes), fmt.Sprintf("%d", cluster.Configuration.NumberOfNodes)),
		output.DiffValue(oldRC.GetCpu(), newRC.GetCpu()),
		output.DiffValue(oldRC.GetRam(), newRC.GetRam()),
		diskLine,
	)
	if oldRC.GetGpu() != "" || newRC.GetGpu() != "" {
		prompt += fmt.Sprintf("\n  GPU:     %s", output.DiffValue(oldRC.GetGpu(), newRC.GetGpu()))
	}

	if oldStorageTier != "" || newStorageTier != "" {
		prompt += fmt.Sprintf("\n  Storage tier: %s", output.DiffValue(oldStorageTier, newStorageTier))
	}

	prompt += "\nProceed?"
	return prompt
}

// downscaleRiskTimeout bounds the risk read. The verdict costs the platform a metrics
// query, and a slow metrics source must not hold up a scale the platform accepts either
// way.
const downscaleRiskTimeout = 5 * time.Second

// The copy to fall back on when the platform reports a status without a reason of its
// own. Its own `reason` is preferred, so that it can be reworded without a CLI release.
const (
	usageExceedsTargetWarning = "Your cluster is currently using more RAM than the selected configuration provides. " +
		"Scaling down may cause your cluster to run out of memory and become unavailable."
	// Also stands in for a verdict that could not be read at all: either way the usage is
	// unknown, so the copy states the consequence without naming a cause.
	usageUnknownWarning = "Your cluster's current RAM usage could not be determined. If you scale down, your cluster " +
		"could run out of memory and become unavailable if its usage exceeds the new configuration."
)

// memoryDownscaleWarning returns the warning to show before moving a cluster from oldPkg
// to newPkg, or "" when there is nothing to warn about.
//
// The verdict is advisory: the platform allows a downscale whatever it says, so a verdict
// that cannot be read is reported as an unknown usage rather than failing the scale.
func memoryDownscaleWarning(
	ctx context.Context,
	log *slog.Logger,
	svc clusterv1.ClusterServiceClient,
	accountID, clusterID string,
	oldPkg, newPkg *bookingv1.Package,
) string {
	// Only a reduction is worth a round trip: the platform answers anything else with
	// "no risk", and only after it has resolved the cluster and the candidate package.
	if !isRAMReduction(oldPkg, newPkg) {
		log.Debug("downscale risk not assessed, not a RAM reduction",
			"current", oldPkg.GetResourceConfiguration().GetRam(),
			"candidate", newPkg.GetResourceConfiguration().GetRam())
		return ""
	}

	ctx, cancel := context.WithTimeout(ctx, downscaleRiskTimeout)
	defer cancel()

	resp, err := svc.GetClusterDownscaleRisk(ctx, &clusterv1.GetClusterDownscaleRiskRequest{
		AccountId: accountID,
		ClusterId: clusterID,
		PackageId: newPkg.GetId(),
	})
	if err != nil {
		// The warning cannot name the cause, so --debug is the only place the difference
		// between an unreachable metrics source and a rejected call is visible.
		log.Debug("downscale risk could not be read", "error", err)
		return usageUnknownWarning
	}

	var memoryRisk *clusterv1.ClusterDownscaleRiskInfo
	for _, risk := range resp.GetDownscaleRisks() {
		if risk.GetResource() == clusterv1.ClusterDownscaleRiskResource_CLUSTER_DOWNSCALE_RISK_RESOURCE_MEMORY {
			memoryRisk = risk
			break
		}
	}

	// A verdict that came back without a memory entry leaves the usage as unknown as an
	// explicit CLUSTER_DOWNSCALE_RISK_STATUS_USAGE_UNKNOWN does.
	if memoryRisk == nil {
		log.Debug("downscale risk verdict carries no memory entry")
		return usageUnknownWarning
	}

	// Without this, a "no risk" verdict and a check that never ran look the same from
	// outside: both print nothing.
	log.Debug("downscale risk assessed", "status", memoryRisk.GetStatus().String())

	reason := memoryRisk.GetReason()
	switch memoryRisk.GetStatus() {
	case clusterv1.ClusterDownscaleRiskStatus_CLUSTER_DOWNSCALE_RISK_STATUS_NO_RISK:
		return ""
	case clusterv1.ClusterDownscaleRiskStatus_CLUSTER_DOWNSCALE_RISK_STATUS_USAGE_EXCEEDS_TARGET:
		if reason == "" {
			return usageExceedsTargetWarning
		}

	// An unspecified status, and one this client does not know, are both as unknown as
	// CLUSTER_DOWNSCALE_RISK_STATUS_USAGE_UNKNOWN itself.
	default:
		if reason == "" {
			return usageUnknownWarning
		}
	}

	return reason
}

// isRAMReduction reports whether newPkg books less RAM per node than oldPkg. RAM that
// cannot be read on either side is not reported as a reduction: the platform assesses
// against the cluster's provisioned total, which this comparison only approximates, and
// the warning is not worth a false positive.
func isRAMReduction(oldPkg, newPkg *bookingv1.Package) bool {
	oldRAM, err := resource.ParseByteQuantity(oldPkg.GetResourceConfiguration().GetRam())
	if err != nil {
		return false
	}

	newRAM, err := resource.ParseByteQuantity(newPkg.GetResourceConfiguration().GetRam())
	if err != nil {
		return false
	}

	return newRAM < oldRAM
}
