// Package framework provides the core tool system infrastructure
package framework

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gemone/libcode/internal/session"
)

// Tool is the interface that all tools must implement
type Tool interface {
	// ID returns the unique identifier for this tool
	ID() string

	// Description returns a human-readable description of what this tool does
	Description() string

	// Parameters returns the schema for tool parameters
	Parameters() *Schema

	// Execute runs the tool with the given parameters and context
	Execute(ctx context.Context, params map[string]any, execCtx *ExecutionContext) (*Result, error)
}

// ExecutionContext provides context for tool execution
type ExecutionContext struct {
	SessionID    string
	MessageID    string
	CallID       string
	Agent        string
	Abort        context.CancelFunc
	Messages     []session.Message
	Extra        map[string]any
	WorkingDir   string
	PermissionAsker
	QuestionAsker
}

// PermissionAsker handles permission requests
type PermissionAsker interface {
	Ask(req *PermissionRequest) error
}

// PermissionRequest represents a permission request
type PermissionRequest struct {
	Permission string   `json:"permission"`
	Patterns   []string `json:"patterns"`
	Always     []string `json:"always"`
	Metadata   any      `json:"metadata"`
}

// QuestionAsker handles question requests to users
type QuestionAsker interface {
	Ask(req *QuestionRequest) ([]string, error)
}

// QuestionRequest represents a question request
type QuestionRequest struct {
	Question string             `json:"question"`
	Header   string             `json:"header"`
	Options  []*QuestionOption  `json:"options"`
	Multiple bool               `json:"multiple,omitempty"`
}

// QuestionOption represents a choice in a question
type QuestionOption struct {
	Label       string `json:"label"`
	Description string `json:"description"`
}

// AskPermission is a convenience method for requesting permissions.
// If always is nil, it defaults to ["*"] (always allow).
func (ec *ExecutionContext) AskPermission(permission string, patterns []string, always []string, metadata map[string]any) error {
	if ec.PermissionAsker == nil {
		return nil
	}
	if always == nil {
		always = []string{"*"}
	}
	return ec.PermissionAsker.Ask(&PermissionRequest{
		Permission: permission,
		Patterns:   patterns,
		Always:     always,
		Metadata:   metadata,
	})
}

// Result represents the result of a tool execution
type Result struct {
	Title       string                 `json:"title"`
	Output      string                 `json:"output"`
	Metadata    map[string]any         `json:"metadata"`
	Attachments []Attachment           `json:"attachments,omitempty"`
	Extra       map[string]any         `json:"extra,omitempty"`
	Diagnostics map[string][]Diagnostic `json:"diagnostics,omitempty"`
}

// Attachment represents a file attachment in tool results
type Attachment struct {
	Type string `json:"type"` // "file"
	MIME string `json:"mime"`
	URL  string `json:"url"`
	Name string `json:"name,omitempty"`
}

// Diagnostic represents a diagnostic error/warning
type Diagnostic struct {
	Range     Range   `json:"range"`
	Severity  int     `json:"severity"` // 1=error, 2=warning, 3=info, 4=hint
	Message   string  `json:"message"`
	Code      *string `json:"code,omitempty"`
	Source    string  `json:"source,omitempty"`
	Tags      []int   `json:"tags,omitempty"`
	Related   []RelatedDiagnostic `json:"relatedInformation,omitempty"`
}

// Range represents a range in a text document
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Position represents a position in a text document
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// RelatedDiagnostic represents a related diagnostic
type RelatedDiagnostic struct {
	Range   Range  `json:"range"`
	Message string `json:"message"`
}

// BaseTool provides a partial implementation of Tool for common functionality
type BaseTool struct {
	id          string
	description string
	schema      *Schema
	executeFunc func(ctx context.Context, params map[string]any, execCtx *ExecutionContext) (*Result, error)
}

// NewBaseTool creates a new base tool
func NewBaseTool(id, description string, schema *Schema, executeFunc func(ctx context.Context, params map[string]any, execCtx *ExecutionContext) (*Result, error)) *BaseTool {
	return &BaseTool{
		id:          id,
		description: description,
		schema:      schema,
		executeFunc: executeFunc,
	}
}

// ID returns the tool's identifier
func (t *BaseTool) ID() string {
	return t.id
}

// Description returns the tool's description
func (t *BaseTool) Description() string {
	return t.description
}

// Parameters returns the tool's parameter schema
func (t *BaseTool) Parameters() *Schema {
	return t.schema
}

// Execute runs the tool
func (t *BaseTool) Execute(ctx context.Context, params map[string]any, execCtx *ExecutionContext) (*Result, error) {
	return t.executeFunc(ctx, params, execCtx)
}

// ValidationError represents a parameter validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error returns the error message
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}

// ValidateParams validates parameters against a schema
func ValidateParams(params map[string]any, schema *Schema) error {
	// Check required fields
	for name, prop := range schema.Properties {
		if prop.Required {
			if _, ok := params[name]; !ok {
				return &ValidationError{
					Field:   name,
					Message: "required field missing",
				}
			}
		}

		// Type validation
		if val, ok := params[name]; ok {
			if err := validateType(name, val, prop.Type); err != nil {
				return err
			}
		}
	}

	// Check for unknown fields
	for name := range params {
		if _, ok := schema.Properties[name]; !ok {
			if !schema.AllowUnknown {
				return &ValidationError{
					Field:   name,
					Message: "unknown parameter",
				}
			}
		}
	}

	return nil
}

func validateType(name string, value any, expectedType string) error {
	if value == nil {
		return nil // nil is valid for optional fields
	}

	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return &ValidationError{
				Field:   name,
				Message: fmt.Sprintf("expected string, got %T", value),
			}
		}
	case "number", "integer":
		switch value.(type) {
		case int, int32, int64, float32, float64:
			// valid
		default:
			return &ValidationError{
				Field:   name,
				Message: fmt.Sprintf("expected number, got %T", value),
			}
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return &ValidationError{
				Field:   name,
				Message: fmt.Sprintf("expected boolean, got %T", value),
			}
		}
	case "array":
		if _, ok := value.([]any); !ok {
			return &ValidationError{
				Field:   name,
				Message: fmt.Sprintf("expected array, got %T", value),
			}
		}
	case "object":
		if _, ok := value.(map[string]any); !ok {
			return &ValidationError{
				Field:   name,
				Message: fmt.Sprintf("expected object, got %T", value),
			}
		}
	}
	return nil
}

// ExecuteWithTimeout executes a tool with a timeout
func ExecuteWithTimeout(ctx context.Context, tool Tool, params map[string]any, execCtx *ExecutionContext, timeout time.Duration) (*Result, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resultChan := make(chan *Result, 1)
	errChan := make(chan error, 1)

	go func() {
		result, err := tool.Execute(ctx, params, execCtx)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- result
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("tool execution timed out after %v", timeout)
	case err := <-errChan:
		return nil, err
	case result := <-resultChan:
		return result, nil
	}
}

// MarshalResult converts a result to JSON
func MarshalResult(result *Result) ([]byte, error) {
	return json.MarshalIndent(result, "", "  ")
}

// UnmarshalParams converts JSON to parameters
func UnmarshalParams(data []byte) (map[string]any, error) {
	var params map[string]any
	if err := json.Unmarshal(data, &params); err != nil {
		return nil, err
	}
	return params, nil
}
