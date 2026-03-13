// Package framework provides permission handling for tools
package framework

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// PermissionType represents different types of permissions
type PermissionType string

const (
	// PermissionRead allows reading files
	PermissionRead PermissionType = "read"
	// PermissionWrite allows writing files
	PermissionWrite PermissionType = "write"
	// PermissionEdit allows editing files
	PermissionEdit PermissionType = "edit"
	// PermissionBash allows executing shell commands
	PermissionBash PermissionType = "bash"
	// PermissionExternalDirectory allows accessing external directories
	PermissionExternalDirectory PermissionType = "external_directory"
	// PermissionNetwork allows network operations
	PermissionNetwork PermissionType = "network"
	// PermissionLSP allows LSP operations
	PermissionLSP PermissionType = "lsp"
)

// PermissionManager handles permission checks and caching
type PermissionManager struct {
	mu             sync.RWMutex
	granted        map[string][]PermissionGrant
	asker          PermissionAsker
	autoApprove    map[PermissionType]bool
	dangerousPaths map[string]bool
}

// PermissionGrant represents a granted permission
type PermissionGrant struct {
	Permission string   `json:"permission"`
	Patterns   []string `json:"patterns"`
	Always     []string `json:"always"`
}

// NewPermissionManager creates a new permission manager
func NewPermissionManager(asker PermissionAsker) *PermissionManager {
	return &PermissionManager{
		granted:        make(map[string][]PermissionGrant),
		asker:          asker,
		autoApprove:    make(map[PermissionType]bool),
		dangerousPaths: make(map[string]bool),
	}
}

// SetAutoApprove configures auto-approval for a permission type
func (pm *PermissionManager) SetAutoApprove(permission PermissionType, approve bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.autoApprove[permission] = approve
}

// AddDangerousPath marks a path as dangerous
func (pm *PermissionManager) AddDangerousPath(path string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.dangerousPaths[path] = true
}

// Ask requests permission for an operation
func (pm *PermissionManager) Ask(ctx context.Context, sessionID string, req *PermissionRequest) error {
	pm.mu.RLock()
	autoApprove := pm.autoApprove[PermissionType(req.Permission)]
	pm.mu.RUnlock()

	// Auto-approve if configured
	if autoApprove {
		return nil
	}

	// Check if already granted
	if pm.isGranted(sessionID, req) {
		return nil
	}

	// Check for dangerous paths
	if pm.isDangerous(req) {
		return fmt.Errorf("operation contains dangerous path: %v", req.Patterns)
	}

	// Ask the user
	if pm.asker != nil {
		err := pm.asker.Ask(req)
		if err != nil {
			return err
		}
	}

	// Cache the grant
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.granted[sessionID] = append(pm.granted[sessionID], PermissionGrant{
		Permission: req.Permission,
		Patterns:   req.Patterns,
		Always:     req.Always,
	})

	return nil
}

// isGranted checks if a permission has already been granted
func (pm *PermissionManager) isGranted(sessionID string, req *PermissionRequest) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	grants, ok := pm.granted[sessionID]
	if !ok {
		return false
	}

	for _, grant := range grants {
		if grant.Permission != req.Permission {
			continue
		}

		// Check if all patterns are covered
		allCovered := true
		for _, pattern := range req.Patterns {
			if !pm.matchesPattern(pattern, grant.Patterns, grant.Always) {
				allCovered = false
				break
			}
		}

		if allCovered {
			return true
		}
	}

	return false
}

// matchesPattern checks if a pattern matches any of the granted patterns
func (pm *PermissionManager) matchesPattern(pattern string, patterns, always []string) bool {
	// Check explicit patterns
	for _, granted := range patterns {
		if pm.patternMatches(pattern, granted) {
			return true
		}
	}

	// Check always patterns (wildcards)
	for _, wildcard := range always {
		if wildcard == "*" || pm.patternMatches(pattern, wildcard) {
			return true
		}
	}

	return false
}

// patternMatches checks if a pattern matches a target
func (pm *PermissionManager) patternMatches(target, pattern string) bool {
	// Simple glob matching
	if pattern == "*" {
		return true
	}

	if strings.Contains(pattern, "*") {
		// Convert glob to regex-like pattern
		globPattern := strings.ReplaceAll(pattern, "*", ".*")
		return strings.HasPrefix(target, strings.TrimSuffix(globPattern, ".*")) ||
		       strings.Contains(target, strings.Trim(strings.TrimSuffix(globPattern, ".*"), ".*"))
	}

	return target == pattern
}

// isDangerous checks if a request involves dangerous paths
func (pm *PermissionManager) isDangerous(req *PermissionRequest) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, pattern := range req.Patterns {
		if pm.dangerousPaths[pattern] {
			return true
		}
		// Check if pattern is a subpath of dangerous path
		for dangerous := range pm.dangerousPaths {
			if strings.HasPrefix(pattern, dangerous) {
				return true
			}
		}
	}

	return false
}

// Clear clears granted permissions for a session
func (pm *PermissionManager) Clear(sessionID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.granted, sessionID)
}

// ClearAll clears all granted permissions
func (pm *PermissionManager) ClearAll() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.granted = make(map[string][]PermissionGrant)
}

// Common permission patterns

// FilePatterns generates patterns for file access
func FilePatterns(paths ...string) []string {
	return paths
}

// AnyPattern returns a pattern matching anything
func AnyPattern() []string {
	return []string{"*"}
}

// CommandPatterns generates patterns for command execution
func CommandPatterns(commands ...string) []string {
	patterns := make([]string, len(commands))
	for i, cmd := range commands {
		patterns[i] = cmd + " *"
	}
	return patterns
}

// AlwaysCommand returns the "always" patterns for command execution
func AlwaysCommand(command string) []string {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return []string{}
	}
	// Build prefix patterns
	result := make([]string, len(parts))
	for i := range parts {
		result[i] = strings.Join(parts[:i+1], " ")
		if i < len(parts)-1 {
			result[i] += " *"
		}
	}
	return result
}
