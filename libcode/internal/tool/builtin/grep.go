// Package builtin provides built-in tool implementations
package builtin

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/gemone/libcode/internal/tool/framework"
)

const (
	maxGrepLineLength = 2000
	maxGrepMatches    = 100
)

// GrepTool searches file contents using ripgrep
type GrepTool struct {
	workingDir string
}

// NewGrepTool creates a new grep tool
func NewGrepTool(workingDir string) *GrepTool {
	return &GrepTool{workingDir: workingDir}
}

// ID returns the tool identifier
func (t *GrepTool) ID() string {
	return "grep"
}

// Description returns the tool description
func (t *GrepTool) Description() string {
	return "Search file contents using regex patterns. Uses ripgrep for fast searching through large codebases."
}

// Parameters returns the parameter schema
func (t *GrepTool) Parameters() *framework.Schema {
	return framework.NewSchema().
		AddProperty("pattern", framework.Property{
			Type:        "string",
			Description: "The regex pattern to search for in file contents",
			Required:    true,
		}).
		AddProperty("path", framework.Property{
			Type:        "string",
			Description: "The directory to search in. Defaults to the current working directory.",
			Format:      "uri",
		}).
		AddProperty("include", framework.Property{
			Type:        "string",
			Description: "File pattern to include in the search (e.g., '*.js', '*.{ts,tsx}')",
		})
}

// Execute executes the grep search
func (t *GrepTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
	pattern, _ := params["pattern"].(string)
	searchPath, _ := params["path"].(string)
	include, _ := params["include"].(string)

	if pattern == "" {
		return nil, fmt.Errorf("pattern is required")
	}

	if searchPath == "" {
		searchPath = t.workingDir
	}

	// Resolve relative path
	searchPath = resolvePath(searchPath, t.workingDir)

	// Ask for permission
	if err := execCtx.AskPermission("grep", []string{searchPath}, nil, map[string]any{
		"pattern": pattern,
		"path":    searchPath,
		"include": include,
	}); err != nil {
		return nil, err
	}

	// Check if ripgrep is available
	rgPath, err := exec.LookPath("rg")
	if err != nil {
		return nil, fmt.Errorf("ripgrep (rg) not found in PATH")
	}

	// Build ripgrep arguments
	args := []string{
		"-nH",          // Line number, filename with match
		"--hidden",     // Search hidden files
		"--no-messages", // Suppress error messages
		"--field-match-separator=|",
		"--regexp",
		pattern,
	}

	if include != "" {
		args = append(args, "--glob", include)
	}

	args = append(args, searchPath)

	// Execute ripgrep
	cmd := exec.CommandContext(ctx, rgPath, args...)

	// Create pipe for stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ripgrep: %w", err)
	}

	// Parse output
	type match struct {
		path    string
		lineNum int
		lineText string
		modTime  int64
	}

	var matches []match
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 3)
		if len(parts) < 3 {
			continue
		}

		filePath := parts[0]
		lineNumStr := parts[1]
		lineText := strings.Join(parts[2:], "|")

		lineNum, err := strconv.Atoi(lineNumStr)
		if err != nil {
			continue
		}

		// Get file modification time
		info, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		matches = append(matches, match{
			path:    filePath,
			lineNum: lineNum,
			lineText: lineText,
			modTime:  info.ModTime().Unix(),
		})
	}

	// Wait for command to finish
	err = cmd.Wait()

	// Exit codes: 0 = matches found, 1 = no matches, 2 = errors (but may still have matches)
	if exitError, ok := err.(*exec.ExitError); ok {
		exitCode := exitError.ExitCode()
		if exitCode == 1 || (exitCode == 2 && len(matches) == 0) {
			return &framework.Result{
				Title:  pattern,
				Output: "No matches found",
				Metadata: map[string]any{
					"matches":   0,
					"truncated": false,
				},
			}, nil
		}

		if exitCode > 2 {
			return nil, fmt.Errorf("ripgrep failed with exit code %d", exitCode)
		}
	} else if err != nil {
		return nil, fmt.Errorf("ripgrep failed: %w", err)
	}

	// Sort by modification time (newest first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].modTime > matches[j].modTime
	})

	totalMatches := len(matches)
	truncated := totalMatches > maxGrepMatches
	if truncated {
		matches = matches[:maxGrepMatches]
	}

	if len(matches) == 0 {
		return &framework.Result{
			Title:  pattern,
			Output: "No matches found",
			Metadata: map[string]any{
				"matches":   0,
				"truncated": false,
			},
		}, nil
	}

	// Format output
	var outputLines []string
	outputLines = append(outputLines, fmt.Sprintf("Found %d matches%s", totalMatches, map[bool]string{true: fmt.Sprintf(" (showing first %d)", maxGrepMatches), false: ""}[truncated]))

	currentFile := ""
	for _, match := range matches {
		if currentFile != match.path {
			if currentFile != "" {
				outputLines = append(outputLines, "")
			}
			currentFile = match.path
			outputLines = append(outputLines, fmt.Sprintf("%s:", match.path))
		}

		lineText := match.lineText
		if len(lineText) > maxGrepLineLength {
			lineText = lineText[:maxGrepLineLength] + "..."
		}

		outputLines = append(outputLines, fmt.Sprintf("  Line %d: %s", match.lineNum, lineText))
	}

	if truncated {
		outputLines = append(outputLines, "")
		outputLines = append(outputLines, fmt.Sprintf("(Results truncated: showing %d of %d matches. Consider using a more specific path or pattern.)",
			maxGrepMatches, totalMatches))
	}

	return &framework.Result{
		Title:  pattern,
		Output: strings.Join(outputLines, "\n"),
		Metadata: map[string]any{
			"matches":   totalMatches,
			"truncated": truncated,
		},
	}, nil
}
