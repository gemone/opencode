// Package mcp provides the Model Context Protocol client implementation
package mcp

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gemone/libcode/internal/mcp/protocol"
)

// Manager manages multiple MCP server connections
type Manager struct {
	mu       sync.RWMutex
	clients  map[string]*Client
	statuses map[string]Status
	configs  map[string]*ServerConfig
}

// ServerConfig holds configuration for an MCP server
type ServerConfig struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"`     // "local" or "remote"
	Enabled  bool              `json:"enabled"`
	URL      string            `json:"url,omitempty"`
	Command  []string          `json:"command,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
	Timeout  int               `json:"timeout,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
	OAuth    *OAuthConfig      `json:"oauth,omitempty"`
}

// OAuthConfig holds OAuth configuration for remote servers
type OAuthConfig struct {
	ClientID     string `json:"clientId,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
	Scope        string `json:"scope,omitempty"`
	Disabled     bool   `json:"disabled,omitempty"`
}

// NewManager creates a new MCP manager
func NewManager() *Manager {
	return &Manager{
		clients:  make(map[string]*Client),
		statuses: make(map[string]Status),
		configs:  make(map[string]*ServerConfig),
	}
}

// Add adds or updates an MCP server configuration
func (m *Manager) Add(config *ServerConfig) (Status, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Store config
	m.configs[config.Name] = config

	// If disabled, just update status
	if !config.Enabled {
		m.statuses[config.Name] = StatusDisabled
		return StatusDisabled, nil
	}

	// Create client
	client, err := m.createClient(config)
	if err != nil {
		status := StatusFailed
		if config.Type == "remote" {
			// Check if it's an auth error
			if isAuthError(err) {
				status = StatusNeedsAuth
			} else if isRegistrationError(err) {
				status = StatusNeedsClientRegistration
			}
		}
		m.statuses[config.Name] = status
		return status, err
	}

	m.clients[config.Name] = client
	m.statuses[config.Name] = StatusConnected

	return StatusConnected, nil
}

// Remove removes an MCP server
func (m *Manager) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Close client if exists
	if client, exists := m.clients[name]; exists {
		if err := client.Close(); err != nil {
			return fmt.Errorf("failed to close client: %w", err)
		}
		delete(m.clients, name)
	}

	delete(m.configs, name)
	delete(m.statuses, name)

	return nil
}

// Connect connects to a specific server
func (m *Manager) Connect(name string) (Status, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	config, exists := m.configs[name]
	if !exists {
		return StatusFailed, fmt.Errorf("server not found: %s", name)
	}

	// Close existing connection if any
	if client, exists := m.clients[name]; exists {
		client.Close()
		delete(m.clients, name)
	}

	// Create and connect new client
	client, err := m.createClient(config)
	if err != nil {
		status := StatusFailed
		if isAuthError(err) {
			status = StatusNeedsAuth
		} else if isRegistrationError(err) {
			status = StatusNeedsClientRegistration
		}
		m.statuses[name] = status
		return status, err
	}

	m.clients[name] = client
	m.statuses[name] = StatusConnected

	return StatusConnected, nil
}

// Disconnect disconnects from a specific server
func (m *Manager) Disconnect(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, exists := m.clients[name]
	if !exists {
		return fmt.Errorf("server not connected: %s", name)
	}

	if err := client.Close(); err != nil {
		return fmt.Errorf("failed to close client: %w", err)
	}

	delete(m.clients, name)
	m.statuses[name] = StatusDisabled

	return nil
}

// Status returns the status of all configured servers
func (m *Manager) Status() map[string]Status {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]Status)
	for name, status := range m.statuses {
		result[name] = status
	}

	// Include all configured servers
	for name := range m.configs {
		if _, exists := result[name]; !exists {
			result[name] = StatusDisabled
		}
	}

	return result
}

// GetClient returns a connected client by name
func (m *Manager) GetClient(name string) (*Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, exists := m.clients[name]
	if !exists {
		return nil, fmt.Errorf("server not connected: %s", name)
	}

	return client, nil
}

// ListTools returns all tools from all connected servers
func (m *Manager) ListTools(ctx context.Context) (map[string]*protocol.Tool, error) {
	// Snapshot clients under lock
	m.mu.RLock()
	clients := make(map[string]*Client, len(m.clients))
	for name, client := range m.clients {
		if m.statuses[name] == StatusConnected {
			clients[name] = client
		}
	}
	m.mu.RUnlock()

	// Parallel execution
	result := make(map[string]*protocol.Tool)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for name, client := range clients {
		wg.Add(1)
		go func(name string, client *Client) {
			defer wg.Done()
			tools, err := client.ListTools(ctx)
			if err != nil {
				return
			}
			mu.Lock()
			for _, tool := range tools.Tools {
				sanitizedName := fmt.Sprintf("%s_%s", sanitizeName(name), sanitizeName(tool.Name))
				result[sanitizedName] = &tool
			}
			mu.Unlock()
		}(name, client)
	}
	wg.Wait()

	return result, nil
}

// ListPrompts returns all prompts from all connected servers
func (m *Manager) ListPrompts(ctx context.Context) (map[string]*protocol.Prompt, error) {
	// Snapshot clients under lock
	m.mu.RLock()
	clients := make(map[string]*Client, len(m.clients))
	for name, client := range m.clients {
		if m.statuses[name] == StatusConnected {
			clients[name] = client
		}
	}
	m.mu.RUnlock()

	// Parallel execution
	result := make(map[string]*protocol.Prompt)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for name, client := range clients {
		wg.Add(1)
		go func(name string, client *Client) {
			defer wg.Done()
			prompts, err := client.ListPrompts(ctx)
			if err != nil {
				return
			}
			mu.Lock()
			for _, prompt := range prompts.Prompts {
				sanitizedName := fmt.Sprintf("%s:%s", sanitizeName(name), sanitizeName(prompt.Name))
				result[sanitizedName] = &prompt
			}
			mu.Unlock()
		}(name, client)
	}
	wg.Wait()

	return result, nil
}

// ListResources returns all resources from all connected servers
func (m *Manager) ListResources(ctx context.Context) (map[string]*protocol.Resource, error) {
	// Snapshot clients under lock
	m.mu.RLock()
	clients := make(map[string]*Client, len(m.clients))
	for name, client := range m.clients {
		if m.statuses[name] == StatusConnected {
			clients[name] = client
		}
	}
	m.mu.RUnlock()

	// Parallel execution
	result := make(map[string]*protocol.Resource)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for name, client := range clients {
		wg.Add(1)
		go func(name string, client *Client) {
			defer wg.Done()
			resources, err := client.ListResources(ctx)
			if err != nil {
				return
			}
			mu.Lock()
			for _, resource := range resources.Resources {
				sanitizedName := fmt.Sprintf("%s:%s", sanitizeName(name), sanitizeName(resource.Name))
				result[sanitizedName] = &resource
			}
			mu.Unlock()
		}(name, client)
	}
	wg.Wait()

	return result, nil
}

// Close closes all connections
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for name, client := range m.clients {
		if err := client.Close(); err != nil {
			lastErr = err
		}
		delete(m.clients, name)
	}

	m.clients = make(map[string]*Client)
	m.statuses = make(map[string]Status)

	return lastErr
}

// createClient creates a client for the given server configuration
func (m *Manager) createClient(config *ServerConfig) (*Client, error) {
	client := NewClient(&ClientConfig{
		Name:    "opencode",
		Version: "0.1.0",
	})

	var tp Transport

	switch config.Type {
	case "local":
		if len(config.Command) == 0 {
			return nil, fmt.Errorf("local server requires command")
		}
		tp = NewStdioTransport(&StdioConfig{
			Command: config.Command[0],
			Args:    config.Command[1:],
		})

	case "remote":
		if config.URL == "" {
			return nil, fmt.Errorf("remote server requires URL")
		}

		// Try SSE first
		tp = NewSSETransport(&SSEConfig{
			URL:     config.URL,
			Headers: config.Headers,
		})

	default:
		return nil, fmt.Errorf("unsupported server type: %s", config.Type)
	}

	// Connect with timeout
	ctx := context.Background()
	timeout := 30 // default 30 seconds
	if config.Timeout > 0 {
		timeout = config.Timeout
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	if err := client.Connect(ctx, tp); err != nil {
		return nil, err
	}

	return client, nil
}

// isAuthError checks if an error is an authentication error
func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "unauthorized") || strings.Contains(errStr, "authentication")
}

// isRegistrationError checks if an error is a client registration error
func isRegistrationError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "registration") || strings.Contains(errStr, "client_id")
}

// sanitizeName sanitizes a name for safe use
func sanitizeName(name string) string {
	var b strings.Builder
	b.Grow(len(name))
	for _, ch := range name {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-' {
			b.WriteRune(ch)
		} else {
			b.WriteByte('_')
		}
	}
	return b.String()
}
