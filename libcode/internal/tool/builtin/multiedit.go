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

// MultiEditTool performs multiple edits in sequence on a file
type MultiEditTool struct {
	workingDir string
}

// NewMultiEditTool creates a new multiedit tool
func NewMultiEditTool(workingDir string) *MultiEditTool {
	return &MultiEditTool{workingDir: workingDir}
}

// ID returns the tool identifier
func (t *MultiEditTool) ID() string {
	return "multiedit"
}

// Description returns the tool description
func (t *MultiEditTool) Description() string {
	return "Perform multiple edits in sequence on a file. More efficient than calling edit multiple times."
}

// Parameters returns the parameter schema
func (t *MultiEditTool) Parameters() *framework.Schema {
	return framework.NewSchema().
		AddProperty("filePath", framework.Property{
			Type:        "string",
			Description: "The absolute path to the file to modify",
			Required:    true,
			Format:      "uri",
		}).
		AddProperty("edits", framework.Property{
			Type:        "array",
			Description: "Array of edit operations to perform sequentially on the file",
			Required:    true,
		})
}

// Execute performs multiple edits on the file
type editOperation struct {
	OldString  string   `json:"oldString"`
	NewString  string   `json:"newString"`
	ReplaceAll bool     `json:"replaceAll,omitempty"`
}

func (t *MultiEditTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
	filePath, _ := params["filePath"].(string)
	rawEdits, _ := params["edits"].([]any)

	if filePath == "" {
		return nil, fmt.Errorf("filePath is required")
	}

	if len(rawEdits) == 0 {
		return nil, fmt.Errorf("at least one edit is required")
	}

	// Resolve path
	filePath = resolvePath(filePath, t.workingDir)

	// Read initial file content
	content, err := readFileContent(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Apply each edit sequentially
	currentContent := content
	var diffOutput strings.Builder

	for i, rawEdit := range rawEdits {
		editMap, ok := rawEdit.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("edit at index %d is not an object", i)
		}

		oldString, _ := editMap["oldString"].(string)
		newString, _ := editMap["newString"].(string)
		replaceAll, _ := editMap["replaceAll"].(bool)

		if oldString == "" {
			return nil, fmt.Errorf("edit at index %d: oldString is required", i)
		}

		// Check if old string exists
		if !strings.Contains(currentContent, oldString) {
			return nil, fmt.Errorf("edit at index %d: oldString not found in file", i)
		}

		// Count occurrences before edit
		occurrences := strings.Count(currentContent, oldString)

		// Apply edit
		var newContent string
		if replaceAll {
			newContent = strings.ReplaceAll(currentContent, oldString, newString)
		} else {
			// Replace only first occurrence
			idx := strings.Index(currentContent, oldString)
			if idx != -1 {
				newContent = currentContent[:idx] + newString + currentContent[idx+len(oldString):]
			} else {
				newContent = currentContent
			}
		}

		// Generate diff for this edit
		diffOutput.WriteString(fmt.Sprintf("--- Edit %d ---\n", i+1))
		if replaceAll && occurrences > 1 {
			diffOutput.WriteString(fmt.Sprintf("Replacing all %d occurrences of:\n", occurrences))
		}
		diffOutput.WriteString(fmt.Sprintf("<<<<< OLD\n%s>>>>>\n\n", oldString))
		diffOutput.WriteString(fmt.Sprintf("<<<<< NEW\n%s>>>>>\n\n", newString))

		currentContent = newContent
	}

	// Write the modified content back
	if err := writeFileContent(filePath, currentContent); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &framework.Result{
		Title:  filepath.Base(filePath),
		Output: fmt.Sprintf("Applied %d edits to %s\n\nDiffs:\n%s", len(rawEdits), filePath, diffOutput.String()),
		Metadata: map[string]any{
			"filePath": filePath,
			"edits":    len(rawEdits),
			"success":  true,
		},
	}, nil
}

// readFileContent reads the content of a file
func readFileContent(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// writeFileContent writes content to a file
func writeFileContent(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
