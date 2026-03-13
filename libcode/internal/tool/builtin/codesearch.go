// Package builtin provides built-in tool implementations
package builtin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gemone/libcode/internal/tool/framework"
)

const (
	defaultCodeSearchTokens = 5000
	minCodeSearchTokens    = 1000
	maxCodeSearchTokens    = 50000
	codeSearchTimeout       = 30 * time.Second
	codesearchBaseURL       = "https://mcp.exa.ai/mcp"
)

// CodeSearchTool searches for code context using Exa's MCP API
type CodeSearchTool struct {
	client *http.Client
}

// NewCodeSearchTool creates a new codesearch tool
func NewCodeSearchTool() *CodeSearchTool {
	return &CodeSearchTool{
		client: &http.Client{
			Timeout: codeSearchTimeout,
		},
	}
}

// ID returns the tool identifier
func (t *CodeSearchTool) ID() string {
	return "codesearch"
}

// Description returns the tool description
func (t *CodeSearchTool) Description() string {
	return "Search for code context, APIs, libraries, and SDKs documentation. Uses Exa's MCP API to find relevant code examples and documentation."
}

// Parameters returns the parameter schema
func (t *CodeSearchTool) Parameters() *framework.Schema {
	return framework.NewSchema().
		AddProperty("query", framework.Property{
			Type:        "string",
			Description: "Search query to find relevant context for APIs, Libraries, and SDKs. For example, 'React useState hook examples', 'Python pandas dataframe filtering', 'Express.js middleware', 'Next js partial prerendering configuration'",
			Required:    true,
		}).
		AddProperty("tokensNum", framework.Property{
			Type:        "integer",
			Description: "Number of tokens to return (1000-50000). Default is 5000 tokens. Adjust this value based on how much context you need - use lower values for focused queries and higher values for comprehensive documentation.",
			Default:     5000,
			Min:         ptr(minCodeSearchTokens),
			Max:         ptr(maxCodeSearchTokens),
		})
}

// Execute executes the code search
func (t *CodeSearchTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
	query, _ := params["query"].(string)
	tokensNum := int(params["tokensNum"].(float64))

	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	if tokensNum < minCodeSearchTokens {
		tokensNum = defaultCodeSearchTokens
	} else if tokensNum > maxCodeSearchTokens {
		tokensNum = maxCodeSearchTokens
	}

	// Ask for permission
	if execCtx.PermissionAsker != nil {
		err := execCtx.PermissionAsker.Ask(&framework.PermissionRequest{
			Permission: "codesearch",
			Patterns:   []string{query},
			Always:     []string{"*"},
			Metadata: map[string]any{
				"query":     query,
				"tokensNum": tokensNum,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	// Build request
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "get_code_context_exa",
			"arguments": map[string]interface{}{
				"query":     query,
				"tokensNum": tokensNum,
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request with context
	reqCtx, cancel := context.WithTimeout(ctx, codeSearchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "POST", codesearchBaseURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	// Execute request
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("code search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("code search error (%d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse JSON response
	var response struct {
		JSONRPC string `json:"jsonrpc"`
		Result  struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Result.Content) > 0 && response.Result.Content[0].Text != "" {
		return &framework.Result{
			Title:  fmt.Sprintf("Code search: %s", query),
			Output: response.Result.Content[0].Text,
			Metadata: map[string]any{
				"query":     query,
				"tokensNum": tokensNum,
			},
		}, nil
	}

	return &framework.Result{
		Title:  fmt.Sprintf("Code search: %s", query),
		Output: "No code snippets or documentation found. Please try a different query, be more specific about the library or programming concept, or check the spelling of framework names.",
		Metadata: map[string]any{
			"query":     query,
			"tokensNum": tokensNum,
		},
	}, nil
}
