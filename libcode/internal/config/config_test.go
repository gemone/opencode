package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	// Clear any existing config
	globalConfig = nil

	cfg, err := Load()
	require.NoError(t, err)

	// Check defaults
	assert.Equal(t, "openai/gpt-4", cfg.DefaultModel)
	assert.True(t, cfg.LSP.Enabled)
	assert.True(t, cfg.MCP.Enabled)
	assert.Equal(t, "INFO", cfg.LogLevel)
	assert.True(t, cfg.Features.TUI)
	assert.False(t, cfg.Features.DesktopMode)
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				DefaultModel: "openai/gpt-4",
			},
			wantErr: false,
		},
		{
			name: "missing model",
			cfg: &Config{
				DefaultModel: "",
			},
			wantErr: true,
		},
		{
			name: "enabled provider without key",
			cfg: &Config{
				DefaultModel: "openai/gpt-4",
				Providers: map[string]ProviderConfig{
					"openai": {
						Enabled: true,
						APIKey:  "",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "enabled provider with env var",
			cfg: &Config{
				DefaultModel: "openai/gpt-4",
				Providers: map[string]ProviderConfig{
					"openai": {
						Enabled: true,
						APIKey:  "",
					},
				},
			},
			wantErr: true, // No env var set in test
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEnvOverride(t *testing.T) {
	// Set environment variable
	os.Setenv("LIBCODE_DEFAULT_MODEL", "anthropic/claude")
	defer os.Unsetenv("LIBCODE_DEFAULT_MODEL")

	globalConfig = nil
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "anthropic/claude", cfg.DefaultModel)
}

func TestGet(t *testing.T) {
	// First call should initialize
	cfg1 := Get()
	assert.NotNil(t, cfg1)

	// Second call should return same instance
	cfg2 := Get()
	assert.Same(t, cfg1, cfg2)
}
