// Package builtin provides built-in tool implementations
package builtin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gemone/libcode/internal/tool/framework"
)

// EditTool edits files by replacing strings
type EditTool struct {
	workingDir string
}

// NewEditTool creates a new edit tool
func NewEditTool(workingDir string) *EditTool {
	return &EditTool{workingDir: workingDir}
}

// ID returns the tool identifier
func (t *EditTool) ID() string {
	return "edit"
}

// Description returns the tool description
func (t *EditTool) Description() string {
	return "Edit a file by replacing oldString with newString. Supports multiple replacement strategies for fuzzy matching."
}

// Parameters returns the parameter schema
func (t *EditTool) Parameters() *framework.Schema {
	return framework.NewSchema().
		AddProperty("filePath", framework.Property{
			Type:        "string",
			Description: "The absolute path to the file to modify",
			Required:    true,
			Format:      "uri",
		}).
		AddProperty("oldString", framework.Property{
			Type:        "string",
			Description: "The text to replace",
			Required:    true,
		}).
		AddProperty("newString", framework.Property{
			Type:        "string",
			Description: "The text to replace it with (must be different from oldString)",
			Required:    true,
		}).
		AddProperty("replaceAll", framework.Property{
			Type:        "boolean",
			Description: "Replace all occurrences of oldString (default false)",
			Default:     false,
		})
}

// Execute edits the file
func (t *EditTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
	filePath, _ := params["filePath"].(string)
	oldString, _ := params["oldString"].(string)
	newString, _ := params["newString"].(string)
	replaceAll, _ := params["replaceAll"].(bool)

	if oldString == newString {
		return nil, fmt.Errorf("no changes to apply: oldString and newString are identical")
	}

	// Resolve relative path
	filePath = resolvePath(filePath, t.workingDir)

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	contentStr := string(content)

	// Ask for permission
	diff := simpleDiff(contentStr, oldString, newString)
	if err := execCtx.AskPermission("edit", []string{filePath}, nil, map[string]any{
		"filepath": filePath,
		"diff":     diff,
	}); err != nil {
		return nil, err
	}

	// Apply replacements
	var newContent string
	var replacements int

	if replaceAll {
		if strings.Contains(contentStr, oldString) {
			newContent = strings.ReplaceAll(contentStr, oldString, newString)
			replacements = strings.Count(contentStr, oldString)
		}
	} else {
		idx := strings.Index(contentStr, oldString)
		if idx >= 0 {
			newContent = contentStr[:idx] + newString + contentStr[idx+len(oldString):]
			replacements = 1
		}
	}

	if replacements == 0 {
		return nil, fmt.Errorf("could not find oldString in the file. It must match exactly, including whitespace, indentation, and line endings")
	}

	// Write file
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	output := fmt.Sprintf("Edit applied successfully. Replaced %d occurrence(s).", replacements)

	return &framework.Result{
		Title:  filepath.Base(filePath),
		Output: output,
		Metadata: map[string]any{
			"filepath":     filePath,
			"replacements": replacements,
		},
	}, nil
}

// simpleDiff creates a simple diff representation
func simpleDiff(content, oldString, newString string) string {
	lines := strings.Split(content, "\n")
	oldLines := strings.Split(oldString, "\n")
	newLines := strings.Split(newString, "\n")

	var diff strings.Builder

	// Find the old string
	for i := range lines {
		if i >= len(lines)-len(oldLines) {
			break
		}

		matched := true
		for j, oldLine := range oldLines {
			if i+j >= len(lines) || lines[i+j] != oldLine {
				matched = false
				break
			}
		}

		if matched {
			diff.WriteString("--- a/file.txt\n")
			diff.WriteString("+++ b/file.txt\n")
			for _, oldLine := range oldLines {
				diff.WriteString("-" + oldLine + "\n")
			}
			for _, newLine := range newLines {
				diff.WriteString("+" + newLine + "\n")
			}
			break
		}
	}

	return diff.String()
}
