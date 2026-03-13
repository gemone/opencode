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

// ApplyPatchTool applies unified diff patches to files
type ApplyPatchTool struct {
	workingDir string
}

// NewApplyPatchTool creates a new apply_patch tool
func NewApplyPatchTool(workingDir string) *ApplyPatchTool {
	return &ApplyPatchTool{workingDir: workingDir}
}

// ID returns the tool identifier
func (t *ApplyPatchTool) ID() string {
	return "apply_patch"
}

// Description returns the tool description
func (t *ApplyPatchTool) Description() string {
	return "Apply unified diff patches. Supports add, update, delete, and move operations with atomic application."
}

// Parameters returns the parameter schema
func (t *ApplyPatchTool) Parameters() *framework.Schema {
	return framework.NewSchema().
		AddProperty("patchText", framework.Property{
			Type:        "string",
			Description: "The full patch text that describes all changes to be made",
			Required:    true,
		})
}

// Execute applies the patch
func (t *ApplyPatchTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
	patchText, _ := params["patchText"].(string)

	if patchText == "" {
		return nil, fmt.Errorf("patchText is required")
	}

	// Parse the patch
	hunks, err := parsePatch(patchText)
	if err != nil {
		return nil, fmt.Errorf("apply_patch verification failed: %w", err)
	}

	if len(hunks) == 0 {
		normalized := strings.ReplaceAll(strings.ReplaceAll(patchText, "\r\n", "\n"), "\r", "\n")
		normalized = strings.TrimSpace(normalized)
		if normalized == "*** Begin Patch\n*** End Patch" {
			return nil, fmt.Errorf("patch rejected: empty patch")
		}
		return nil, fmt.Errorf("apply_patch verification failed: no hunks found")
	}

	// Collect file changes for validation and permission checking
	fileChanges, err := t.collectFileChanges(hunks)
	if err != nil {
		return nil, err
	}

	// Build relative paths for permission checking
	relativePaths := make([]string, 0, len(fileChanges))
	for _, change := range fileChanges {
		relPath, err := filepath.Rel(t.workingDir, change.filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to get relative path: %w", err)
		}
		relativePaths = append(relativePaths, relPath)
	}

	// Build total diff for permission metadata
	totalDiff := strings.Builder{}
	for _, change := range fileChanges {
		totalDiff.WriteString(change.diff)
		totalDiff.WriteString("\n")
	}

	// Build files metadata
	filesMetadata := make([]map[string]any, 0, len(fileChanges))
	for _, change := range fileChanges {
		metadata := map[string]any{
			"filePath":   change.filePath,
			"type":       change.changeType,
			"diff":       change.diff,
			"before":     change.oldContent,
			"after":      change.newContent,
			"additions":  change.additions,
			"deletions":  change.deletions,
		}
		if change.movePath != "" {
			metadata["movePath"] = change.movePath
		}
		filesMetadata = append(filesMetadata, metadata)
	}

	// Ask for permission
	if execCtx.PermissionAsker != nil {
		err := execCtx.PermissionAsker.Ask(&framework.PermissionRequest{
			Permission: "edit",
			Patterns:   relativePaths,
			Always:     []string{"*"},
			Metadata: map[string]any{
				"filepath": strings.Join(relativePaths, ", "),
				"diff":     totalDiff.String(),
				"files":    filesMetadata,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	// Apply changes atomically
	updates, err := t.applyChanges(fileChanges)
	if err != nil {
		return nil, fmt.Errorf("failed to apply changes: %w", err)
	}

	// Generate output summary
	summaryLines := make([]string, 0, len(fileChanges))
	for _, change := range fileChanges {
		relPath, err := filepath.Rel(t.workingDir, change.filePath)
		if err != nil {
			continue
		}
		relPath = filepath.ToSlash(relPath)

		switch change.changeType {
		case "add":
			summaryLines = append(summaryLines, fmt.Sprintf("A %s", relPath))
		case "delete":
			summaryLines = append(summaryLines, fmt.Sprintf("D %s", relPath))
		case "move":
			moveRelPath, err := filepath.Rel(t.workingDir, change.movePath)
			if err != nil {
				continue
			}
			summaryLines = append(summaryLines, fmt.Sprintf("M %s -> %s", relPath, filepath.ToSlash(moveRelPath)))
		default: // update
			summaryLines = append(summaryLines, fmt.Sprintf("M %s", relPath))
		}
	}

	output := fmt.Sprintf("Success. Updated the following files:\n%s", strings.Join(summaryLines, "\n"))

	return &framework.Result{
		Title:    "Patch Applied",
		Output:   output,
		Metadata: map[string]any{
			"diff":       totalDiff.String(),
			"files":      filesMetadata,
			"updates":    updates,
			"fileCount":  len(fileChanges),
		},
	}, nil
}

// fileChange represents a single file change
type fileChange struct {
	filePath    string
	movePath    string
	oldContent  string
	newContent  string
	changeType  string // "add", "update", "delete", "move"
	diff        string
	additions   int
	deletions   int
}

// collectFileChanges collects and validates all file changes from hunks
func (t *ApplyPatchTool) collectFileChanges(hunks []hunk) ([]fileChange, error) {
	changes := make([]fileChange, 0, len(hunks))

	for _, hunk := range hunks {
		filePath := filepath.Join(t.workingDir, hunk.path)

		switch hunk.hunkType {
		case "add":
			content := hunk.contents
			if !strings.HasSuffix(content, "\n") {
				content += "\n"
			}

			diff := generateUnifiedDiff(filePath, "", content)
			additions, deletions := countLineChanges("", content)

			changes = append(changes, fileChange{
				filePath:   filePath,
				oldContent: "",
				newContent: content,
				changeType: "add",
				diff:       diff,
				additions:  additions,
				deletions:  deletions,
			})

		case "update", "move":
			// Check if file exists
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				return nil, fmt.Errorf("apply_patch verification failed: failed to read file to update: %s", filePath)
			}

			// Read current content
			oldContent, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read file: %w", err)
			}

			oldContentStr := string(oldContent)

			// Apply chunks to get new content
			newContent, err := applyChunks(oldContentStr, hunk.chunks)
			if err != nil {
				return nil, fmt.Errorf("apply_patch verification failed: %w", err)
			}

			diff := generateUnifiedDiff(filePath, oldContentStr, newContent)
			additions, deletions := countLineChanges(oldContentStr, newContent)

			changeType := "update"
			movePath := ""
			if hunk.movePath != "" {
				changeType = "move"
				movePath = filepath.Join(t.workingDir, hunk.movePath)
			}

			changes = append(changes, fileChange{
				filePath:   filePath,
				movePath:   movePath,
				oldContent: oldContentStr,
				newContent: newContent,
				changeType: changeType,
				diff:       diff,
				additions:  additions,
				deletions:  deletions,
			})

		case "delete":
			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("apply_patch verification failed: failed to read file to delete: %w", err)
			}

			oldContentStr := string(content)
			diff := generateUnifiedDiff(filePath, oldContentStr, "")
			deletions := strings.Count(oldContentStr, "\n") + 1

			changes = append(changes, fileChange{
				filePath:   filePath,
				oldContent: oldContentStr,
				newContent: "",
				changeType: "delete",
				diff:       diff,
				additions:  0,
				deletions:  deletions,
			})
		}
	}

	return changes, nil
}

// applyChanges applies all file changes atomically
func (t *ApplyPatchTool) applyChanges(changes []fileChange) ([]map[string]any, error) {
	updates := make([]map[string]any, 0, len(changes)*2)

	for _, change := range changes {
		switch change.changeType {
		case "add":
			// Create parent directories
			dir := filepath.Dir(change.filePath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create directory: %w", err)
			}

			if err := os.WriteFile(change.filePath, []byte(change.newContent), 0644); err != nil {
				return nil, fmt.Errorf("failed to write file: %w", err)
			}

			updates = append(updates, map[string]any{
				"file":  change.filePath,
				"event": "add",
			})

		case "update":
			if err := os.WriteFile(change.filePath, []byte(change.newContent), 0644); err != nil {
				return nil, fmt.Errorf("failed to write file: %w", err)
			}

			updates = append(updates, map[string]any{
				"file":  change.filePath,
				"event": "change",
			})

		case "move":
			// Create parent directories for move target
			moveDir := filepath.Dir(change.movePath)
			if err := os.MkdirAll(moveDir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create directory: %w", err)
			}

			// Write to new location
			if err := os.WriteFile(change.movePath, []byte(change.newContent), 0644); err != nil {
				return nil, fmt.Errorf("failed to write file: %w", err)
			}

			// Remove old file
			if err := os.Remove(change.filePath); err != nil {
				return nil, fmt.Errorf("failed to remove old file: %w", err)
			}

			updates = append(updates, map[string]any{
				"file":  change.filePath,
				"event": "unlink",
			})
			updates = append(updates, map[string]any{
				"file":  change.movePath,
				"event": "add",
			})

		case "delete":
			if err := os.Remove(change.filePath); err != nil {
				return nil, fmt.Errorf("failed to delete file: %w", err)
			}

			updates = append(updates, map[string]any{
				"file":  change.filePath,
				"event": "unlink",
			})
		}
	}

	return updates, nil
}

// hunk represents a patch hunk
type hunk struct {
	hunkType  string // "add", "update", "delete", "move"
	path      string
	movePath  string
	contents  string
	chunks    []chunk
}

// chunk represents a change chunk
type chunk struct {
	oldLines      []string
	newLines      []string
	changeContext string
	isEndOfFile   bool
}

// parsePatch parses patch text into hunks
func parsePatch(patchText string) ([]hunk, error) {
	// Strip heredoc markers if present
	patchText = stripHeredoc(strings.TrimSpace(patchText))

	lines := strings.Split(patchText, "\n")
	hunks := []hunk{}

	// Find begin/end markers
	beginIdx := -1
	endIdx := -1

	for i, line := range lines {
		if strings.TrimSpace(line) == "*** Begin Patch" {
			beginIdx = i
		}
		if strings.TrimSpace(line) == "*** End Patch" {
			endIdx = i
		}
	}

	if beginIdx == -1 || endIdx == -1 || beginIdx >= endIdx {
		return nil, fmt.Errorf("invalid patch format: missing Begin/End markers")
	}

	// Parse hunks between markers
	i := beginIdx + 1
	for i < endIdx {
		line := strings.TrimSpace(lines[i])

		if strings.HasPrefix(line, "*** Add File:") {
			path := strings.TrimPrefix(line, "*** Add File:")
			path = strings.TrimSpace(path)

			// Parse content
			var content strings.Builder
			i++
			for i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "***") {
				if strings.HasPrefix(lines[i], "+") {
					content.WriteString(strings.TrimPrefix(lines[i], "+"))
					content.WriteString("\n")
				}
				i++
			}

			// Remove trailing newline
			contentStr := content.String()
			if strings.HasSuffix(contentStr, "\n") {
				contentStr = contentStr[:len(contentStr)-1]
			}

			hunks = append(hunks, hunk{
				hunkType: "add",
				path:     path,
				contents: contentStr,
			})

		} else if strings.HasPrefix(line, "*** Delete File:") {
			path := strings.TrimPrefix(line, "*** Delete File:")
			path = strings.TrimSpace(path)

			hunks = append(hunks, hunk{
				hunkType: "delete",
				path:     path,
			})
			i++

		} else if strings.HasPrefix(line, "*** Update File:") {
			path := strings.TrimPrefix(line, "*** Update File:")
			path = strings.TrimSpace(path)

			var movePath string
			i++

			// Check for move directive
			if i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), "*** Move to:") {
				moveLine := strings.TrimSpace(lines[i])
				movePath = strings.TrimPrefix(moveLine, "*** Move to:")
				movePath = strings.TrimSpace(movePath)
				i++
			}

			// Parse chunks
			chunks := []chunk{}
			for i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "***") {
				line := lines[i]
				trimmed := strings.TrimSpace(line)

				// Start of chunk - either @@ marker or line with prefix
				if strings.HasPrefix(trimmed, "@@") || strings.HasPrefix(line, " ") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, "+") {
					oldLines := []string{}
					newLines := []string{}
					isEndOfFile := false

					// Skip the @@ marker if present
					if strings.HasPrefix(trimmed, "@@") {
						i++
					}

					// Parse change lines
					for i < len(lines) {
						changeLine := lines[i]
						changeTrimmed := strings.TrimSpace(changeLine)

						// Check for end markers
						if strings.HasPrefix(changeTrimmed, "@@") || strings.HasPrefix(changeTrimmed, "***") {
							break
						}
						if changeTrimmed == "*** End of File" {
							isEndOfFile = true
							i++
							break
						}

						// Skip empty lines between chunks
						if changeTrimmed == "" {
							i++
							continue
						}

						if strings.HasPrefix(changeLine, " ") {
							// Context line - appears in both
							content := strings.TrimPrefix(changeLine, " ")
							oldLines = append(oldLines, content)
							newLines = append(newLines, content)
						} else if strings.HasPrefix(changeLine, "-") {
							// Remove line - only in old
							oldLines = append(oldLines, strings.TrimPrefix(changeLine, "-"))
						} else if strings.HasPrefix(changeLine, "+") {
							// Add line - only in new
							newLines = append(newLines, strings.TrimPrefix(changeLine, "+"))
						}

						i++
					}

					if len(oldLines) > 0 || len(newLines) > 0 {
						chunks = append(chunks, chunk{
							oldLines:    oldLines,
							newLines:    newLines,
							isEndOfFile: isEndOfFile,
						})
					}
				} else {
					i++
				}
			}

			hunks = append(hunks, hunk{
				hunkType: "update",
				path:     path,
				movePath: movePath,
				chunks:   chunks,
			})

		} else {
			i++
		}
	}

	return hunks, nil
}

