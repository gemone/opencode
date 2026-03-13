// Package builtin provides built-in tool implementations
package builtin

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gemone/libcode/internal/tool/framework"
)

// LSPTool provides Language Server Protocol operations
type LSPTool struct {
	workingDir string
}

// NewLSPTool creates a new LSP tool
func NewLSPTool(workingDir string) *LSPTool {
	return &LSPTool{workingDir: workingDir}
}

// ID returns the tool identifier
func (t *LSPTool) ID() string {
	return "lsp"
}

// Description returns the tool description
func (t *LSPTool) Description() string {
	return "Interact with Language Server Protocol for code intelligence features like go-to-definition, find-references, and diagnostics."
}

// Parameters returns the parameter schema
func (t *LSPTool) Parameters() *framework.Schema {
	return framework.NewSchema().
		AddProperty("operation", framework.Property{
			Type:        "string",
			Description: "LSP operation to perform",
			Required:    true,
			Enum:        []string{"go_to_definition", "find_references", "hover", "diagnostics", "document_symbols"},
		}).
		AddProperty("filePath", framework.Property{
			Type:        "string",
			Description: "Path to the file to analyze",
			Required:    true,
			Format:      "uri",
		}).
		AddProperty("line", framework.Property{
			Type:        "integer",
			Description: "Line number (1-based)",
			Required:    false,
		}).
		AddProperty("character", framework.Property{
			Type:        "integer",
			Description: "Character offset (1-based)",
			Required:    false,
		})
}

// Execute performs LSP operations
func (t *LSPTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
	operation, _ := params["operation"].(string)
	filePath, _ := params["filePath"].(string)
	line := int(params["line"].(float64))
	character := int(params["character"].(float64))

	if operation == "" {
		return nil, fmt.Errorf("operation is required")
	}

	if filePath == "" {
		return nil, fmt.Errorf("filePath is required")
	}

	// Check if gopls is available for Go files
	if strings.HasSuffix(filePath, ".go") {
		goplsPath, err := exec.LookPath("gopls")
		if err != nil {
			return &framework.Result{
				Title:  "LSP not available",
				Output: "gopls (Go language server) not found in PATH. Install it with: go install golang.org/x/tools/gopls@latest",
				Metadata: map[string]any{
					"available": false,
					"reason":    "gopls not found",
				},
			}, nil
		}

		return t.executeGoPLS(ctx, goplsPath, operation, filePath, line, character)
	}

	// For other file types, return a generic response
	return &framework.Result{
		Title:  "LSP operation",
		Output: fmt.Sprintf("LSP operation '%s' on %s\n\nNote: Full LSP support requires language-specific servers to be installed.", operation, filePath),
		Metadata: map[string]any{
			"operation": operation,
			"filePath":  filePath,
			"supported": false,
		},
	}, nil
}

// executeGoPLS executes gopls commands
func (t *LSPTool) executeGoPLS(ctx context.Context, goplsPath, operation, filePath string, line, character int) (*framework.Result, error) {
	var args []string
	var outputTitle string

	switch operation {
	case "go_to_definition":
		args = []string{"definition", filePath, fmt.Sprintf("%d:%d", line, character)}
		outputTitle = "Go to definition"

	case "find_references":
		args = []string{"references", filePath, fmt.Sprintf("%d:%d", line, character)}
		outputTitle = "Find references"

	case "hover":
		args = []string{"hover", filePath, fmt.Sprintf("%d:%d", line, character)}
		outputTitle = "Hover information"

	case "diagnostics":
		args = []string{"check", filePath}
		outputTitle = "Diagnostics"

	case "document_symbols":
		args = []string{"symbols", filePath}
		outputTitle = "Document symbols"

	default:
		return nil, fmt.Errorf("unsupported LSP operation: %s", operation)
	}

	// Execute gopls
	cmd := exec.CommandContext(ctx, goplsPath, args...)
	cmd.Dir = t.workingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("gopls %s failed: %w\nOutput: %s", operation, err, string(output))
	}

	return &framework.Result{
		Title:  outputTitle,
		Output: string(output),
		Metadata: map[string]any{
			"operation": operation,
			"filePath":  filePath,
			"supported": true,
			"server":    "gopls",
		},
	}, nil
}
