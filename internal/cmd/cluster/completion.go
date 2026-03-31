package cluster

import (
	"github.com/spf13/cobra"
)

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
