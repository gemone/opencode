// Package builtin provides built-in tool implementations
package builtin

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gemone/libcode/internal/tool/framework"
)

const (
	defaultReadLimit = 2000
	maxLineLength    = 2000
	maxBytes         = 50 * 1024
	maxBytesLabel    = "50 KB"
)

// ReadTool reads files and directories
type ReadTool struct {
	workingDir string
}

// NewReadTool creates a new read tool
func NewReadTool(workingDir string) *ReadTool {
	return &ReadTool{workingDir: workingDir}
}

// ID returns the tool identifier
func (t *ReadTool) ID() string {
	return "read"
}

// Description returns the tool description
func (t *ReadTool) Description() string {
	return "Read the contents of a file or directory. Supports reading images, PDFs, and text files with offset/limit pagination."
}

// Parameters returns the parameter schema
func (t *ReadTool) Parameters() *framework.Schema {
	return framework.NewSchema().
		AddProperty("filePath", framework.Property{
			Type:        "string",
			Description: "The absolute path to the file or directory to read",
			Required:    true,
			Format:      "uri",
		}).
		AddProperty("offset", framework.Property{
			Type:        "integer",
			Description: "The line number to start reading from (1-indexed)",
			Default:     0,
			Min:         ptr(0.0),
		}).
		AddProperty("limit", framework.Property{
			Type:        "integer",
			Description: "The maximum number of lines to read (defaults to 2000)",
			Default:     2000,
			Min:         ptr(1.0),
		})
}

func ptr(v float64) *float64 {
	return &v
}

// Execute reads the file or directory
func (t *ReadTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
	filePath, _ := params["filePath"].(string)
	offset := int(params["offset"].(float64))
	limit := int(params["limit"].(float64))

	if offset < 1 {
		offset = 1
	}
	if limit <= 0 {
		limit = defaultReadLimit
	}

	// Resolve relative path
	filePath = resolvePath(filePath, t.workingDir)

	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	// Ask for permission
	if err := execCtx.AskPermission("read", []string{filePath}, nil, nil); err != nil {
		return nil, err
	}

	title := filepath.Base(filePath)

	// Handle directory
	if info.IsDir() {
		return t.readDirectory(ctx, filePath, offset, limit, title)
	}

	// Handle file
	return t.readFile(ctx, filePath, offset, limit, title, info.Size())
}

// readDirectory reads a directory's contents
func (t *ReadTool) readDirectory(ctx context.Context, filePath string, offset, limit int, title string) (*framework.Result, error) {
	entries, err := os.ReadDir(filePath)
	if err != nil {
		return nil, err
	}

	// Sort entries
	var entryNames []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		} else if entry.Type()&os.ModeSymlink != 0 {
			// Check if symlink points to directory
			if info, err := os.Stat(filepath.Join(filePath, entry.Name())); err == nil && info.IsDir() {
				name += "/"
			}
		}
		entryNames = append(entryNames, name)
	}

	// Simple string sort
	for i := 0; i < len(entryNames); i++ {
		for j := i + 1; j < len(entryNames); j++ {
			if entryNames[i] > entryNames[j] {
				entryNames[i], entryNames[j] = entryNames[j], entryNames[i]
			}
		}
	}

	start := offset - 1
	end := start + limit
	if end > len(entryNames) {
		end = len(entryNames)
	}
	sliced := entryNames[start:end]
	truncated := end < len(entryNames)

	var output strings.Builder
	output.WriteString(fmt.Sprintf("<path>%s</path>\n", filePath))
	output.WriteString("<type>directory</type>\n")
	output.WriteString("<entries>\n")
	output.WriteString(strings.Join(sliced, "\n"))
	if truncated {
		output.WriteString(fmt.Sprintf("\n(Showing %d of %d entries. Use 'offset' parameter to read beyond entry %d)", len(sliced), len(entryNames), offset+len(sliced)))
	} else {
		output.WriteString(fmt.Sprintf("\n(%d entries)", len(entryNames)))
	}
	output.WriteString("\n</entries>")

	preview := strings.Join(sliced, "\n")
	if len(preview) > 1000 {
		preview = preview[:1000]
	}

	return &framework.Result{
		Title:  title,
		Output: output.String(),
		Metadata: map[string]any{
			"preview":   preview,
			"truncated": truncated,
			"loaded":    []string{},
		},
	}, nil
}

