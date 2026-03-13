// Package config provides configuration management for libcode
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	// AI Provider Settings
	Providers map[string]ProviderConfig `mapstructure:"providers"`

	// Model Settings
	DefaultModel string `mapstructure:"default_model"`
	ModelVariant string `mapstructure:"model_variant"`

	// Session Settings
	SessionDir string `mapstructure:"session_dir"`
	MaxHistory int    `mapstructure:"max_history"`

	// LSP Settings
	LSP LSPConfig `mapstructure:"lsp"`

	// MCP Settings
	MCP MCPConfig `mapstructure:"mcp"`

	// Logging
	LogLevel string `mapstructure:"log_level"`
	LogFile  string `mapstructure:"log_file"`

	// Feature Flags
	Features FeatureFlags `mapstructure:"features"`
}

type ProviderConfig struct {
	APIKey      string `mapstructure:"api_key"`
	BaseURL     string `mapstructure:"base_url"`
	Enabled     bool   `mapstructure:"enabled"`
	Model       string `mapstructure:"model"`
}

type LSPConfig struct {
	Enabled       bool     `mapstructure:"enabled"`
	ServerTimeout int      `mapstructure:"server_timeout"`
	Languages     []string `mapstructure:"languages"`
}

type MCPConfig struct {
	Enabled       bool     `mapstructure:"enabled"`
	ServerTimeout int      `mapstructure:"server_timeout"`
	Servers       []string `mapstructure:"servers"`
}

type FeatureFlags struct {
	TUI            bool `mapstructure:"tui"`
	DesktopMode    bool `mapstructure:"desktop_mode"`
	StreamResponse bool `mapstructure:"stream_response"`
}

var globalConfig *Config

// Load reads configuration from multiple sources in priority order:
// 1. Environment variables (LIBCODE_*)
// 2. Config file (~/.config/libcode/config.yaml)
// 3. Flags (set via SetFlags)
// 4. Defaults
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Config file paths
	configHome, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config dir: %w", err)
	}

	configPath := filepath.Join(configHome, "libcode")
	v.AddConfigPath(configPath)
	v.AddConfigPath(".")
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Environment variables
	v.SetEnvPrefix("LIBCODE")
	v.AutomaticEnv()

	// Read config file (if exists)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		// Config file not found, use defaults
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	globalConfig = cfg
	return cfg, nil
}

// setDefaults establishes the default configuration values
func setDefaults(v *viper.Viper) {
	// Model settings
	v.SetDefault("default_model", "openai/gpt-4")
	v.SetDefault("model_variant", "")

	// Session settings
	v.SetDefault("session_dir", "~/.local/share/libcode/sessions")
	v.SetDefault("max_history", 100)

	// LSP settings
	v.SetDefault("lsp.enabled", true)
	v.SetDefault("lsp.server_timeout", 30)
	v.SetDefault("lsp.languages", []string{"typescript", "go", "python", "rust"})

	// MCP settings
	v.SetDefault("mcp.enabled", true)
	v.SetDefault("mcp.server_timeout", 30)

	// Logging
	v.SetDefault("log_level", "INFO")
	v.SetDefault("log_file", "")

	// Features
	v.SetDefault("features.tui", true)
	v.SetDefault("features.desktop_mode", false)
	v.SetDefault("features.stream_response", true)
}

// Get returns the global configuration
func Get() *Config {
	if globalConfig == nil {
		cfg, err := Load()
		if err != nil {
			// Return defaults if load fails
			v := viper.New()
			setDefaults(v)
			cfg = &Config{}
			v.Unmarshal(cfg)
			globalConfig = cfg
		}
	}
	return globalConfig
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	if c.DefaultModel == "" {
		return fmt.Errorf("default_model is required")
	}

	// Validate provider configs
	for name, provider := range c.Providers {
		if provider.Enabled && provider.APIKey == "" {
			// Check environment variable
			envVar := fmt.Sprintf("LIBCODE_PROVIDERS_%s_APIKEY", name)
			if os.Getenv(envVar) == "" {
				return fmt.Errorf("provider %s is enabled but has no API key", name)
			}
		}
	}

	return nil
}
