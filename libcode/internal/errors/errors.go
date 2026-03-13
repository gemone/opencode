// Package errors provides named error types for libcode
package errors

import (
	"fmt"
)

// Error types for different error categories

// ErrConfig indicates a configuration error
type ErrConfig struct {
	Field   string
	Message string
	Err     error
}

func (e *ErrConfig) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("config error in %s: %s: %v", e.Field, e.Message, e.Err)
	}
	return fmt.Sprintf("config error in %s: %s", e.Field, e.Message)
}

func (e *ErrConfig) Unwrap() error {
	return e.Err
}

// ErrProvider indicates an AI provider error
type ErrProvider struct {
	Provider string
	Message  string
	Err      error
}

func (e *ErrProvider) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("provider error [%s]: %s: %v", e.Provider, e.Message, e.Err)
	}
	return fmt.Sprintf("provider error [%s]: %s", e.Provider, e.Message)
}

func (e *ErrProvider) Unwrap() error {
	return e.Err
}

// ErrTool indicates a tool execution error
type ErrTool struct {
	Tool    string
	Message string
	Err     error
}

func (e *ErrTool) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("tool error [%s]: %s: %v", e.Tool, e.Message, e.Err)
	}
	return fmt.Sprintf("tool error [%s]: %s", e.Tool, e.Message)
}

func (e *ErrTool) Unwrap() error {
	return e.Err
}

// ErrMCP indicates an MCP protocol error
type ErrMCP struct {
	Server  string
	Message string
	Err     error
}

func (e *ErrMCP) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("mcp error [%s]: %s: %v", e.Server, e.Message, e.Err)
	}
	return fmt.Sprintf("mcp error [%s]: %s", e.Server, e.Message)
}

func (e *ErrMCP) Unwrap() error {
	return e.Err
}

// ErrLSP indicates an LSP client error
type ErrLSP struct {
	Language string
	Message  string
	Err      error
}

func (e *ErrLSP) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("lsp error [%s]: %s: %v", e.Language, e.Message, e.Err)
	}
	return fmt.Sprintf("lsp error [%s]: %s", e.Language, e.Message)
}

func (e *ErrLSP) Unwrap() error {
	return e.Err
}

// ErrPTY indicates a PTY operation error
type ErrPTY struct {
	Message string
	Err     error
}

func (e *ErrPTY) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("pty error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("pty error: %s", e.Message)
}

func (e *ErrPTY) Unwrap() error {
	return e.Err
}

// ErrSession indicates a session management error
type ErrSession struct {
	SessionID string
	Message   string
	Err       error
}

func (e *ErrSession) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("session error [%s]: %s: %v", e.SessionID, e.Message, e.Err)
	}
	return fmt.Sprintf("session error [%s]: %s", e.SessionID, e.Message)
}

func (e *ErrSession) Unwrap() error {
	return e.Err
}

// ErrPermission indicates a permission/authorization error
type ErrPermission struct {
	Resource string
	Action   string
	Reason   string
}

func (e *ErrPermission) Error() string {
	return fmt.Sprintf("permission denied: cannot %s %s: %s", e.Action, e.Resource, e.Reason)
}

// ErrValidation indicates an input validation error
type ErrValidation struct {
	Field   string
	Message string
	Value   interface{}
}

func (e *ErrValidation) Error() string {
	return fmt.Sprintf("validation error on %s: %s (value: %v)", e.Field, e.Message, e.Value)
}

// Helper functions to create errors

// ConfigError creates a new configuration error
func ConfigError(field, message string, err error) *ErrConfig {
	return &ErrConfig{Field: field, Message: message, Err: err}
}

// ProviderError creates a new provider error
func ProviderError(provider, message string, err error) *ErrProvider {
	return &ErrProvider{Provider: provider, Message: message, Err: err}
}

// ToolError creates a new tool error
func ToolError(tool, message string, err error) *ErrTool {
	return &ErrTool{Tool: tool, Message: message, Err: err}
}

// MCPError creates a new MCP error
func MCPError(server, message string, err error) *ErrMCP {
	return &ErrMCP{Server: server, Message: message, Err: err}
}

// LSPError creates a new LSP error
func LSPError(language, message string, err error) *ErrLSP {
	return &ErrLSP{Language: language, Message: message, Err: err}
}

// PTYError creates a new PTY error
func PTYError(message string, err error) *ErrPTY {
	return &ErrPTY{Message: message, Err: err}
}

// SessionError creates a new session error
func SessionError(sessionID, message string, err error) *ErrSession {
	return &ErrSession{SessionID: sessionID, Message: message, Err: err}
}

// PermissionError creates a new permission error
func PermissionError(resource, action, reason string) *ErrPermission {
	return &ErrPermission{Resource: resource, Action: action, Reason: reason}
}

// ValidationError creates a new validation error
func ValidationError(field, message string, value interface{}) *ErrValidation {
	return &ErrValidation{Field: field, Message: message, Value: value}
}

// IsPermissionError checks if an error is a permission error
func IsPermissionError(err error) bool {
	_, ok := err.(*ErrPermission)
	return ok
}

// IsConfigError checks if an error is a configuration error
func IsConfigError(err error) bool {
	_, ok := err.(*ErrConfig)
	return ok
}
