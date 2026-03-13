// Package toolcall provides tool calling functionality for AI providers
package toolcall

import "encoding/json"

// Tool represents a tool that can be called by the AI
type Tool struct {
	// Name is the tool identifier
	Name string `json:"name"`

	// Description explains what the tool does
	Description string `json:"description"`

	// InputSchema defines the expected parameters (JSON Schema)
	InputSchema json.RawMessage `json:"input_schema"`
}

// Call represents a tool call request
type Call struct {
	// ID uniquely identifies this call
	ID string `json:"id"`

	// Tool is the name of the tool to call
	Tool string `json:"tool"`

	// Input contains the parameters for the tool
	Input json.RawMessage `json:"input"`
}

// CallChunk represents incremental tool call information during streaming
type CallChunk struct {
	// ID uniquely identifies this call
	ID string `json:"id"`

	// Tool is the name of the tool to call
	Tool string `json:"tool,omitempty"`

	// Index is the position in the sequence of tool calls
	Index int `json:"index"`

	// Input contains incremental input data
	Input json.RawMessage `json:"input,omitempty"`

	// Status indicates the current status
	Status string `json:"status"` // "pending", "in_progress", "complete"
}

// Result represents the result of a tool call
type Result struct {
	// ID matches the Call ID
	ID string `json:"id"`

	// Tool is the name of the tool that was called
	Tool string `json:"tool"`

	// Output contains the tool's output
	Output string `json:"output"`

	// Error contains any error that occurred
	Error string `json:"error,omitempty"`

	// Metadata contains additional information
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Registry manages available tools
type Registry struct {
	tools map[string]Tool
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(tool Tool) {
	r.tools[tool.Name] = tool
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// List returns all registered tools
func (r *Registry) List() []Tool {
	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Validate validates a tool call against the tool's schema
func (r *Registry) Validate(call Call) error {
	tool, ok := r.Get(call.Tool)
	if !ok {
		return &ToolError{Tool: call.Tool, Message: "tool not found"}
	}

	// TODO: Validate input against schema
	_ = tool.InputSchema

	return nil
}

// ToolError represents a tool-related error
type ToolError struct {
	Tool    string
	Message string
	Err     error
}

func (e *ToolError) Error() string {
	if e.Err != nil {
		return e.Tool + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Tool + ": " + e.Message
}

func (e *ToolError) Unwrap() error {
	return e.Err
}
