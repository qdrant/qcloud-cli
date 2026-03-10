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
		Long:  "qcloud is the command-line interface for Qdrant Cloud.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.New(configPath)
			if err != nil {
				return err
			}
			cfg.BindPFlag(config.KeyManagementKey, cmd.Root().PersistentFlags().Lookup("api-key"))
			cfg.BindPFlag(config.KeyAccountID, cmd.Root().PersistentFlags().Lookup("account-id"))
			cfg.BindPFlag("json", cmd.Root().PersistentFlags().Lookup("json"))
			cfg.BindPFlag(config.KeyEndpoint, cmd.Root().PersistentFlags().Lookup("endpoint"))
			s.Config = cfg
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Global persistent flags.
	cmd.PersistentFlags().StringVar(&configPath, "config", "", "Config file path (default ~/.config/qcloud/config.json)")
	cmd.PersistentFlags().String("api-key", "", "API management key")
	cmd.PersistentFlags().String("account-id", "", "Qdrant Cloud account ID")
	cmd.PersistentFlags().String("endpoint", "", "API endpoint (default api.cloud.qdrant.io:443)")
	cmd.PersistentFlags().Bool("json", false, "Output as JSON")

	// Register subcommands.
	cmd.AddCommand(version.NewCommand(s))
	cmd.AddCommand(cluster.NewCommand(s))

	return cmd
}
