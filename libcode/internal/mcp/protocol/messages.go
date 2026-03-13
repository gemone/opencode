// Package protocol provides the Model Context Protocol types
package protocol

// InitializeParams represents parameters for the initialize request
type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities     ClientCapabilities    `json:"capabilities"`
	ClientInfo       ClientInfo            `json:"clientInfo"`
	Meta             map[string]any        `json:"meta,omitempty"`
}

// InitializeResult represents the result of an initialize request
type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities     ServerCapabilities    `json:"capabilities"`
	ServerInfo       ServerInfo            `json:"serverInfo"`
	Meta             map[string]any        `json:"meta,omitempty"`
}

// ClientCapabilities describes client capabilities
type ClientCapabilities struct {
	Roots      *RootsCapability      `json:"roots,omitempty"`
	Sampling   *SamplingCapability   `json:"sampling,omitempty"`
	Resources  *ResourcesCapability  `json:"resources,omitempty"`
	Tools      *ToolsCapability      `json:"tools,omitempty"`
	Prompts    *PromptsCapability    `json:"prompts,omitempty"`
}

// ServerCapabilities describes server capabilities
type ServerCapabilities struct {
	Roots      *RootsCapability      `json:"roots,omitempty"`
	Sampling   *SamplingCapability   `json:"sampling,omitempty"`
	Resources  *ResourcesCapability  `json:"resources,omitempty"`
	Tools      *ToolsCapability      `json:"tools,omitempty"`
	Prompts    *PromptsCapability    `json:"prompts,omitempty"`
	Logging    *LoggingCapability    `json:"logging,omitempty"`
}

// ClientInfo describes the client
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ServerInfo describes the server
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// RootsCapability describes root listing support
type RootsCapability struct {
	ListChanged bool `json:"listChanged"`
}

// SamplingCapability describes sampling support
type SamplingCapability struct{}

// ResourcesCapability describes resource support
type ResourcesCapability struct {
	Subscribe      bool `json:"subscribe"`
	ListChanged    bool `json:"listChanged"`
}

// ToolsCapability describes tool support
type ToolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

// PromptsCapability describes prompt support
type PromptsCapability struct {
	ListChanged bool `json:"listChanged"`
}

// LoggingCapability describes logging support
type LoggingCapability struct{}

// Tool represents an MCP tool definition
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]any         `json:"inputSchema"`
	Metadata    map[string]any         `json:"metadata,omitempty"`
}

// ListToolsResult is the result of listing tools
type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}

// CallToolParams are parameters for calling a tool
type CallToolParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

// CallToolResult is the result of calling a tool
type CallToolResult struct {
	Content []Content      `json:"content"`
	IsError bool           `json:"isError,omitempty"`
	Meta    map[string]any `json:"meta,omitempty"`
}

// Content represents MCP content
type Content struct {
	Type     string                 `json:"type"`
	Text     string                 `json:"text,omitempty"`
	Data     any                    `json:"data,omitempty"`
	Metadata map[string]any         `json:"metadata,omitempty"`
}

// Resource represents an MCP resource
type Resource struct {
	URI         string            `json:"uri"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	MimeType    string            `json:"mimeType,omitempty"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
}

// ListResourcesResult is the result of listing resources
type ListResourcesResult struct {
	Resources []Resource `json:"resources"`
}

// ReadResourceParams are parameters for reading a resource
type ReadResourceParams struct {
	URI string `json:"uri"`
}

// ReadResourceResult is the result of reading a resource
type ReadResourceResult struct {
	Contents []ResourceContent `json:"contents"`
}

// ResourceContent represents content from a resource
type ResourceContent struct {
	URI      string         `json:"uri"`
	MimeType string         `json:"mimeType,omitempty"`
	Text     string         `json:"text,omitempty"`
	Blob     []byte         `json:"blob,omitempty"`
	Metadata map[string]any `json:"meta,omitempty"`
}

// Prompt represents an MCP prompt template
type Prompt struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Arguments   []PromptArgument       `json:"arguments,omitempty"`
	Metadata    map[string]any         `json:"metadata,omitempty"`
}

// PromptArgument represents a prompt argument
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// ListPromptsResult is the result of listing prompts
type ListPromptsResult struct {
	Prompts []Prompt `json:"prompts"`
}

// GetPromptParams are parameters for getting a prompt
type GetPromptParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

// GetPromptResult is the result of getting a prompt
type GetPromptResult struct {
	Description string         `json:"description,omitempty"`
	Messages    []PromptMessage `json:"messages"`
}

// PromptMessage represents a message in a prompt
type PromptMessage struct {
	Role    string         `json:"role"`
	Content PromptContent  `json:"content"`
}

// PromptContent represents content in a prompt message
type PromptContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// SetLevelParams are parameters for setting logging level
type SetLevelParams struct {
	Level string `json:"level"` // "debug", "info", "notice", "warning", "error", "critical", "alert", "emergency"
}