// readFile reads a text file
func (t *ReadTool) readFile(ctx context.Context, filePath string, offset, limit int, title string, fileSize int64) (*framework.Result, error) {
	// Check for binary file
	if isBinary, _ := isBinaryFile(filePath, fileSize); isBinary {
		return nil, fmt.Errorf("cannot read binary file: %s", filePath)
	}

	// Check for image or PDF
	mimeType := getMimeType(filePath)
	if strings.HasPrefix(mimeType, "image/") && mimeType != "image/svg+xml" {
		return t.readImageFile(filePath, title, mimeType)
	}
	if mimeType == "application/pdf" {
		return t.readPDFFile(filePath, title, mimeType)
	}

	// Read text file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	var bytes int
	lineNum := 0
	truncatedByBytes := false
	hasMoreLines := false

	scanner := func(r io.Reader) *bufio.Scanner {
		s := bufio.NewScanner(r)
		s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		return s
	}(file)

	for scanner.Scan() {
		lineNum++
		if lineNum < offset {
			continue
		}

		if len(lines) >= limit {
			hasMoreLines = true
			continue
		}

		line := scanner.Text()
		if len(line) > maxLineLength {
			line = line[:maxLineLength] + "... (line truncated)"
		}

		lineBytes := len(line) + 1 // +1 for newline
		if bytes+lineBytes > maxBytes {
			truncatedByBytes = true
			hasMoreLines = true
			break
		}

		lines = append(lines, line)
		bytes += lineBytes
	}

	if lineNum > 0 && offset > lineNum {
		return nil, fmt.Errorf("offset %d is out of range for this file (%d lines)", offset, lineNum)
	}

	// Format output
	var output strings.Builder
	output.WriteString(fmt.Sprintf("<path>%s</path>\n", filePath))
	output.WriteString("<type>file</type>\n")
	output.WriteString("<content>\n")

	for i, line := range lines {
		output.WriteString(fmt.Sprintf("%d: %s\n", offset+i, line))
	}

	totalLines := lineNum
	lastReadLine := offset + len(lines) - 1
	nextOffset := lastReadLine + 1
	truncated := hasMoreLines || truncatedByBytes

	if truncatedByBytes {
		output.WriteString(fmt.Sprintf("\n(Output capped at %s. Showing lines %d-%d. Use offset=%d to continue.)\n", maxBytesLabel, offset, lastReadLine, nextOffset))
	} else if hasMoreLines {
		output.WriteString(fmt.Sprintf("\n(Showing lines %d-%d of %d. Use offset=%d to continue.)\n", offset, lastReadLine, totalLines, nextOffset))
	} else {
		output.WriteString(fmt.Sprintf("\n(End of file - total %d lines)\n", totalLines))
	}

	output.WriteString("</content>")

	// Build preview
	preview := strings.Join(lines, "\n")
	if len(preview) > 1000 {
		preview = preview[:1000]
	}

	return &framework.Result{
		Title:  title,
		Output: output.String(),
		Metadata: map[string]any{
			"preview":   preview,
			"truncated": truncated,
			"loaded":    []string{},
		},
	}, nil
}

// readImageFile reads an image file as base64
func (t *ReadTool) readImageFile(filePath, title, mimeType string) (*framework.Result, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	base64Data := base64.StdEncoding.EncodeToString(data)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)

	msg := "Image read successfully"
	return &framework.Result{
		Title:  title,
		Output: msg,
		Metadata: map[string]any{
			"preview":   msg,
			"truncated": false,
			"loaded":    []string{},
		},
		Attachments: []framework.Attachment{
			{
				Type: "file",
				MIME: mimeType,
				URL:  dataURL,
			},
		},
	}, nil
}

// readPDFFile reads a PDF file as base64
func (t *ReadTool) readPDFFile(filePath, title, mimeType string) (*framework.Result, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	base64Data := base64.StdEncoding.EncodeToString(data)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)

	msg := "PDF read successfully"
	return &framework.Result{
		Title:  title,
		Output: msg,
		Metadata: map[string]any{
			"preview":   msg,
			"truncated": false,
			"loaded":    []string{},
		},
		Attachments: []framework.Attachment{
			{
				Type: "file",
				MIME: mimeType,
				URL:  dataURL,
			},
		},
	}, nil
}

// isBinaryFile checks if a file is binary
func isBinaryFile(filePath string, fileSize int64) (bool, error) {
	if fileSize == 0 {
		return false, nil
	}

	// Check extension
	ext := strings.ToLower(filepath.Ext(filePath))
	binaryExts := map[string]bool{
		".zip": true, ".tar": true, ".gz": true, ".exe": true,
		".dll": true, ".so": true, ".class": true, ".jar": true,
		".war": true, ".7z": true, ".doc": true, ".docx": true,
		".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true,
		".odt": true, ".ods": true, ".odp": true, ".bin": true,
		".dat": true, ".obj": true, ".o": true, ".a": true,
		".lib": true, ".wasm": true, ".pyc": true, ".pyo": true,
	}
	if binaryExts[ext] {
		return true, nil
	}

	// Read first 4KB to check for binary content
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	buffer := make([]byte, 4096)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false, err
	}

	nonPrintableCount := 0
	for i := 0; i < n; i++ {
		if buffer[i] == 0 {
			return true, nil
		}
		if buffer[i] < 9 || (buffer[i] > 13 && buffer[i] < 32) {
			nonPrintableCount++
		}
	}

	// If >30% non-printable, consider binary
	return float64(nonPrintableCount)/float64(n) > 0.3, nil
}

// getMimeType returns the MIME type of a file
func getMimeType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".pdf":  "application/pdf",
		".txt":  "text/plain",
		".md":   "text/markdown",
		".html": "text/html",
		".css":  "text/css",
		".js":   "text/javascript",
		".json": "application/json",
		".xml":  "application/xml",
	}
	if mt, ok := mimeTypes[ext]; ok {
		return mt
	}
	return "application/octet-stream"
}
