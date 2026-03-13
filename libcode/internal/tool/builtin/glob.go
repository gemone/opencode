// Package builtin provides built-in tool implementations
package builtin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gemone/libcode/internal/tool/framework"
)

// GlobTool finds files matching a pattern
type GlobTool struct {
	workingDir string
}

// NewGlobTool creates a new glob tool
func NewGlobTool(workingDir string) *GlobTool {
	return &GlobTool{workingDir: workingDir}
}

// ID returns the tool identifier
func (t *GlobTool) ID() string {
	return "glob"
}

// Description returns the tool description
func (t *GlobTool) Description() string {
	return "Find files matching a glob pattern (e.g., '**/*.go', '*.txt'). Returns up to 100 results sorted by modification time."
}

// Parameters returns the parameter schema
func (t *GlobTool) Parameters() *framework.Schema {
	return framework.NewSchema().
		AddProperty("pattern", framework.Property{
			Type:        "string",
			Description: "The glob pattern to match files against",
			Required:    true,
		}).
		AddProperty("path", framework.Property{
			Type:        "string",
			Description: "The directory to search in. Defaults to working directory if omitted",
			Format:      "uri",
		})
}

// Execute finds files matching the pattern
func (t *GlobTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
	pattern, _ := params["pattern"].(string)
	searchPath, _ := params["path"].(string)

	if searchPath == "" {
		searchPath = t.workingDir
	}

	// Resolve relative path
	searchPath = resolvePath(searchPath, t.workingDir)

	// Ask for permission
	if err := execCtx.AskPermission("glob", []string{searchPath}, nil, map[string]any{
		"pattern": pattern,
		"path":    searchPath,
	}); err != nil {
		return nil, err
	}

	// Find files
	const limit = 100
	var files []fileInfo

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip directories
		if info.IsDir() {
			// Skip common ignore directories
			if shouldSkipDir(filepath.Base(path)) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check glob match
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil {
			return nil
		}
		if !matched {
			// Try relative path match
			relPath, _ := filepath.Rel(searchPath, path)
			matched, _ = filepath.Match(pattern, relPath)
			if !matched {
				return nil
			}
		}

		// Add file
		files = append(files, fileInfo{
			path: path,
			mtime: info.ModTime().Unix(),
		})

		return nil
	})

	if err != nil && err != context.Canceled {
		return nil, err
	}

	// Sort by modification time (newest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].mtime > files[j].mtime
	})

	// Truncate results
	truncated := len(files) > limit
	if len(files) > limit {
		files = files[:limit]
	}

	// Build output
	var output []string
	if len(files) == 0 {
		output = append(output, "No files found")
	} else {
		for _, f := range files {
			output = append(output, f.path)
		}
		if truncated {
			output = append(output, "")
			output = append(output, fmt.Sprintf("(Results are truncated: showing first %d results. Consider using a more specific path or pattern.)", limit))
		}
	}

	return &framework.Result{
		Title:  filepath.Base(searchPath),
		Output: strings.Join(output, "\n"),
		Metadata: map[string]any{
			"count":     len(files),
			"truncated": truncated,
		},
	}, nil
}

type fileInfo struct {
	path string
	mtime int64
}

// shouldSkipDir returns true if a directory should be skipped
func shouldSkipDir(name string) bool {
	skipDirs := map[string]bool{
		"node_modules":     true,
		"__pycache__":      true,
		".git":             true,
		"dist":             true,
		"build":            true,
		"target":           true,
		"vendor":           true,
		"bin":              true,
		"obj":              true,
		".idea":            true,
		".vscode":          true,
		".zig-cache":       true,
		"zig-out":          true,
		".coverage":        true,
		"coverage":         true,
		"tmp":              true,
		"temp":             true,
		".cache":           true,
		"cache":            true,
		"logs":             true,
		".venv":            true,
		"venv":             true,
		"env":              true,
		".vite":            true,
		"next":             true,
		".next":            true,
		"out":              true,
	}
	return skipDirs[name]
}
