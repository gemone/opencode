// Package builtin provides built-in tool implementations
package builtin

import (
	"context"
	"testing"

	"github.com/gemone/libcode/internal/tool/framework"
)

func TestPlanExitTool_ID(t *testing.T) {
	tool := NewPlanExitTool("/test/dir")
	if tool.ID() != "plan_exit" {
		t.Errorf("Expected ID 'plan_exit', got '%s'", tool.ID())
	}
}

func TestPlanExitTool_Description(t *testing.T) {
	tool := NewPlanExitTool("/test/dir")
	desc := tool.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
	if !contains(desc, "plan") || !contains(desc, "build agent") {
		t.Error("Description should mention plan and build agent")
	}
}

func TestPlanExitTool_Parameters(t *testing.T) {
	tool := NewPlanExitTool("/test/dir")
	params := tool.Parameters()
	if params == nil {
		t.Error("Parameters should not be nil")
	}
}

func TestPlanExitTool_Execute_NoPermissionAsker(t *testing.T) {
	tool := NewPlanExitTool("/test/dir")
	ctx := context.Background()
	execCtx := &framework.ExecutionContext{}

	result, err := tool.Execute(ctx, nil, execCtx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.Title != "Plan complete - ready to switch to build agent" {
		t.Errorf("Unexpected title: %s", result.Title)
	}

	if result.Metadata["action"] != "request_agent_switch" {
		t.Errorf("Expected action 'request_agent_switch', got %v", result.Metadata["action"])
	}

	if result.Metadata["target_agent"] != "build" {
		t.Errorf("Expected target_agent 'build', got %v", result.Metadata["target_agent"])
	}
}

func TestPlanExitTool_Execute_WithPermissionAsker(t *testing.T) {
	tool := NewPlanExitTool("/test/dir")
	ctx := context.Background()

	asker := &mockPermissionAsker{}
	execCtx := &framework.ExecutionContext{
		PermissionAsker: asker,
	}

	result, err := tool.Execute(ctx, nil, execCtx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if !asker.called {
		t.Error("PermissionAsker.Ask should have been called")
	}

	permission := asker.request.Permission
	if permission != "plan_exit" {
		t.Errorf("Expected permission 'plan_exit', got '%s'", permission)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

type mockPermissionAsker struct {
	called  bool
	request *framework.PermissionRequest
}

func (m *mockPermissionAsker) Ask(req *framework.PermissionRequest) error {
	m.called = true
	m.request = req
	return nil
}
