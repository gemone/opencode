// Package builtin provides built-in tool implementations
package builtin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gemone/libcode/internal/tool/framework"
)

const (
	defaultNumResults = 8
	searchTimeout     = 25 * time.Second
)

// WebSearchTool searches the web using Exa API
type WebSearchTool struct {
	client *http.Client
}

// NewWebSearchTool creates a new web search tool
func NewWebSearchTool() *WebSearchTool {
	return &WebSearchTool{
		client: &http.Client{
			Timeout: searchTimeout,
		},
	}
}

// ID returns the tool identifier
func (t *WebSearchTool) ID() string {
	return "websearch"
}

// Description returns the tool description
func (t *WebSearchTool) Description() string {
	return "Search the web for current information. Use this when you need up-to-date facts or recent events."
}

// Parameters returns the parameter schema
func (t *WebSearchTool) Parameters() *framework.Schema {
	return framework.NewSchema().
		AddProperty("query", framework.Property{
			Type:        "string",
			Description: "Web search query",
			Required:    true,
		}).
		AddProperty("numResults", framework.Property{
			Type:        "integer",
			Description: "Number of search results to return (default: 8)",
			Default:     defaultNumResults,
		}).
		AddProperty("livecrawl", framework.Property{
			Type:        "string",
			Description: "Live crawl mode - 'fallback': use live crawling as backup, 'preferred': prioritize live crawling",
			Default:     "fallback",
			Enum:        []string{"fallback", "preferred"},
		}).
		AddProperty("searchType", framework.Property{
			Type:        "string",
			Description: "Search type - 'auto': balanced, 'fast': quick results, 'deep': comprehensive",
			Default:     "auto",
			Enum:        []string{"auto", "fast", "deep"},
		})
}

// Execute performs web search
func (t *WebSearchTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
	query, _ := params["query"].(string)
	numResults := int(params["numResults"].(float64))
	livecrawl, _ := params["livecrawl"].(string)
	searchType, _ := params["searchType"].(string)

	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	if numResults <= 0 {
		numResults = defaultNumResults
	}

	if livecrawl == "" {
		livecrawl = "fallback"
	}

	if searchType == "" {
		searchType = "auto"
	}

	// Ask for permission
	if execCtx.PermissionAsker != nil {
		err := execCtx.PermissionAsker.Ask(&framework.PermissionRequest{
			Permission: "websearch",
			Patterns:   []string{query},
			Always:     []string{"*"},
			Metadata: map[string]any{
				"query":      query,
				"numResults": numResults,
				"livecrawl":  livecrawl,
				"type":       searchType,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	// Create search request
	searchRequest := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "web_search_exa",
			"arguments": map[string]any{
				"query":     query,
				"type":      searchType,
				"numResults": numResults,
				"livecrawl": livecrawl,
			},
		},
	}

	// Marshal request
	reqBody, err := json.Marshal(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request with context
	reqCtx, cancel := context.WithTimeout(ctx, searchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "POST", "https://mcp.exa.ai/mcp", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	// Execute request
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search error (%d): %s", resp.StatusCode, string(body))
	}

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse SSE response
	lines := strings.Split(string(responseBody), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			var sseResponse struct {
				JSONRPC string `json:"jsonrpc"`
				Result  struct {
					Content []struct {
						Type string `json:"type"`
						Text string `json:"text"`
					} `json:"content"`
				} `json:"result"`
			}

			if err := json.Unmarshal([]byte(data), &sseResponse); err != nil {
				continue
			}

			if len(sseResponse.Result.Content) > 0 && sseResponse.Result.Content[0].Text != "" {
				return &framework.Result{
					Title:  fmt.Sprintf("Web search: %s", query),
					Output: sseResponse.Result.Content[0].Text,
					Metadata: map[string]any{
						"query":      query,
						"numResults": numResults,
						"livecrawl":  livecrawl,
						"type":       searchType,
					},
				}, nil
			}
		}
	}

	return &framework.Result{
		Title:  fmt.Sprintf("Web search: %s", query),
		Output: "No search results found. Please try a different query.",
		Metadata: map[string]any{
			"query": query,
		},
	}, nil
}
