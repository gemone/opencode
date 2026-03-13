// Package framework provides the tool registry system
package framework

import (
	"fmt"
	"strings"
	"sync"
)

// Registry manages tool registration and lookup
type Registry struct {
	mu     sync.RWMutex
	tools  map[string]Tool
	alias  map[string]string // alias -> tool ID
	groups map[string][]string // group -> tool IDs
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools:  make(map[string]Tool),
		alias:  make(map[string]string),
		groups: make(map[string][]string),
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(tool Tool) error {
	if tool == nil {
		return fmt.Errorf("cannot register nil tool")
	}

	id := tool.ID()
	if id == "" {
		return fmt.Errorf("tool must have a non-empty ID")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[id]; exists {
		return fmt.Errorf("tool already registered: %s", id)
	}

	r.tools[id] = tool
	return nil
}

// RegisterAlias adds an alias for a tool
func (r *Registry) RegisterAlias(alias, toolID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[toolID]; !exists {
		return fmt.Errorf("tool not found: %s", toolID)
	}

	r.alias[alias] = toolID
	return nil
}

// RegisterGroup adds tools to a group
func (r *Registry) RegisterGroup(group string, toolIDs ...string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Verify all tools exist
	for _, toolID := range toolIDs {
		if _, exists := r.tools[toolID]; !exists {
			return fmt.Errorf("tool not found: %s", toolID)
		}
	}

	r.groups[group] = append(r.groups[group], toolIDs...)
	return nil
}

// Get retrieves a tool by ID or alias
func (r *Registry) Get(id string) (Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check direct ID
	if tool, ok := r.tools[id]; ok {
		return tool, nil
	}

	// Check alias
	if toolID, ok := r.alias[id]; ok {
		if tool, ok := r.tools[toolID]; ok {
			return tool, nil
		}
	}

	return nil, fmt.Errorf("tool not found: %s", id)
}

// List returns all registered tool IDs
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.tools))
	for id := range r.tools {
		ids = append(ids, id)
	}
	return ids
}

// ListByGroup returns tool IDs in a group
func (r *Registry) ListByGroup(group string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if toolIDs, ok := r.groups[group]; ok {
		result := make([]string, len(toolIDs))
		copy(result, toolIDs)
		return result
	}
	return nil
}

// Find searches for tools by name pattern
func (r *Registry) Find(pattern string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []string
	pattern = strings.ToLower(pattern)

	for id := range r.tools {
		if strings.Contains(strings.ToLower(id), pattern) {
			results = append(results, id)
		}
	}

	// Also check aliases
	for alias, toolID := range r.alias {
		if strings.Contains(strings.ToLower(alias), pattern) {
			if _, ok := r.tools[toolID]; ok {
				results = append(results, toolID)
			}
		}
	}

	return results
}

// Count returns the number of registered tools
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}

// Unregister removes a tool from the registry
func (r *Registry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[id]; !exists {
		return fmt.Errorf("tool not found: %s", id)
	}

	delete(r.tools, id)

	// Remove any aliases pointing to this tool
	for alias, toolID := range r.alias {
		if toolID == id {
			delete(r.alias, alias)
		}
	}

	// Remove from groups
	for group, toolIDs := range r.groups {
		var filtered []string
		for _, toolID := range toolIDs {
			if toolID != id {
				filtered = append(filtered, toolID)
			}
		}
		r.groups[group] = filtered
	}

	return nil
}

// Clear removes all tools from the registry
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tools = make(map[string]Tool)
	r.alias = make(map[string]string)
	r.groups = make(map[string][]string)
}

// GetToolSchema returns the schema for a tool
func (r *Registry) GetToolSchema(id string) (*Schema, error) {
	tool, err := r.Get(id)
	if err != nil {
		return nil, err
	}
	return tool.Parameters(), nil
}

// ValidateToolParams validates parameters for a specific tool
func (r *Registry) ValidateToolParams(id string, params map[string]any) error {
	tool, err := r.Get(id)
	if err != nil {
		return err
	}
	return ValidateParams(params, tool.Parameters())
}

// DefaultRegistry is the default global registry
var DefaultRegistry = NewRegistry()
