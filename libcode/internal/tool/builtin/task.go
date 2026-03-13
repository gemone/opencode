// Package builtin provides built-in tool implementations
package builtin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gemone/libcode/internal/tool/framework"
)

// TaskTool spawns subagent tasks for parallel execution
type TaskTool struct {
	workingDir string
}

// NewTaskTool creates a new task tool
func NewTaskTool(workingDir string) *TaskTool {
	return &TaskTool{workingDir: workingDir}
}

// ID returns the tool identifier
func (t *TaskTool) ID() string {
	return "task"
}

// Description returns the tool description
func (t *TaskTool) Description() string {
	return "Spawn subagent tasks for parallel execution. Creates isolated sessions for specialized agents to work on specific tasks."
}

// Parameters returns the parameter schema
func (t *TaskTool) Parameters() *framework.Schema {
	return framework.NewSchema().
		AddProperty("description", framework.Property{
			Type:        "string",
			Description: "A short (3-5 words) description of the task",
			Required:    true,
		}).
		AddProperty("prompt", framework.Property{
			Type:        "string",
			Description: "The task for the agent to perform",
			Required:    true,
		}).
		AddProperty("subagent_type", framework.Property{
			Type:        "string",
			Description: "The type of specialized agent to use for this task (e.g., 'executor', 'planner', 'architect')",
			Required:    true,
		}).
		AddProperty("task_id", framework.Property{
			Type:        "string",
			Description: "Optional: resume a previous task by providing its task_id. Creates a fresh session if omitted.",
			Required:    false,
		}).
		AddProperty("command", framework.Property{
			Type:        "string",
			Description: "Optional: the command that triggered this task",
			Required:    false,
		})
}

// Execute spawns a subagent task
func (t *TaskTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
	description, _ := params["description"].(string)
	prompt, _ := params["prompt"].(string)
	subagentType, _ := params["subagent_type"].(string)
	taskID, _ := params["task_id"].(string)
	command, _ := params["command"].(string)

	// Validate required parameters
	if description == "" {
		return nil, fmt.Errorf("description is required")
	}
	if prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}
	if subagentType == "" {
		return nil, fmt.Errorf("subagent_type is required")
	}

	// Ask for task permission
	if execCtx.PermissionAsker != nil {
		err := execCtx.PermissionAsker.Ask(&framework.PermissionRequest{
			Permission: "task",
			Patterns:   []string{subagentType},
			Always:     []string{"*"},
			Metadata: map[string]any{
				"description":   description,
				"subagent_type": subagentType,
				"task_id":       taskID,
				"command":       command,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	// Generate a session ID for this task
	sessionID := taskID
	if sessionID == "" {
		// In full implementation, this would create a new session
		sessionID = fmt.Sprintf("task-%d", time.Now().UnixNano())
	}

	// In full implementation, this would:
	// 1. Create an isolated session with parentID = execCtx.SessionID
	// 2. Configure permissions for the subagent (e.g., deny todowrite, deny task recursion)
	// 3. Execute the prompt using the specified subagent type
	// 4. Return the actual result from the subagent

	// For now, return a formatted result indicating the task was created
	output := []string{
		fmt.Sprintf("task_id: %s (for resuming to continue this task if needed)", sessionID),
		"",
		"<task_result>",
		fmt.Sprintf("Task '%s' delegated to %s subagent", description, subagentType),
		"",
		fmt.Sprintf("Prompt: %s", prompt),
		"",
		"Note: This is a simplified implementation. In the full version,",
		"the subagent would execute the prompt and return its results here.",
		"</task_result>",
	}

	return &framework.Result{
		Title:  description,
		Output: strings.Join(output, "\n"),
		Metadata: map[string]any{
			"sessionId":    sessionID,
			"subagentType": subagentType,
			"description":  description,
			"prompt":       prompt,
			"command":      command,
			"isResumed":    taskID != "",
		},
	}, nil
}