// stripHeredoc removes heredoc wrapper markers
func stripHeredoc(input string) string {
	// Match patterns like: cat <<'EOF'\n...\nEOF or <<EOF\n...\nEOF
	lines := strings.Split(input, "\n")
	if len(lines) < 3 {
		return input
	}

	firstLine := strings.TrimSpace(lines[0])
	if strings.HasPrefix(firstLine, "cat <<") || strings.HasPrefix(firstLine, "<<") {
		// Extract delimiter
		parts := strings.Fields(firstLine)
		if len(parts) >= 2 {
			delimiter := strings.Trim(strings.TrimPrefix(parts[len(parts)-1], "<<"), "'\"")
			// Find closing delimiter
			for i := len(lines) - 1; i >= 0; i-- {
				if strings.TrimSpace(lines[i]) == delimiter {
					// Return content between markers
					return strings.Trim(strings.Join(lines[1:i], "\n"), "\n")
				}
			}
		}
	}

	return input
}

// applyChunks applies update chunks to file content
func applyChunks(content string, chunks []chunk) (string, error) {
	lines := strings.Split(content, "\n")

	// Remove trailing empty line if present
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	// Compute replacements
	replacements, err := computeReplacements(lines, chunks)
	if err != nil {
		return "", err
	}

	// Apply replacements in reverse order
	result := make([]string, len(lines))
	copy(result, lines)

	for i := len(replacements) - 1; i >= 0; i-- {
		repl := replacements[i]
		startIdx := repl.startIdx
		oldLen := repl.oldLen
		newSegment := repl.newLines

		// Remove old lines
		result = append(result[:startIdx], result[startIdx+oldLen:]...)

		// Insert new lines at the startIdx position
		if len(newSegment) > 0 {
			before := result[:startIdx]
			after := result[startIdx:]
			result = append(before, append(newSegment, after...)...)
		}
	}

	// Ensure trailing newline
	if len(result) == 0 || result[len(result)-1] != "" {
		result = append(result, "")
	}

	return strings.Join(result, "\n"), nil
}

