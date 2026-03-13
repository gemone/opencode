// Package builtin provides built-in tool implementations
package builtin

import (
	"context"
	"fmt"

	"github.com/gemone/libcode/internal/tool/framework"
)

// InvalidTool represents an invalid or error tool call
type InvalidTool struct {
	toolName string
	errorMsg string
}

// NewInvalidTool creates a new invalid tool
func NewInvalidTool(toolName, errorMsg string) *InvalidTool {
	return &InvalidTool{
		toolName: toolName,
		errorMsg: errorMsg,
	}
}

// ID returns the tool identifier
func (t *InvalidTool) ID() string {
	return "invalid"
}

// Description returns the tool description
func (t *InvalidTool) Description() string {
	return "Error tool for invalid tool calls. Do not use directly."
}

// Parameters returns the parameter schema
func (t *InvalidTool) Parameters() *framework.Schema {
	return framework.NewSchema().
		AddProperty("tool", framework.Property{
			Type:        "string",
			Description: "The tool name that was called",
			Required:    true,
		}).
		AddProperty("error", framework.Property{
			Type:        "string",
			Description: "The error message describing what went wrong",
			Required:    true,
		})
}

// Execute returns an error result
func (t *InvalidTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
	tool, _ := params["tool"].(string)
	errorMsg, _ := params["error"].(string)

	return &framework.Result{
		Title:  "Invalid Tool Call",
		Output: fmt.Sprintf("The arguments provided to the tool are invalid: %s", errorMsg),
		Metadata: map[string]any{
			"tool":     tool,
			"error":    errorMsg,
			"original": params,
		},
	}, nil
}
