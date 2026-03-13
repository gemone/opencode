package builtin

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gemone/libcode/internal/tool/framework"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadTool(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "Hello, World!\nLine 2\nLine 3"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	tool := NewReadTool(tmpDir)

	// Test reading a file
	ctx := context.Background()
	params := map[string]any{
		"filePath": testFile,
		"offset":   float64(1),
		"limit":    float64(10),
	}

	result, err := tool.Execute(ctx, params, &framework.ExecutionContext{
		WorkingDir:      tmpDir,
		PermissionAsker: &testAsker{granted: make(map[string]bool)},
	})

	require.NoError(t, err)
	assert.Contains(t, result.Output, "1: Hello, World!")
	assert.Contains(t, result.Output, "2: Line 2")
	assert.Equal(t, "test.txt", result.Title)
}

func TestWriteTool(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewWriteTool(tmpDir)

	ctx := context.Background()
	testFile := filepath.Join(tmpDir, "newfile.txt")
	params := map[string]any{
		"filePath": testFile,
		"content":  "Hello, World!",
	}

	result, err := tool.Execute(ctx, params, &framework.ExecutionContext{
		WorkingDir:      tmpDir,
		PermissionAsker: &testAsker{granted: make(map[string]bool)},
	})

	require.NoError(t, err)
	assert.Contains(t, result.Output, "Created new file")

	// Verify file was created
	content, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", string(content))
}

func TestEditTool(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	originalContent := "Hello World\nGoodbye World\n"
	err := os.WriteFile(testFile, []byte(originalContent), 0644)
	require.NoError(t, err)

	tool := NewEditTool(tmpDir)

	ctx := context.Background()
	params := map[string]any{
		"filePath":  testFile,
		"oldString": "Hello World",
		"newString": "Hello Go",
	}

	result, err := tool.Execute(ctx, params, &framework.ExecutionContext{
		WorkingDir:      tmpDir,
		PermissionAsker: &testAsker{granted: make(map[string]bool)},
	})

	require.NoError(t, err)
	assert.Contains(t, result.Output, "Edit applied successfully")

	// Verify file was edited
	content, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Hello Go")
	assert.NotContains(t, string(content), "Hello World")
}

func TestGlobTool(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "test1.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test2.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("text"), 0644)

	tool := NewGlobTool(tmpDir)

	ctx := context.Background()
	params := map[string]any{
		"pattern": "*.go",
	}

	result, err := tool.Execute(ctx, params, &framework.ExecutionContext{
		WorkingDir:      tmpDir,
		PermissionAsker: &testAsker{granted: make(map[string]bool)},
	})

	require.NoError(t, err)
	assert.Contains(t, result.Output, "test1.go")
	assert.Contains(t, result.Output, "test2.go")
	assert.NotContains(t, result.Output, "test.txt")
}

func TestLsTool(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test structure
	os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "subdir", "file2.txt"), []byte("content2"), 0644)

	tool := NewLsTool(tmpDir)

	ctx := context.Background()
	params := map[string]any{
		"path": tmpDir,
	}

	result, err := tool.Execute(ctx, params, &framework.ExecutionContext{
		WorkingDir:      tmpDir,
		PermissionAsker: &testAsker{granted: make(map[string]bool)},
	})

	require.NoError(t, err)
	assert.Contains(t, result.Output, "file1.txt")
	assert.Contains(t, result.Output, "subdir")
}

func TestToolIDs(t *testing.T) {
	var tool framework.Tool

	tool = NewReadTool("/tmp")
	assert.Equal(t, "read", tool.ID())

	tool = NewWriteTool("/tmp")
	assert.Equal(t, "write", tool.ID())

	tool = NewEditTool("/tmp")
	assert.Equal(t, "edit", tool.ID())

	tool = NewGlobTool("/tmp")
	assert.Equal(t, "glob", tool.ID())

	tool = NewLsTool("/tmp")
	assert.Equal(t, "ls", tool.ID())
}

func TestToolDescriptions(t *testing.T) {
	tools := []struct {
		name        string
		tool        framework.Tool
		hasKeywords bool
	}{
		{"read", NewReadTool("/tmp"), true},
		{"write", NewWriteTool("/tmp"), true},
		{"edit", NewEditTool("/tmp"), true},
		{"glob", NewGlobTool("/tmp"), true},
		{"ls", NewLsTool("/tmp"), true},
	}

	for _, tt := range tools {
		t.Run(tt.name, func(t *testing.T) {
			desc := tt.tool.Description()
			assert.NotEmpty(t, desc)
			if tt.hasKeywords {
				// Check that description contains relevant keywords
				assert.Contains(t, desc, "file")
			}
		})
	}
}

// testAsker is a test implementation of PermissionAsker
type testAsker struct {
	granted map[string]bool
}

func (m *testAsker) Ask(req *framework.PermissionRequest) error {
	for _, pattern := range req.Patterns {
		m.granted[pattern] = true
	}
	return nil
}
