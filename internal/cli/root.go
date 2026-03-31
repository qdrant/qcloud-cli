package cli

import (
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/backup"
	"github.com/qdrant/qcloud-cli/internal/cmd/cloudprovider"
	"github.com/qdrant/qcloud-cli/internal/cmd/cloudregion"
	"github.com/qdrant/qcloud-cli/internal/cmd/cluster"
	contextcmd "github.com/qdrant/qcloud-cli/internal/cmd/context"
	"github.com/qdrant/qcloud-cli/internal/cmd/hybrid"
	packagecmd "github.com/qdrant/qcloud-cli/internal/cmd/package"
	"github.com/qdrant/qcloud-cli/internal/cmd/selfupgrade"
	"github.com/qdrant/qcloud-cli/internal/cmd/version"
	"github.com/qdrant/qcloud-cli/internal/state"
	"github.com/qdrant/qcloud-cli/internal/state/config"
)

// NewRootCommand creates the root cobra command with all subcommands registered.
func NewRootCommand(s *state.State) *cobra.Command {
	var configPath string
	var debug bool

	cmd := &cobra.Command{
		Use:   "qcloud",
		Short: "Qdrant Cloud CLI",
		Long: `The command-line interface for Qdrant Cloud.

Get started:
  qcloud context set default --api-key <KEY> --account-id <ID>
  qcloud cluster list

Documentation: https://github.com/qdrant/qcloud-cli`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := s.Config.Load(configPath); err != nil {
				return err
			}
			if debug {
				s.Logger = slog.New(slog.NewTextHandler(cmd.ErrOrStderr(), &slog.HandlerOptions{
					Level: slog.LevelDebug,
				}))
			}
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging to stderr")
	cmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Config file path (env: QDRANT_CLOUD_CONFIG, default ~/.config/qcloud/config.yaml)")
	cmd.PersistentFlags().String("api-key", "", "Management API Key (env: QDRANT_CLOUD_API_KEY)")
	cmd.PersistentFlags().String("account-id", "", "Qdrant Cloud Account ID (env: QDRANT_CLOUD_ACCOUNT_ID)")
	cmd.PersistentFlags().String("endpoint", "", "gRPC API endpoint (env: QDRANT_CLOUD_ENDPOINT, default grpc.cloud.qdrant.io:443)")
	cmd.PersistentFlags().Bool("json", false, "Output as JSON")
	cmd.PersistentFlags().String("context", "", "Override the active context (env: QDRANT_CLOUD_CONTEXT)")
	_ = cmd.MarkFlagFilename("config")

	s.Config.BindPFlag("config", cmd.PersistentFlags().Lookup("config"))
	s.Config.BindPFlag("json", cmd.PersistentFlags().Lookup("json"))
	s.Config.BindPFlag("context", cmd.PersistentFlags().Lookup("context"))
	s.Config.BindPFlag(config.KeyAPIKey, cmd.PersistentFlags().Lookup("api-key"))
	s.Config.BindPFlag(config.KeyAccountID, cmd.PersistentFlags().Lookup("account-id"))
	s.Config.BindPFlag(config.KeyEndpoint, cmd.PersistentFlags().Lookup("endpoint"))

	cmd.AddCommand(version.NewCommand(s))
	cmd.AddCommand(cluster.NewCommand(s))
	cmd.AddCommand(cloudprovider.NewCommand(s))
	cmd.AddCommand(cloudregion.NewCommand(s))
	cmd.AddCommand(contextcmd.NewCommand(s))
	cmd.AddCommand(backup.NewCommand(s))
	cmd.AddCommand(hybrid.NewCommand(s))
	cmd.AddCommand(packagecmd.NewCommand(s))
	cmd.AddCommand(selfupgrade.NewCommand(s))

	return cmd
}
