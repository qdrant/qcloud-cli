package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	defaultAPIEndpoint = "grpc.cloud.qdrant.io:443"
	envPrefix          = "QDRANT_CLOUD"

	// KeyManagementKey is the config key for the Management API key.
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

// New creates a new Config backed by a fresh viper instance. No I/O is performed.
func New() *Config {
	v := viper.New()
	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	v.AutomaticEnv()
	v.SetDefault(KeyEndpoint, defaultAPIEndpoint)
	return &Config{v: v}
}

// Load reads the config file. If configPath is non-empty it is used directly;
// otherwise QDRANT_CLOUD_CONFIG env var is checked (via viper), then the
// default ~/.config/qcloud/config.json location is used.
// A missing config file is not an error.
func (c *Config) Load(configPath string) error {
	if configPath == "" {
		configPath = c.v.GetString("config")
	}
	if configPath != "" {
		c.v.SetConfigFile(configPath)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("resolving home directory: %w", err)
		}
		c.v.SetConfigName("config")
		c.v.SetConfigType("json")
		c.v.AddConfigPath(filepath.Join(home, ".config", "qcloud"))
	}

	if err := c.v.ReadInConfig(); err != nil {
		if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); !ok {
			return fmt.Errorf("reading config: %w", err)
		}
	}
	return nil
}

// BindPFlag binds a viper key to a pflag.Flag.
func (c *Config) BindPFlag(key string, flag *pflag.Flag) {
	_ = c.v.BindPFlag(key, flag)
}

// SetDefault sets a default value for the given key (lowest viper priority).
func (c *Config) SetDefault(key, value string) {
	c.v.SetDefault(key, value)
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
