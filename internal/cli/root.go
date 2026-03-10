package cli

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/cluster"
	"github.com/qdrant/qcloud-cli/internal/cmd/version"
	"github.com/qdrant/qcloud-cli/internal/state"
	"github.com/qdrant/qcloud-cli/internal/state/config"
)

// NewRootCommand creates the root cobra command with all subcommands registered.
func NewRootCommand(s *state.State) *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "qcloud",
		Short: "Qdrant Cloud CLI",
		Long:  "The command-line interface for Qdrant Cloud",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return s.Config.Load(configPath)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Config file path (default ~/.config/qcloud/config.json)")
	cmd.PersistentFlags().String("api-key", "", "Management API Key")
	cmd.PersistentFlags().String("account-id", "", "Qdrant Cloud Account ID")
	cmd.PersistentFlags().String("endpoint", "", "API endpoint (default api.cloud.qdrant.io:443)")
	cmd.PersistentFlags().Bool("json", false, "Output as JSON")

	s.Config.BindPFlag("config", cmd.PersistentFlags().Lookup("config"))
	s.Config.BindPFlag(config.KeyManagementKey, cmd.PersistentFlags().Lookup("api-key"))
	s.Config.BindPFlag(config.KeyAccountID, cmd.PersistentFlags().Lookup("account-id"))
	s.Config.BindPFlag("json", cmd.PersistentFlags().Lookup("json"))
	s.Config.BindPFlag(config.KeyEndpoint, cmd.PersistentFlags().Lookup("endpoint"))

	cmd.AddCommand(version.NewCommand(s))
	cmd.AddCommand(cluster.NewCommand(s))

	return cmd
}
