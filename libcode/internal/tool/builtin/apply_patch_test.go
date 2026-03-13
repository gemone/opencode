package builtin

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gemone/libcode/internal/tool/framework"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyPatchTool_AddFile(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewApplyPatchTool(tmpDir)

	ctx := context.Background()
	patchText := `*** Begin Patch
*** Add File: newfile.txt
+Hello, World!
+Line 2
*** End Patch`

	params := map[string]any{
		"patchText": patchText,
	}

	result, err := tool.Execute(ctx, params, &framework.ExecutionContext{
		WorkingDir:      tmpDir,
		PermissionAsker: &testAsker{granted: make(map[string]bool)},
	})

	require.NoError(t, err)
	assert.Contains(t, result.Output, "Success")

	// Verify file was created
	content, err := os.ReadFile(filepath.Join(tmpDir, "newfile.txt"))
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!\nLine 2\n", string(content))
}

func TestApplyPatchTool_UpdateFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	originalContent := "Line 1\nLine 2\nLine 3\n"
	err := os.WriteFile(testFile, []byte(originalContent), 0644)
	require.NoError(t, err)

	tool := NewApplyPatchTool(tmpDir)

	ctx := context.Background()
	patchText := `*** Begin Patch
*** Update File: test.txt
 Line 1
-Line 2
+Updated Line 2
 Line 3
*** End Patch`

	params := map[string]any{
		"patchText": patchText,
	}

	result, err := tool.Execute(ctx, params, &framework.ExecutionContext{
		WorkingDir:      tmpDir,
		PermissionAsker: &testAsker{granted: make(map[string]bool)},
	})

	require.NoError(t, err)
	assert.Contains(t, result.Output, "Success")

	// Verify file was updated
	content, err := os.ReadFile(testFile)
	require.NoError(t, err)
	contentStr := string(content)
	assert.Contains(t, contentStr, "Updated Line 2")
	// Check that "Line 2" is not on its own line
	lines := strings.Split(contentStr, "\n")
	found := false
	for _, line := range lines {
		if line == "Line 2" {
			found = true
			break
		}
	}
	assert.False(t, found, "Found 'Line 2' as a separate line in the file")
}

func TestApplyPatchTool_DeleteFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	originalContent := "Line 1\nLine 2\n"
	err := os.WriteFile(testFile, []byte(originalContent), 0644)
	require.NoError(t, err)

	tool := NewApplyPatchTool(tmpDir)

	ctx := context.Background()
	patchText := `*** Begin Patch
*** Delete File: test.txt
*** End Patch`

	params := map[string]any{
		"patchText": patchText,
	}

	result, err := tool.Execute(ctx, params, &framework.ExecutionContext{
		WorkingDir:      tmpDir,
		PermissionAsker: &testAsker{granted: make(map[string]bool)},
	})

	require.NoError(t, err)
	assert.Contains(t, result.Output, "Success")

	// Verify file was deleted
	_, err = os.Stat(testFile)
	assert.True(t, os.IsNotExist(err))
}

func TestParsePatch(t *testing.T) {
	patchText := `*** Begin Patch
*** Add File: test.txt
+Hello
+World
*** End Patch`

	hunks, err := parsePatch(patchText)
	require.NoError(t, err)
	assert.Len(t, hunks, 1)
	assert.Equal(t, "add", hunks[0].hunkType)
	assert.Equal(t, "test.txt", hunks[0].path)
	assert.Equal(t, "Hello\nWorld", hunks[0].contents)
}

func TestApplyPatchTool_EmptyPatch(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewApplyPatchTool(tmpDir)

	ctx := context.Background()
	patchText := `*** Begin Patch
*** End Patch`

	params := map[string]any{
		"patchText": patchText,
	}

	_, err := tool.Execute(ctx, params, &framework.ExecutionContext{
		WorkingDir:      tmpDir,
		PermissionAsker: &testAsker{granted: make(map[string]bool)},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty patch")
}

func TestApplyPatchTool_InvalidPatch(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewApplyPatchTool(tmpDir)

	ctx := context.Background()
	patchText := `This is not a valid patch`

	params := map[string]any{
		"patchText": patchText,
	}

	_, err := tool.Execute(ctx, params, &framework.ExecutionContext{
		WorkingDir:      tmpDir,
		PermissionAsker: &testAsker{granted: make(map[string]bool)},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid patch format")
}
