// Package plugin provides a native Go plugin system for libcode
package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gemone/libcode/internal/tool/framework"
)

// Plugin represents a libcode plugin
type Plugin interface {
	// Name returns the plugin name
	Name() string

	// Version returns the plugin version
	Version() string

	// Description returns the plugin description
	Description() string

	// Init initializes the plugin with the given context
	Init(ctx context.Context, manager *Manager) error

	// Cleanup cleans up plugin resources
	Cleanup() error

	// Tools returns tools provided by this plugin
	Tools() []framework.Tool

	// Hooks returns lifecycle hooks provided by this plugin
	Hooks() *Hooks
}

// Hooks contains lifecycle hook functions
type Hooks struct {
	// EventHook is called when an event occurs
	EventHook func(ctx context.Context, event *Event) error

	// ConfigHook is called when configuration is loaded
	ConfigHook func(ctx context.Context, config map[string]any) error

	// ToolBeforeHook is called before tool execution
	ToolBeforeHook func(ctx context.Context, tool string, callID string, args map[string]any) (map[string]any, error)

	// ToolAfterHook is called after tool execution
	ToolAfterHook func(ctx context.Context, tool string, callID string, args map[string]any, result *ToolResult) error

	// ShellEnvHook is called to modify shell environment
	ShellEnvHook func(ctx context.Context, cwd string, env map[string]string) (map[string]string, error)

	// PermissionHook is called when permission is needed
	PermissionHook func(ctx context.Context, permission *Permission) (*PermissionDecision, error)
}

