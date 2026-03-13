// Package builtin provides built-in tool implementations
package builtin

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gemone/libcode/internal/tool/framework"
)

// ExternalDirectoryTool checks if a path is outside the working directory
type ExternalDirectoryTool struct {
	workingDir string
}

// NewExternalDirectoryTool creates a new external directory tool
func NewExternalDirectoryTool(workingDir string) *ExternalDirectoryTool {
	return &ExternalDirectoryTool{workingDir: workingDir}
}

// ID returns the tool identifier
func (t *ExternalDirectoryTool) ID() string {
	return "external-directory"
}

// Description returns the tool description
func (t *ExternalDirectoryTool) Description() string {
	return "Check and request permission for accessing directories outside the working directory."
}

// Parameters returns the parameter schema
func (t *ExternalDirectoryTool) Parameters() *framework.Schema {
	return framework.NewSchema().
		AddProperty("path", framework.Property{
			Type:        "string",
			Description: "The absolute path to check",
			Required:    true,
			Format:      "uri",
		}).
		AddProperty("kind", framework.Property{
			Type:        "string",
			Description: "Kind of path: 'file' or 'directory'",
			Default:     "file",
			Enum:        []string{"file", "directory"},
		}).
		AddProperty("bypass", framework.Property{
			Type:        "boolean",
			Description: "Bypass the check (for internal use)",
			Default:     false,
		})
}

// Execute checks and requests permission for external directory access
func (t *ExternalDirectoryTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
	targetPath, _ := params["path"].(string)
	kind, _ := params["kind"].(string)
	bypass, _ := params["bypass"].(bool)

	if kind == "" {
		kind = "file"
	}

	// Resolve path
	targetPath = resolvePath(targetPath, t.workingDir)

	// Check if already in working directory
	if t.containsPath(targetPath) {
		return &framework.Result{
			Title:  "Path within working directory",
			Output: fmt.Sprintf("Path is within working directory, no external access needed: %s", targetPath),
			Metadata: map[string]any{
				"external": false,
			},
		}, nil
	}

	if bypass {
		return &framework.Result{
			Title:  "External check bypassed",
			Output: fmt.Sprintf("External check bypassed for: %s", targetPath),
			Metadata: map[string]any{
				"external": true,
				"bypassed": true,
			},
		}, nil
	}

	// Get parent directory for permission
	parentDir := targetPath
	if kind == "file" {
		parentDir = filepath.Dir(targetPath)
	}

	// Normalize path for glob pattern
	globPattern := filepath.Join(parentDir, "*")

	// Ask for permission
	if err := execCtx.AskPermission("external_directory", []string{globPattern}, []string{globPattern}, map[string]any{
		"filepath":  targetPath,
		"parentDir": parentDir,
		"kind":      kind,
	}); err != nil {
		return nil, err
	}

	return &framework.Result{
		Title:  "External directory access granted",
		Output: fmt.Sprintf("Permission granted for external directory: %s", parentDir),
		Metadata: map[string]any{
			"external": true,
			"path":     targetPath,
			"kind":     kind,
		},
	}, nil
}

// containsPath checks if a path is within the working directory
func (t *ExternalDirectoryTool) containsPath(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	absWorkingDir, err := filepath.Abs(t.workingDir)
	if err != nil {
		return false
	}

	relPath, err := filepath.Rel(absWorkingDir, absPath)
	if err != nil {
		return false
	}

	// Check if path starts with ".." which would be outside
	if strings.HasPrefix(relPath, "..") {
		return false
	}

	return true
}
