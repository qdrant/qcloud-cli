package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const (
	defaultAPIEndpoint = "grpc.cloud.qdrant.io:443"
	envPrefix          = "QDRANT_CLOUD"

	// KeyAPIKey is the config key for the Management API key.
	KeyAPIKey = "api_key"
	// KeyAccountID is the config key for the account ID.
	KeyAccountID = "account_id"
	// KeyEndpoint is the config key for the API endpoint.
	KeyEndpoint = "endpoint"
)

// ContextEntry holds the configuration for a single named context.
type ContextEntry struct {
	Name      string `mapstructure:"name"        yaml:"name"                  json:"name"`
	Endpoint  string `mapstructure:"endpoint"    yaml:"endpoint,omitempty"    json:"endpoint,omitempty"`
	APIKey    string `mapstructure:"api_key"     yaml:"api_key,omitempty"     json:"api_key,omitempty"`
	AccountID string `mapstructure:"account_id"  yaml:"account_id,omitempty"  json:"account_id,omitempty"`
}

// File holds the top-level structure of the config file.
type File struct {
	CurrentContext string         `mapstructure:"current_context" yaml:"current_context,omitempty" json:"current_context,omitempty"`
	Contexts       []ContextEntry `mapstructure:"contexts"        yaml:"contexts,omitempty"        json:"contexts,omitempty"`
}

// Config wraps a viper instance to provide typed access to configuration values.
type Config struct {
	v        *viper.Viper
	filePath string // from v.ConfigFileUsed() after Load
	file     File   // parsed file contents, for write-back
}

// DefaultConfigPath returns the default config file path (~/.config/qcloud/config.yaml).
// The config file can also be written as config.yaml or config.yml.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "qcloud", "config.yaml")
}

// New creates a new Config backed by a fresh viper instance with defaults applied. No I/O is performed.
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
// default ~/.config/qcloud/config.yaml location is used.
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
		c.v.AddConfigPath(filepath.Join(home, ".config", "qcloud"))
	}

	if err := c.v.ReadInConfig(); err != nil {
		if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); !ok {
			return fmt.Errorf("reading config: %w", err)
		}
	}

	c.filePath = c.v.ConfigFileUsed()
	c.file = File{}

	if err := c.v.Unmarshal(&c.file); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	activeContext := c.v.GetString("context")
	if activeContext == "" {
		activeContext = c.file.CurrentContext
	}

	// Inject active context values at config-file priority so env vars and flags still win.
	for _, ctx := range c.file.Contexts {
		if ctx.Name == activeContext {
			flat := map[string]any{}
			if ctx.Endpoint != "" {
				flat[KeyEndpoint] = ctx.Endpoint
			}
			if ctx.APIKey != "" {
				flat[KeyAPIKey] = ctx.APIKey
			}
			if ctx.AccountID != "" {
				flat[KeyAccountID] = ctx.AccountID
			}
			_ = c.v.MergeConfigMap(flat)
			break
		}
	}

	return nil
}

// BindPFlag binds a viper key to a pflag.Flag.
func (c *Config) BindPFlag(key string, flag *pflag.Flag) {
	_ = c.v.BindPFlag(key, flag)
}

// APIKey returns the management api key from config/env/flags.
func (c *Config) APIKey() string {
	return c.v.GetString(KeyAPIKey)
}

// AccountID returns the account ID from config/env/flags.
func (c *Config) AccountID() string {
	return c.v.GetString(KeyAccountID)
}

// SetAPIKey overrides the api key with the highest viper priority.
// FOR TESTING ONLY: this bypasses the normal config precedence (file → env var → flag)
// and should not be used in production code paths.
func (c *Config) SetAPIKey(k string) {
	c.v.Set(KeyAPIKey, k)
}

// SetAccountID overrides the account ID with the highest viper priority.
// FOR TESTING ONLY: this bypasses the normal config precedence (file → env var → flag)
// and should not be used in production code paths.
func (c *Config) SetAccountID(id string) {
	c.v.Set(KeyAccountID, id)
}

// Endpoint returns the API endpoint from config/env/flags.
func (c *Config) Endpoint() string {
	return c.v.GetString(KeyEndpoint)
}

// JSONOutput returns whether JSON output is enabled.
func (c *Config) JSONOutput() bool {
	return c.v.GetBool("json")
}

// CurrentContext returns the current_context value from the config file.
func (c *Config) CurrentContext() string {
	return c.file.CurrentContext
}

// ActiveContext returns the active context name: --context flag if set, else current_context.
func (c *Config) ActiveContext() string {
	if ctx := c.v.GetString("context"); ctx != "" {
		return ctx
	}
	return c.file.CurrentContext
}

// ContextNames returns a sorted list of all context names.
func (c *Config) ContextNames() []string {
	names := make([]string, 0, len(c.file.Contexts))
	for _, ctx := range c.file.Contexts {
		names = append(names, ctx.Name)
	}
	sort.Strings(names)
	return names
}

// GetContext returns the ContextEntry for the named context, or false if not found.
func (c *Config) GetContext(name string) (ContextEntry, bool) {
	for _, ctx := range c.file.Contexts {
		if ctx.Name == name {
			return ctx, true
		}
	}
	return ContextEntry{}, false
}

// ConfigFilePath returns the path to the loaded config file.
func (c *Config) ConfigFilePath() string {
	return c.filePath
}

// SetCurrentContext sets the current_context in the file data.
func (c *Config) SetCurrentContext(name string) {
	c.file.CurrentContext = name
}

// UpsertContext creates or updates a named context in the file data.
func (c *Config) UpsertContext(ctx ContextEntry) {
	for i, existing := range c.file.Contexts {
		if existing.Name == ctx.Name {
			c.file.Contexts[i] = ctx
			return
		}
	}
	c.file.Contexts = append(c.file.Contexts, ctx)
}

// DeleteContext removes a named context from the file data.
// If the deleted context is the current_context, it is also cleared.
func (c *Config) DeleteContext(name string) {
	filtered := c.file.Contexts[:0]
	for _, ctx := range c.file.Contexts {
		if ctx.Name != name {
			filtered = append(filtered, ctx)
		}
	}
	c.file.Contexts = filtered
	if c.file.CurrentContext == name {
		c.file.CurrentContext = ""
	}
}

// WriteToFile writes file data to the config file.
// If no file was loaded, it writes to DefaultConfigPath() as YAML.
func (c *Config) WriteToFile() error {
	path := c.filePath
	if path == "" {
		path = DefaultConfigPath()
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return fmt.Errorf("creating config dir: %w", err)
		}
	}

	var data []byte
	var err error
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".json" {
		data, err = json.MarshalIndent(c.file, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling config: %w", err)
		}
		data = append(data, '\n')
	} else {
		data, err = yaml.Marshal(c.file)
		if err != nil {
			return fmt.Errorf("marshaling config: %w", err)
		}
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	if c.filePath == "" {
		c.filePath = path
	}

	return nil
}
