package cluster

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/state"
)

// cloudProviderFn reads the cloud provider and region from the command's flags.
func cloudProviderFn(cmd *cobra.Command) (string, *string) {
	provider, _ := cmd.Flags().GetString("cloud-provider")
	region, _ := cmd.Flags().GetString("cloud-region")
	if region != "" {
		return provider, &region
	}
	return provider, nil
}

func cpuCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return completion.CPUCompletion(s, cloudProviderFn)
}

func ramCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return completion.RAMCompletion(s, cloudProviderFn)
}

func diskCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return completion.DiskCompletion(s, cloudProviderFn)
}

func gpuCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return completion.GPUCompletion(s, cloudProviderFn)
}

func packageCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return completion.PackageNameCompletion(s, cloudProviderFn)
}

// diskPerformanceCompletion returns a static completion function for the --disk-performance flag.
func diskPerformanceCompletion() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{diskPerfBalanced, diskPerfCostOptimised, diskPerfPerformance}, cobra.ShellCompDirectiveNoFileComp
	}
}

// restartModeCompletion returns a static completion function for the --restart-mode flag.
func restartModeCompletion() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{restartModeRolling, restartModeParallel, restartModeAutomatic}, cobra.ShellCompDirectiveNoFileComp
	}
}

// rebalanceStrategyCompletion returns a static completion function for the --rebalance-strategy flag.
func rebalanceStrategyCompletion() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{rebalanceStrategyByCount, rebalanceStrategyBySize, rebalanceStrategyByCountAndSize}, cobra.ShellCompDirectiveNoFileComp
	}
}

// dbLogLevelCompletion returns a static completion function for the --db-log-level flag.
func dbLogLevelCompletion() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{dbLogLevelTrace, dbLogLevelDebug, dbLogLevelInfo, dbLogLevelWarn, dbLogLevelError, dbLogLevelOff}, cobra.ShellCompDirectiveNoFileComp
	}
}

// auditLogRotationCompletion returns a static completion function for the --audit-log-rotation flag.
func auditLogRotationCompletion() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{auditLogRotationDaily, auditLogRotationHourly}, cobra.ShellCompDirectiveNoFileComp
	}
}

func serviceTypeCompletion() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{serviceTypeClusterIP, serviceTypeNodePort, serviceTypeLoadBalancer}, cobra.ShellCompDirectiveNoFileComp
	}
}
