// Package provider provides a registry for AI providers
package provider

import (
	"fmt"
	"sync"

	"github.com/gemone/libcode/internal/ai"
	"github.com/gemone/libcode/internal/ai/provider/openai"
)

// Registry manages available AI providers
type Registry struct {
	mu        sync.RWMutex
	providers map[string]ai.Provider
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]ai.Provider),
	}
}

// Register adds a provider to the registry
func (r *Registry) Register(name string, provider ai.Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[name] = provider
}

// Get retrieves a provider by name
func (r *Registry) Get(name string) (ai.Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return provider, nil
}

// List returns all registered providers
func (r *Registry) List() map[string]ai.Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]ai.Provider, len(r.providers))
	for k, v := range r.providers {
		result[k] = v
	}
	return result
}

// InitializeFromConfig initializes providers from configuration
func (r *Registry) InitializeFromConfig(config map[string]ProviderConfig) error {
	for name, cfg := range config {
		if !cfg.Enabled {
			continue
		}

		var provider ai.Provider
		var err error

		switch name {
		case "openai":
			provider = openai.NewProvider(openai.Config{
				APIKey:  cfg.APIKey,
				BaseURL: cfg.BaseURL,
			})
		// Add more providers here
		// case "anthropic":
		//     provider = anthropic.NewProvider(...)
		default:
			return fmt.Errorf("unknown provider: %s", name)
		}

		if err != nil {
			return fmt.Errorf("failed to initialize provider %s: %w", name, err)
		}

		r.Register(name, provider)
	}

	return nil
}

// ProviderConfig holds configuration for a provider
type ProviderConfig struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
	Enabled bool   `json:"enabled"`
	Model   string `json:"model"`
}

// DefaultRegistry returns a registry with default providers
func DefaultRegistry() *Registry {
	reg := NewRegistry()

	// Initialize OpenAI from environment
	openaiProvider := openai.NewProviderFromEnv()
	if openaiProvider != nil {
		reg.Register("openai", openaiProvider)
	}

	return reg
}
