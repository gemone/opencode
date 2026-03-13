// Package builtin provides built-in tool implementations
package builtin

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gemone/libcode/internal/tool/framework"
)

// Todo represents a task item
type Todo struct {
	ID       string   `json:"id"`
	Subject  string   `json:"subject"`
	Status   string   `json:"status"`
	Priority int      `json:"priority"`
	Blocks   []string `json:"blocks,omitempty"`
	BlockedBy []string `json:"blockedBy,omitempty"`
}

// TodoWriteTool writes/updates the todo list
type TodoWriteTool struct {
	workingDir string
}

// NewTodoWriteTool creates a new todo write tool
func NewTodoWriteTool(workingDir string) *TodoWriteTool {
	return &TodoWriteTool{workingDir: workingDir}
}

// ID returns the tool identifier
func (t *TodoWriteTool) ID() string {
	return "todowrite"
}

// Description returns the tool description
func (t *TodoWriteTool) Description() string {
	return "Update the todo list with tasks to track."
}

// Parameters returns the parameter schema
func (t *TodoWriteTool) Parameters() *framework.Schema {
	return framework.NewSchema().
		AddProperty("todos", framework.Property{
			Type:        "array",
			Description: "The updated todo list",
			Required:    true,
		})
}

// Execute updates the todo list
func (t *TodoWriteTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
	rawTodos, _ := params["todos"].([]any)

	// Ask for permission
	if execCtx.PermissionAsker != nil {
		err := execCtx.PermissionAsker.Ask(&framework.PermissionRequest{
			Permission: "todowrite",
			Patterns:   []string{"*"},
			Always:     []string{"*"},
			Metadata:   map[string]any{},
		})
		if err != nil {
			return nil, err
		}
	}

	// Convert todos to JSON
	todos := make([]Todo, 0, len(rawTodos))
	activeCount := 0

	for i, rawTodo := range rawTodos {
		todoMap, ok := rawTodo.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("todo at index %d is not an object", i)
		}

		todo := Todo{
			ID:       fmt.Sprintf("%v", todoMap["id"]),
			Subject:  fmt.Sprintf("%v", todoMap["subject"]),
			Status:   fmt.Sprintf("%v", todoMap["status"]),
			Priority: int(todoMap["priority"].(float64)),
		}

		if blocks, ok := todoMap["blocks"].([]any); ok {
			for _, b := range blocks {
				todo.Blocks = append(todo.Blocks, fmt.Sprintf("%v", b))
			}
		}

		if blockedBy, ok := todoMap["blockedBy"].([]any); ok {
			for _, b := range blockedBy {
				todo.BlockedBy = append(todo.BlockedBy, fmt.Sprintf("%v", b))
			}
		}

		todos = append(todos, todo)
		if todo.Status != "completed" {
			activeCount++
		}
	}

	// Marshal todos to JSON for storage
	todosJSON, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal todos: %w", err)
	}

	return &framework.Result{
		Title:  fmt.Sprintf("%d todos", activeCount),
		Output: string(todosJSON),
		Metadata: map[string]any{
			"activeCount": activeCount,
			"totalTodos": len(todos),
			"todos":      todos,
		},
	}, nil
}

// TodoReadTool reads the current todo list
type TodoReadTool struct {
	workingDir string
}

// NewTodoReadTool creates a new todo read tool
func NewTodoReadTool(workingDir string) *TodoReadTool {
	return &TodoReadTool{workingDir: workingDir}
}

// ID returns the tool identifier
func (t *TodoReadTool) ID() string {
	return "todoread"
}

// Description returns the tool description
func (t *TodoReadTool) Description() string {
	return "Read your todo list to see current tasks."
}

// Parameters returns the parameter schema
func (t *TodoReadTool) Parameters() *framework.Schema {
	return framework.NewSchema()
}

// Execute reads the todo list
func (t *TodoReadTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
	// Ask for permission
	if execCtx.PermissionAsker != nil {
		err := execCtx.PermissionAsker.Ask(&framework.PermissionRequest{
			Permission: "todoread",
			Patterns:   []string{"*"},
			Always:     []string{"*"},
			Metadata:   map[string]any{},
		})
		if err != nil {
			return nil, err
		}
	}

	// Return empty todo list for now (would be stored in session in real implementation)
	todos := []Todo{}

	todosJSON, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal todos: %w", err)
	}

	return &framework.Result{
		Title:  "0 todos",
		Output: string(todosJSON),
		Metadata: map[string]any{
			"activeCount": 0,
			"totalTodos": 0,
			"todos":      todos,
		},
	}, nil
}
