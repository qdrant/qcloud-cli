package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	defaultAPIEndpoint = "api.cloud.qdrant.io:443"
	envPrefix          = "QDRANT_CLOUD"

	// KeyManagementKey is the config key for the API management key.
	KeyManagementKey = "management_key"
	// KeyAccountID is the config key for the account ID.
	KeyAccountID = "account_id"
	// KeyEndpoint is the config key for the API endpoint.
	KeyEndpoint = "endpoint"
)

// Config wraps a viper instance to provide typed access to configuration values.
type Config struct {
	v *viper.Viper
}

// DefaultConfigPath returns the default config file path (~/.config/qcloud/config.json).
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "qcloud", "config.json")
}

// New creates a new Config backed by a fresh viper instance.
func New(configPath string) (*Config, error) {
	v := viper.New()

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigFile(DefaultConfigPath())
	}
	v.SetConfigType("json")
	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()

	v.SetDefault(KeyEndpoint, defaultAPIEndpoint)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	return &Config{v: v}, nil
}

// BindPFlag binds a viper key to a pflag.Flag.
func (c *Config) BindPFlag(key string, flag *pflag.Flag) {
	_ = c.v.BindPFlag(key, flag)
}

// APIKey returns the management key from config/env/flags.
func (c *Config) APIKey() string {
	return c.v.GetString(KeyManagementKey)
}

// AccountID returns the account ID from config/env/flags.
func (c *Config) AccountID() string {
	return c.v.GetString(KeyAccountID)
}

// Endpoint returns the API endpoint from config/env/flags.
func (c *Config) Endpoint() string {
	return c.v.GetString(KeyEndpoint)
}

// JSONOutput returns whether JSON output is enabled.
func (c *Config) JSONOutput() bool {
	return c.v.GetBool("json")
}
