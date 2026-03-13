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

// LsTool lists directory contents
type LsTool struct {
	workingDir string
}

// dirEntry represents a file or directory in the tree
type dirEntry struct {
	name     string
	isDir    bool
	children []*dirEntry
}

// NewLsTool creates a new ls tool
func NewLsTool(workingDir string) *LsTool {
	return &LsTool{workingDir: workingDir}
}

// ID returns the tool identifier
func (t *LsTool) ID() string {
	return "ls"
}

// Description returns the tool description
func (t *LsTool) Description() string {
	return "List files and directories in a tree structure. Ignores common build/cache directories."
}

// Parameters returns the parameter schema
func (t *LsTool) Parameters() *framework.Schema {
	return framework.NewSchema().
		AddProperty("path", framework.Property{
			Type:        "string",
			Description: "The absolute path to the directory to list",
			Format:      "uri",
		}).
		AddProperty("ignore", framework.Property{
			Type:        "array",
			Description: "List of glob patterns to ignore",
			Items:       &framework.Property{Type: "string"},
		})
}

// Execute lists the directory
func (t *LsTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
	searchPath, _ := params["path"].(string)

	if searchPath == "" {
		searchPath = t.workingDir
	}

	// Resolve relative path
	searchPath = resolvePath(searchPath, t.workingDir)

	// Ask for permission
	if err := execCtx.AskPermission("ls", []string{searchPath}, nil, map[string]any{
		"path": searchPath,
	}); err != nil {
		return nil, err
	}

	// Get ignore patterns
	ignore := defaultIgnorePatterns()
	if ignorePatterns, ok := params["ignore"].([]any); ok {
		for _, p := range ignorePatterns {
			if pattern, ok := p.(string); ok {
				ignore = append(ignore, "!"+pattern)
			}
		}
	}

	// Collect files and directories
	const limit = 100

	root := &dirEntry{name: filepath.Base(searchPath), isDir: true, children: nil}

	// Build tree structure
	fileCount := 0

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		relPath, err := filepath.Rel(searchPath, path)
		if err != nil || relPath == "." {
			return nil
		}

		// Skip ignored directories
		if info.IsDir() && shouldSkipDir(filepath.Base(path)) {
			return filepath.SkipDir
		}

		// Check ignore patterns
		if shouldIgnore(relPath, ignore) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Stop if we've reached the limit
		if fileCount >= limit {
			return nil
		}

		// Add to tree
		parts := strings.Split(relPath, string(filepath.Separator))
		addEntry(root, parts, 0, info.IsDir())
		fileCount++

		return nil
	})

	if err != nil && err != context.Canceled {
		return nil, err
	}

	// Build tree output
	var output strings.Builder
	output.WriteString(fmt.Sprintf("%s/\n", searchPath))
	renderTreeEntry(root, "", "", &output)

	truncated := fileCount >= limit

	return &framework.Result{
		Title:  filepath.Base(searchPath),
		Output: output.String(),
		Metadata: map[string]any{
			"count":     fileCount,
			"truncated": truncated,
		},
	}, nil
}

// addEntry adds a file/directory to the tree
func addEntry(parent *dirEntry, parts []string, index int, isDir bool) {
	if index >= len(parts) {
		return
	}

	name := parts[index]

	// Find or create child entry
	var child *dirEntry
	for _, c := range parent.children {
		if c.name == name {
			child = c
			break
		}
	}

	if child == nil {
		child = &dirEntry{
			name:     name,
			isDir:    isDir,
			children: nil,
		}
		parent.children = append(parent.children, child)
	}

	// Recurse for subdirectories
	if index < len(parts)-1 {
		if child.children == nil {
			child.children = make([]*dirEntry, 0)
		}
		addEntry(child, parts, index+1, true)
	}
}

// renderTreeEntry renders a tree entry
func renderTreeEntry(entry *dirEntry, prefix, childPrefix string, output *strings.Builder) {
	// Sort children: directories first, then alphabetically
	sort.Slice(entry.children, func(i, j int) bool {
		// Directories first
		if entry.children[i].isDir != entry.children[j].isDir {
			return entry.children[i].isDir
		}
		return entry.children[i].name < entry.children[j].name
	})

	for i, child := range entry.children {
		isLast := i == len(entry.children)-1

		connector := "├── "
		if isLast {
			connector = "└── "
		}

		suffix := ""
		if child.isDir {
			suffix = "/"
		}

		output.WriteString(prefix + connector + child.name + suffix + "\n")

		// Render children
		if child.isDir && len(child.children) > 0 {
			newPrefix := childPrefix
			if isLast {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
			renderTreeEntry(child, newPrefix, newPrefix, output)
		}
	}
}

// shouldIgnore checks if a path matches any ignore pattern
func shouldIgnore(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.HasPrefix(pattern, "!") {
			// Negation pattern - this is for excluding from ignore
			continue
		}
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
		// Check if path contains pattern
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

// defaultIgnorePatterns returns default ignore patterns
func defaultIgnorePatterns() []string {
	return []string{
		"!node_modules",
		"!__pycache__",
		"!.git",
		"!dist",
		"!build",
		"!target",
		"!vendor",
		"!bin",
		"!obj",
		"!.idea",
		"!.vscode",
		"!.zig-cache",
		"!zig-out",
		"!.coverage",
		"!coverage",
		"!tmp",
		"!temp",
		"!.cache",
		"!cache",
		"!logs",
		"!.venv",
		"!venv",
		"!env",
	}
}
