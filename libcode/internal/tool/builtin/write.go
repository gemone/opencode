// Package builtin provides built-in tool implementations
package builtin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gemone/libcode/internal/tool/framework"
)

// WriteTool writes content to a file
type WriteTool struct {
	workingDir string
}

// NewWriteTool creates a new write tool
func NewWriteTool(workingDir string) *WriteTool {
	return &WriteTool{workingDir: workingDir}
}

// ID returns the tool identifier
func (t *WriteTool) ID() string {
	return "write"
}

// Description returns the tool description
func (t *WriteTool) Description() string {
	return "Write content to a file. Creates the file if it doesn't exist, overwrites if it does."
}

// Parameters returns the parameter schema
func (t *WriteTool) Parameters() *framework.Schema {
	return framework.NewSchema().
		AddProperty("filePath", framework.Property{
			Type:        "string",
			Description: "The absolute path to the file to write",
			Required:    true,
			Format:      "uri",
		}).
		AddProperty("content", framework.Property{
			Type:        "string",
			Description: "The content to write to the file",
			Required:    true,
		})
}

// Execute writes content to the file
func (t *WriteTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
	filePath, _ := params["filePath"].(string)
	content, _ := params["content"].(string)

	// Resolve relative path
	filePath = resolvePath(filePath, t.workingDir)

	// Check if file exists
	exists := true
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		exists = false
	}

	// Ask for permission
	if err := execCtx.AskPermission("write", []string{filePath}, nil, map[string]any{
		"filepath": filePath,
		"exists":   exists,
	}); err != nil {
		return nil, err
	}

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	output := "Wrote file successfully."
	if exists {
		output = fmt.Sprintf("Updated file successfully.\nFile: %s", filePath)
	} else {
		output = fmt.Sprintf("Created new file.\nFile: %s", filePath)
	}

	return &framework.Result{
		Title:  filepath.Base(filePath),
		Output: output,
		Metadata: map[string]any{
			"filepath": filePath,
			"exists":   exists,
			"size":     len(content),
		},
	}, nil
}
