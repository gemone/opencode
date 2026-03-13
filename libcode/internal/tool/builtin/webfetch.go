// Package builtin provides built-in tool implementations
package builtin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gemone/libcode/internal/tool/framework"
)

const (
	maxResponseSize = 5 * 1024 * 1024 // 5MB
	defaultTimeout  = 30 * time.Second
	maxTimeout      = 120 * time.Second
)

// WebFetchTool fetches content from web URLs
type WebFetchTool struct {
	client *http.Client
}

// NewWebFetchTool creates a new web fetch tool
func NewWebFetchTool() *WebFetchTool {
	return &WebFetchTool{
		client: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// ID returns the tool identifier
func (t *WebFetchTool) ID() string {
	return "webfetch"
}

// Description returns the tool description
func (t *WebFetchTool) Description() string {
	return "Fetch content from web URLs. Supports text, markdown, and HTML formats."
}

// Parameters returns the parameter schema
func (t *WebFetchTool) Parameters() *framework.Schema {
	return framework.NewSchema().
		AddProperty("url", framework.Property{
			Type:        "string",
			Description: "The URL to fetch content from",
			Required:    true,
		}).
		AddProperty("format", framework.Property{
			Type:        "string",
			Description: "The format to return content in (text, markdown, or html)",
			Default:     "markdown",
			Enum:        []string{"text", "markdown", "html"},
		}).
		AddProperty("timeout", framework.Property{
			Type:        "integer",
			Description: "Optional timeout in seconds (max 120)",
			Default:     int(defaultTimeout.Seconds()),
		})
}

// Execute fetches the URL
func (t *WebFetchTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
	url, _ := params["url"].(string)
	format, _ := params["format"].(string)

	// Safely extract timeout with default fallback
	var timeoutSec int
	if timeoutVal, ok := params["timeout"]; ok {
		if timeoutFloat, ok := timeoutVal.(float64); ok {
			timeoutSec = int(timeoutFloat)
		}
	}

	// Validate URL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("URL must start with http:// or https://")
	}

	if format == "" {
		format = "markdown"
	}

	// Ask for permission
	if execCtx.PermissionAsker != nil {
		err := execCtx.PermissionAsker.Ask(&framework.PermissionRequest{
			Permission: "webfetch",
			Patterns:   []string{url},
			Always:     []string{"*"},
			Metadata: map[string]any{
				"url":     url,
				"format":  format,
				"timeout": timeoutSec,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	// Set timeout
	timeout := time.Duration(timeoutSec) * time.Second
	if timeout > maxTimeout {
		timeout = maxTimeout
	}
	if timeout == 0 {
		timeout = defaultTimeout
	}

	// Create request with context
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	switch format {
	case "markdown":
		req.Header.Set("Accept", "text/markdown;q=1.0, text/x-markdown;q=0.9, text/plain;q=0.8, text/html;q=0.7, */*;q=0.1")
	case "text":
		req.Header.Set("Accept", "text/plain;q=1.0, text/markdown;q=0.9, text/html;q=0.8, */*;q=0.1")
	case "html":
		req.Header.Set("Accept", "text/html;q=1.0, application/xhtml+xml;q=0.9, text/plain;q=0.8, */*;q=0.1")
	default:
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	}

	// Execute request
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	// Check content length
	if resp.ContentLength > maxResponseSize {
		return nil, fmt.Errorf("response too large (exceeds 5MB limit)")
	}

	// Read response with limit
	limitedReader := io.LimitReader(resp.Body, maxResponseSize+1)
	content, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if len(content) > maxResponseSize {
		return nil, fmt.Errorf("response too large (exceeds 5MB limit)")
	}

	contentType := resp.Header.Get("Content-Type")
	title := fmt.Sprintf("%s (%s)", url, contentType)

	// Check if image
	mime := strings.ToLower(strings.Split(contentType, ";")[0])
	isImage := strings.HasPrefix(mime, "image/") && mime != "image/svg+xml"

	if isImage {
		// Return image as base64
		return &framework.Result{
			Title:  title,
			Output: "Image fetched successfully",
			Metadata: map[string]any{
				"mimeType": mime,
				"isImage":  true,
			},
		}, nil
	}

	// Convert content based on format
	contentStr := string(content)
	switch format {
	case "markdown":
		if strings.Contains(contentType, "text/html") {
			contentStr = htmlToMarkdown(contentStr)
		}
	case "text":
		if strings.Contains(contentType, "text/html") {
			contentStr = htmlToText(contentStr)
		}
	case "html":
		// Already in HTML format
	default:
		// Return as-is
	}

	return &framework.Result{
		Title:  title,
		Output: contentStr,
		Metadata: map[string]any{
			"url":        url,
			"format":     format,
			"contentType": contentType,
			"size":       len(content),
		},
	}, nil
}

// htmlToMarkdown converts HTML to simplified markdown
func htmlToMarkdown(html string) string {
	// Simplified HTML to Markdown conversion
	// For production, use a proper library like github.com/JohannesKaufmann/html-to-markdown
	content := html

	// Remove script and style tags
	content = removeTags(content, "script")
	content = removeTags(content, "style")
	content = removeTags(content, "noscript")

	// Convert headers
	content = strings.ReplaceAll(content, "<h1>", "# ")
	content = strings.ReplaceAll(content, "</h1>", "")
	content = strings.ReplaceAll(content, "<h2>", "## ")
	content = strings.ReplaceAll(content, "</h2>", "")
	content = strings.ReplaceAll(content, "<h3>", "### ")
	content = strings.ReplaceAll(content, "</h3>", "")

	// Convert bold/italic
	content = strings.ReplaceAll(content, "<strong>", "**")
	content = strings.ReplaceAll(content, "</strong>", "**")
	content = strings.ReplaceAll(content, "<b>", "**")
	content = strings.ReplaceAll(content, "</b>", "**")
	content = strings.ReplaceAll(content, "<em>", "*")
	content = strings.ReplaceAll(content, "</em>", "*")
	content = strings.ReplaceAll(content, "<i>", "*")
	content = strings.ReplaceAll(content, "</i>", "*")

	// Convert code blocks
	content = strings.ReplaceAll(content, "<code>", "`")
	content = strings.ReplaceAll(content, "</code>", "`")
	content = strings.ReplaceAll(content, "<pre>", "```\n")
	content = strings.ReplaceAll(content, "</pre>", "\n```")

	// Convert links
	content = strings.ReplaceAll(content, "<a href=\"", "[")
	content = replaceAllBetween(content, "\">", "</a>", "](", ")")

	// Convert line breaks
	content = strings.ReplaceAll(content, "<br>", "\n")
	content = strings.ReplaceAll(content, "<br/>", "\n")
	content = strings.ReplaceAll(content, "<br />", "\n")

	// Remove remaining HTML tags
	content = removeHTMLTags(content)

	// Clean up whitespace
	lines := strings.Split(content, "\n")
	var cleaned []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}

	return strings.Join(cleaned, "\n")
}

// htmlToText converts HTML to plain text
func htmlToText(html string) string {
	content := html

	// Remove script and style tags
	content = removeTags(content, "script")
	content = removeTags(content, "style")
	content = removeTags(content, "noscript")

	// Remove all HTML tags
	content = removeHTMLTags(content)

	// Clean up whitespace
	lines := strings.Split(content, "\n")
	var cleaned []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}

	return strings.Join(cleaned, "\n")
}

// removeTags removes all occurrences of a specific HTML tag
func removeTags(content, tag string) string {
	// Simple tag removal - for production use a proper HTML parser
	openTag := "<" + tag
	closeTag := "</" + tag + ">"

	result := content
	for {
		idx := strings.Index(result, openTag)
		if idx == -1 {
			break
		}

		closeIdx := strings.Index(result[idx:], ">")
		if closeIdx == -1 {
			break
		}

		result = result[:idx] + result[idx+closeIdx+1:]
	}

	result = strings.ReplaceAll(result, closeTag, "")
	return result
}

// removeHTMLTags removes all HTML tags from content
func removeHTMLTags(content string) string {
	result := content
	inTag := false
	var cleaned strings.Builder

	for i := 0; i < len(result); i++ {
		if result[i] == '<' {
			inTag = true
			continue
		}
		if result[i] == '>' && inTag {
			inTag = false
			continue
		}
		if !inTag {
			cleaned.WriteByte(result[i])
		}
	}

	return cleaned.String()
}

// replaceAllBetween replaces text between two markers
func replaceAllBetween(content, start, end, newStart, newEnd string) string {
	result := content
	for {
		startIdx := strings.Index(result, start)
		if startIdx == -1 {
			break
		}
		endIdx := strings.Index(result[startIdx:], end)
		if endIdx == -1 {
			break
		}
		endIdx += startIdx

		before := result[:startIdx]
		after := result[endIdx+len(end):]
		middle := result[startIdx+len(start) : endIdx]

		result = before + newStart + middle + newEnd + after
	}
	return result
}
