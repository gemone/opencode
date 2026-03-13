// Package builtin provides built-in tool implementations
package builtin

import (
	"context"

	"github.com/gemone/libcode/internal/tool/framework"
)

// PlanExitTool exits planning mode and transitions to build mode
type PlanExitTool struct {
	workingDir string
}

// NewPlanExitTool creates a new plan exit tool
func NewPlanExitTool(workingDir string) *PlanExitTool {
	return &PlanExitTool{workingDir: workingDir}
}

// ID returns the tool identifier
func (t *PlanExitTool) ID() string {
	return "plan_exit"
}

// Description returns the tool description
func (t *PlanExitTool) Description() string {
	return `Use this tool when you have completed the planning phase and are ready to exit plan agent.

This tool will ask the user if they want to switch to build agent to start implementing the plan.

Call this tool:
- After you have written a complete plan to the plan file
- After you have clarified any questions with the user
- When you are confident the plan is ready for implementation

Do NOT call this tool:
- Before you have created or finalized the plan
- If you still have unanswered questions about the implementation
- If the user has indicated they want to continue planning`
}

// Parameters returns the parameter schema
func (t *PlanExitTool) Parameters() *framework.Schema {
	return framework.NewSchema()
}

// Execute executes the plan exit tool
func (t *PlanExitTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
	// Ask for permission
	if execCtx.PermissionAsker != nil {
		err := execCtx.PermissionAsker.Ask(&framework.PermissionRequest{
			Permission: "plan_exit",
			Patterns:   []string{"*"},
			Always:     []string{"*"},
			Metadata: map[string]any{
				"action": "exit_plan_mode",
			},
		})
		if err != nil {
			return nil, err
		}
	}

	// Return a result indicating the user should be asked about switching to build agent
	// The actual question asking and agent switching will be handled by the system
	// based on the metadata in the result
	return &framework.Result{
		Title:  "Plan complete - ready to switch to build agent",
		Output: `Plan is complete.

The system should ask the user: "Plan is complete. Would you like to switch to the build agent and start implementing?"

Options:
- Yes: Switch to build agent and start implementing the plan
- No: Stay with plan agent to continue refining the plan

If approved, a synthetic user message will be created to switch to the build agent with the instruction: "The plan has been approved, you can now edit files. Execute the plan"`,
		Metadata: map[string]any{
			"action":              "request_agent_switch",
			"target_agent":        "build",
			"question":            "Plan is complete. Would you like to switch to the build agent and start implementing?",
			"confirm_label":       "Yes",
			"confirm_description": "Switch to build agent and start implementing the plan",
			"reject_label":        "No",
			"reject_description":  "Stay with plan agent to continue refining the plan",
			"synthetic_message":   "The plan has been approved, you can now edit files. Execute the plan",
		},
	}, nil
}
