// Package builtin provides built-in tool implementations
package builtin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gemone/libcode/internal/tool/framework"
)

const (
	maxToolLines = 2000
	maxToolBytes = 50 * 1024
)

// TruncationTool handles output truncation for large tool outputs
type TruncationTool struct {
	workingDir string
	outputDir  string
}

// NewTruncationTool creates a new truncation tool
func NewTruncationTool(workingDir string) *TruncationTool {
	outputDir := filepath.Join(workingDir, ".opencode", "tool-output")
	return &TruncationTool{
		workingDir: workingDir,
		outputDir:  outputDir,
	}
}

// ID returns the tool identifier
func (t *TruncationTool) ID() string {
	return "truncation"
}

// Description returns the tool description
func (t *TruncationTool) Description() string {
	return "Handle output truncation for large tool outputs. Saves full output to disk and returns a preview."
}

// Parameters returns the parameter schema
func (t *TruncationTool) Parameters() *framework.Schema {
	return framework.NewSchema().
		AddProperty("text", framework.Property{
			Type:        "string",
			Description: "The text content to potentially truncate",
			Required:    true,
		}).
		AddProperty("maxLines", framework.Property{
			Type:        "integer",
			Description: "Maximum number of lines to keep",
			Default:     maxToolLines,
			Min:         ptr(1.0),
		}).
		AddProperty("maxBytes", framework.Property{
			Type:        "integer",
			Description: "Maximum number of bytes to keep",
			Default:     maxToolBytes,
			Min:         ptr(1.0),
		}).
		AddProperty("direction", framework.Property{
			Type:        "string",
			Description: "Direction: 'head' (keep start) or 'tail' (keep end)",
			Default:     "head",
			Enum:        []string{"head", "tail"},
		})
}

// Execute truncates the text if needed
func (t *TruncationTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
	text, _ := params["text"].(string)
	maxLines := int(params["maxLines"].(float64))
	maxBytes := int(params["maxBytes"].(float64))
	direction, _ := params["direction"].(string)

	if direction == "" {
		direction = "head"
	}

	lines := strings.Split(text, "\n")
	totalBytes := len(text)

	// Check if truncation is needed
	if len(lines) <= maxLines && totalBytes <= maxBytes {
		return &framework.Result{
			Title:  "No truncation needed",
			Output: text,
			Metadata: map[string]any{
				"truncated": false,
			},
		}, nil
	}

	// Truncate
	var out []string
	var bytes int
	hitBytes := false

	if direction == "head" {
		for i := 0; i < len(lines) && i < maxLines; i++ {
			size := len(lines[i]) + 1
			if bytes+size > maxBytes {
				hitBytes = true
				break
			}
			out = append(out, lines[i])
			bytes += size
		}
	} else {
		for i := len(lines) - 1; i >= 0 && len(out) < maxLines; i-- {
			size := len(lines[i]) + 1
			if bytes+size > maxBytes {
				hitBytes = true
				break
			}
			out = append([]string{lines[i]}, out...)
			bytes += size
		}
	}

	// Create output directory
	if err := os.MkdirAll(t.outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Save full output to file
	outputFile := filepath.Join(t.outputDir, fmt.Sprintf("tool_%d.txt", getCurrentTimestamp()))
	if err := os.WriteFile(outputFile, []byte(text), 0644); err != nil {
		return nil, fmt.Errorf("failed to save full output: %w", err)
	}

	// Build preview
	preview := strings.Join(out, "\n")
	var removed int
	var unit string

	if hitBytes {
		removed = totalBytes - bytes
		unit = "bytes"
	} else {
		removed = len(lines) - len(out)
		unit = "lines"
	}

	message := fmt.Sprintf(
		"The tool call succeeded but the output was truncated. Full output saved to: %s\nUse Grep to search the full content or Read with offset/limit to view specific sections.",
		outputFile,
	)

	if direction == "head" {
		message = fmt.Sprintf("%s\n\n...%d %s truncated...\n\n%s", preview, removed, unit, message)
	} else {
		message = fmt.Sprintf("...%d %s truncated...\n\n%s\n\n%s", removed, unit, message, preview)
	}

	return &framework.Result{
		Title:  "Output truncated",
		Output: message,
		Metadata: map[string]any{
			"truncated": true,
			"outputPath": outputFile,
			"removed":    removed,
			"unit":       unit,
		},
	}, nil
}

// getCurrentTimestamp returns the current Unix timestamp in milliseconds
func getCurrentTimestamp() int64 {
	return time.Now().UnixMilli()
}