// replacement represents a text replacement
type replacement struct {
	startIdx int
	oldLen   int
	newLines []string
}

// computeReplacements computes replacements from chunks
func computeReplacements(lines []string, chunks []chunk) ([]replacement, error) {
	replacements := []replacement{}
	lineIndex := 0

	for _, chunk := range chunks {
		// Handle pure addition (no old lines)
		if len(chunk.oldLines) == 0 {
			insertionIdx := len(lines)
			if len(lines) > 0 && lines[len(lines)-1] == "" {
				insertionIdx = len(lines) - 1
			}

			replacements = append(replacements, replacement{
				startIdx: insertionIdx,
				oldLen:   0,
				newLines: chunk.newLines,
			})
			continue
		}

		// Try to match old lines
		pattern := chunk.oldLines
		newSlice := chunk.newLines
		found := seekSequence(lines, pattern, lineIndex, chunk.isEndOfFile)

		// Retry without trailing empty line
		if found == -1 && len(pattern) > 0 && pattern[len(pattern)-1] == "" {
			pattern = pattern[:len(pattern)-1]
			if len(newSlice) > 0 && newSlice[len(newSlice)-1] == "" {
				newSlice = newSlice[:len(newSlice)-1]
			}
			found = seekSequence(lines, pattern, lineIndex, chunk.isEndOfFile)
		}

		if found != -1 {
			replacements = append(replacements, replacement{
				startIdx: found,
				oldLen:   len(pattern),
				newLines: newSlice,
			})
			lineIndex = found + len(pattern)
		} else {
			return nil, fmt.Errorf("failed to find expected lines:\n%s", strings.Join(chunk.oldLines, "\n"))
		}
	}

	// Sort by index
	sort.Slice(replacements, func(i, j int) bool {
		return replacements[i].startIdx < replacements[j].startIdx
	})

	return replacements, nil
}