// Event represents a plugin event
type Event struct {
	Type    string                 `json:"type"`
	Source  string                 `json:"source"`
	Data    map[string]any         `json:"data"`
	Context map[string]any         `json:"context,omitempty"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Title    string                 `json:"title"`
	Output   string                 `json:"output"`
	Metadata map[string]any         `json:"metadata,omitempty"`
	Error    string                 `json:"error,omitempty"`
}

// Permission represents a permission request
type Permission struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Tool        string                 `json:"tool,omitempty"`
	Resource    string                 `json:"resource,omitempty"`
}

// PermissionDecision represents a permission decision
type PermissionDecision struct {
	Status string                 `json:"status"` // "ask", "deny", "allow"
	Reason string                 `json:"reason,omitempty"`
}

// Manager manages plugin lifecycle
type Manager struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
	tools   *framework.Registry
	hooks   map[string][]*Hooks
}

// NewManager creates a new plugin manager
func NewManager(tools *framework.Registry) *Manager {
	return &Manager{
		plugins: make(map[string]Plugin),
		tools:   tools,
		hooks:   make(map[string][]*Hooks),
	}
}

// Register registers a plugin
func (m *Manager) Register(plugin Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := plugin.Name()
	if _, exists := m.plugins[name]; exists {
		return fmt.Errorf("plugin already registered: %s", name)
	}

	// Initialize plugin
	ctx := context.Background()
	if err := plugin.Init(ctx, m); err != nil {
		return fmt.Errorf("failed to initialize plugin %s: %w", name, err)
	}

	m.plugins[name] = plugin

	// Register tools
	for _, tool := range plugin.Tools() {
		if err := m.tools.Register(tool); err != nil {
			// Cleanup on error
			delete(m.plugins, name)
			plugin.Cleanup()
			return fmt.Errorf("failed to register tool from plugin %s: %w", name, err)
		}
	}

	// Register hooks
	hooks := plugin.Hooks()
	if hooks != nil {
		m.registerHooks(name, hooks)
	}

	return nil
}

// Unregister unregisters a plugin
func (m *Manager) Unregister(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	// Unregister hooks
	m.unregisterHooks(name)

	// Unregister tools
	for _, tool := range plugin.Tools() {
		m.tools.Unregister(tool.ID())
	}

	// Cleanup plugin
	if err := plugin.Cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup plugin %s: %w", name, err)
	}

	delete(m.plugins, name)
	return nil
}

// Get retrieves a plugin by name
func (m *Manager) Get(name string) (Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, exists := m.plugins[name]
	return plugin, exists
}

// List returns all registered plugins
func (m *Manager) List() []Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Plugin, 0, len(m.plugins))
	for _, plugin := range m.plugins {
		result = append(result, plugin)
	}
	return result
}

// TriggerEvent triggers an event hook
func (m *Manager) TriggerEvent(ctx context.Context, event *Event) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, hooks := range m.hooks["event"] {
		if hooks.EventHook != nil {
			if err := hooks.EventHook(ctx, event); err != nil {
				return err
			}
		}
	}
	return nil
}

// TriggerConfig triggers a config hook
func (m *Manager) TriggerConfig(ctx context.Context, config map[string]any) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, hooks := range m.hooks["config"] {
		if hooks.ConfigHook != nil {
			if err := hooks.ConfigHook(ctx, config); err != nil {
				return err
			}
		}
	}
	return nil
}

// TriggerToolBefore triggers a tool before hook
func (m *Manager) TriggerToolBefore(ctx context.Context, tool string, callID string, args map[string]any) (map[string]any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, hooks := range m.hooks["tool.before"] {
		if hooks.ToolBeforeHook != nil {
			newArgs, err := hooks.ToolBeforeHook(ctx, tool, callID, args)
			if err != nil {
				return nil, err
			}
			if newArgs != nil {
				args = newArgs
			}
		}
	}
	return args, nil
}

// TriggerToolAfter triggers a tool after hook
func (m *Manager) TriggerToolAfter(ctx context.Context, tool string, callID string, args map[string]any, result *ToolResult) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, hooks := range m.hooks["tool.after"] {
		if hooks.ToolAfterHook != nil {
			if err := hooks.ToolAfterHook(ctx, tool, callID, args, result); err != nil {
				return err
			}
		}
	}
	return nil
}

// TriggerShellEnv triggers a shell environment hook
func (m *Manager) TriggerShellEnv(ctx context.Context, cwd string, env map[string]string) (map[string]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, hooks := range m.hooks["shell.env"] {
		if hooks.ShellEnvHook != nil {
			newEnv, err := hooks.ShellEnvHook(ctx, cwd, env)
			if err != nil {
				return nil, err
			}
			if newEnv != nil {
				env = newEnv
			}
		}
	}
	return env, nil
}

// TriggerPermission triggers a permission hook
func (m *Manager) TriggerPermission(ctx context.Context, permission *Permission) (*PermissionDecision, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, hooks := range m.hooks["permission"] {
		if hooks.PermissionHook != nil {
			decision, err := hooks.PermissionHook(ctx, permission)
			if err != nil {
				return nil, err
			}
			if decision != nil {
				return decision, nil
			}
		}
	}
	return nil, nil
}

// registerHooks registers hooks for a plugin
func (m *Manager) registerHooks(pluginName string, hooks *Hooks) {
	if hooks.EventHook != nil {
		m.hooks["event"] = append(m.hooks["event"], hooks)
	}
	if hooks.ConfigHook != nil {
		m.hooks["config"] = append(m.hooks["config"], hooks)
	}
	if hooks.ToolBeforeHook != nil {
		m.hooks["tool.before"] = append(m.hooks["tool.before"], hooks)
	}
	if hooks.ToolAfterHook != nil {
		m.hooks["tool.after"] = append(m.hooks["tool.after"], hooks)
	}
	if hooks.ShellEnvHook != nil {
		m.hooks["shell.env"] = append(m.hooks["shell.env"], hooks)
	}
	if hooks.PermissionHook != nil {
		m.hooks["permission"] = append(m.hooks["permission"], hooks)
	}
}

// unregisterHooks unregisters hooks for a plugin
func (m *Manager) unregisterHooks(pluginName string) {
	// Remove hooks registered by this plugin
	for key := range m.hooks {
		// In a real implementation, we'd track which hooks belong to which plugin
		// For now, we clear all hooks (simplified)
		delete(m.hooks, key)
	}
}

// Close closes the plugin manager and cleans up all plugins
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for name, plugin := range m.plugins {
		if err := plugin.Cleanup(); err != nil {
			lastErr = err
		}
		delete(m.plugins, name)
	}

	m.hooks = make(map[string][]*Hooks)
	return lastErr
}

// BasePlugin provides a base implementation for plugins
type BasePlugin struct {
	name        string
	version     string
	description string
	tools       []framework.Tool
	hooks       *Hooks
}

// NewBasePlugin creates a new base plugin
func NewBasePlugin(name, version, description string) *BasePlugin {
	return &BasePlugin{
		name:        name,
		version:     version,
		description: description,
		tools:       []framework.Tool{},
		hooks:       &Hooks{},
	}
}

// Name returns the plugin name
func (p *BasePlugin) Name() string {
	return p.name
}

// Version returns the plugin version
func (p *BasePlugin) Version() string {
	return p.version
}

// Description returns the plugin description
func (p *BasePlugin) Description() string {
	return p.description
}

// Init initializes the plugin
func (p *BasePlugin) Init(ctx context.Context, manager *Manager) error {
	return nil
}

// Cleanup cleans up plugin resources
func (p *BasePlugin) Cleanup() error {
	return nil
}

// Tools returns tools provided by this plugin
func (p *BasePlugin) Tools() []framework.Tool {
	return p.tools
}

// Hooks returns lifecycle hooks
func (p *BasePlugin) Hooks() *Hooks {
	return p.hooks
}

// AddTool adds a tool to the plugin
func (p *BasePlugin) AddTool(tool framework.Tool) {
	p.tools = append(p.tools, tool)
}

// SetEventHook sets the event hook
func (p *BasePlugin) SetEventHook(hook func(ctx context.Context, event *Event) error) {
	p.hooks.EventHook = hook
}

// SetConfigHook sets the config hook
func (p *BasePlugin) SetConfigHook(hook func(ctx context.Context, config map[string]any) error) {
	p.hooks.ConfigHook = hook
}

// SetToolBeforeHook sets the tool before hook
func (p *BasePlugin) SetToolBeforeHook(hook func(ctx context.Context, tool string, callID string, args map[string]any) (map[string]any, error)) {
	p.hooks.ToolBeforeHook = hook
}

// SetToolAfterHook sets the tool after hook
func (p *BasePlugin) SetToolAfterHook(hook func(ctx context.Context, tool string, callID string, args map[string]any, result *ToolResult) error) {
	p.hooks.ToolAfterHook = hook
}

// SetShellEnvHook sets the shell environment hook
func (p *BasePlugin) SetShellEnvHook(hook func(ctx context.Context, cwd string, env map[string]string) (map[string]string, error)) {
	p.hooks.ShellEnvHook = hook
}

// SetPermissionHook sets the permission hook
func (p *BasePlugin) SetPermissionHook(hook func(ctx context.Context, permission *Permission) (*PermissionDecision, error)) {
	p.hooks.PermissionHook = hook
}

// MarshalJSON implements json.Marshaler for Event
func (e *Event) MarshalJSON() ([]byte, error) {
	type Alias Event
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type: "plugin.event",
		Alias: (*Alias)(e),
	})
}