// seekSequence finds a sequence of lines in the file
func seekSequence(lines, pattern []string, startIndex int, eof bool) int {
	if len(pattern) == 0 {
		return -1
	}

	// EOF anchor - try from end first
	if eof {
		fromEnd := len(lines) - len(pattern)
		if fromEnd >= startIndex {
			matches := true
			for j := 0; j < len(pattern); j++ {
				if lines[fromEnd+j] != pattern[j] {
					matches = false
					break
				}
			}
			if matches {
				return fromEnd
			}
		}
	}

	// Forward search
	for i := startIndex; i <= len(lines)-len(pattern); i++ {
		matches := true
		for j := 0; j < len(pattern); j++ {
			if lines[i+j] != pattern[j] {
				matches = false
				break
			}
		}
		if matches {
			return i
		}
	}

	return -1
}

// generateUnifiedDiff creates a unified diff string
func generateUnifiedDiff(filePath, oldContent, newContent string) string {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	var diff strings.Builder

	diff.WriteString(fmt.Sprintf("--- a/%s\n", filepath.Base(filePath)))
	diff.WriteString(fmt.Sprintf("+++ b/%s\n", filepath.Base(filePath)))
	diff.WriteString("@@ -1 +1 @@\n")

	// Simple diff - mark all lines as changed
	maxLen := len(oldLines)
	if len(newLines) > maxLen {
		maxLen = len(newLines)
	}

	for i := 0; i < maxLen; i++ {
		oldLine := ""
		newLine := ""

		if i < len(oldLines) {
			oldLine = oldLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}

		if oldLine != newLine {
			if oldLine != "" {
				diff.WriteString("-" + oldLine + "\n")
			}
			if newLine != "" {
				diff.WriteString("+" + newLine + "\n")
			}
		} else if oldLine != "" {
			diff.WriteString(" " + oldLine + "\n")
		}
	}

	return diff.String()
}

// countLineChanges counts additions and deletions
func countLineChanges(oldContent, newContent string) (additions, deletions int) {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	// Simple counting - non-zero in old is deletion, non-zero in new is addition
	for _, line := range oldLines {
		if line != "" {
			deletions++
		}
	}

	for _, line := range newLines {
		if line != "" {
			additions++
		}
	}

	return additions, deletions
}
