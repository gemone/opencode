// Package openai provides OpenAI provider implementation
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gemone/libcode/internal/ai"
	"github.com/gemone/libcode/internal/ai/streaming"
	"github.com/gemone/libcode/internal/ai/toolcall"
)

const (
	defaultBaseURL = "https://api.openai.com/v1"
	defaultModel   = "gpt-4"
)

// Provider implements the AI provider interface for OpenAI
type Provider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// Config holds the provider configuration
type Config struct {
	APIKey  string
	BaseURL string
}

// NewProvider creates a new OpenAI provider
func NewProvider(cfg Config) *Provider {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	return &Provider{
		apiKey:  cfg.APIKey,
		baseURL: baseURL,
		client: &http.Client{},
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "openai"
}

// SupportsToolCalling returns true if OpenAI supports tool calling
func (p *Provider) SupportsToolCalling() bool {
	return true
}

// SupportsStreaming returns true if OpenAI supports streaming
func (p *Provider) SupportsStreaming() bool {
	return true
}

// Stream sends a request and returns a streaming response
func (p *Provider) Stream(ctx context.Context, req *ai.Request) (<-chan *ai.Chunk, error) {
	chunkCh := make(chan *ai.Chunk, 10)

	go func() {
		defer close(chunkCh)

		body, err := p.buildRequestBody(req)
		if err != nil {
			chunkCh <- &ai.Chunk{Error: err}
			return
		}

		httpReq, err := p.createRequest(ctx, body, req)
		if err != nil {
			chunkCh <- &ai.Chunk{Error: err}
			return
		}

		resp, err := p.client.Do(httpReq)
		if err != nil {
			chunkCh <- &ai.Chunk{Error: err}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			chunkCh <- &ai.Chunk{Error: fmt.Errorf("API error: %s", resp.Status)}
			return
		}

		p.parseStream(ctx, resp.Body, chunkCh, req)
	}()

	return chunkCh, nil
}

// Complete sends a request and returns a complete response
func (p *Provider) Complete(ctx context.Context, req *ai.Request) (*ai.Response, error) {
	// For simplicity, use streaming and aggregate
	chunkCh, err := p.Stream(ctx, req)
	if err != nil {
		return nil, err
	}

	var buffer streaming.ChunkBuffer
	var usage *ai.Usage
	var finishReason string

	for chunk := range chunkCh {
		if chunk.Error != nil {
			return nil, chunk.Error
		}
		buffer.Add(chunk)
		if chunk.Usage != nil {
			usage = chunk.Usage
		}
		if chunk.FinishReason != nil {
			finishReason = *chunk.FinishReason
		}
	}

	response := &ai.Response{
		Message: ai.Message{
			Role:    "assistant",
			Content: buffer.Content(),
		},
		ToolCalls:   convertToolCalls(buffer.ToolCalls()),
		Model:       req.Model,
		FinishReason: finishReason,
	}

	if usage != nil {
		response.Usage = *usage
	}

	return response, nil
}

// buildRequestBody creates the request body for the API
func (p *Provider) buildRequestBody(req *ai.Request) ([]byte, error) {
	body := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   true,
	}

	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}
	if req.TopP > 0 {
		body["top_p"] = req.TopP
	}
	if req.MaxTokens > 0 {
		body["max_tokens"] = req.MaxTokens
	}

	// Add tools if provided
	if len(req.Tools) > 0 {
		body["tools"] = req.Tools
		if req.ToolChoice != "" {
			body["tool_choice"] = req.ToolChoice
		}
	}

	return json.Marshal(body)
}

// createRequest creates the HTTP request
func (p *Provider) createRequest(ctx context.Context, body []byte, req *ai.Request) (*http.Request, error) {
	url := fmt.Sprintf("%s/chat/completions", p.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	return httpReq, nil
}

// parseStream parses the streaming response
func (p *Provider) parseStream(ctx context.Context, body io.Reader, chunkCh chan<- *ai.Chunk, req *ai.Request) {
	parser := streaming.NewSSEParser(body)
	eventCh := parser.Parse(ctx)

	for event := range eventCh {
		if event.Error != nil {
			chunkCh <- &ai.Chunk{Error: event.Error}
			return
		}

		if event.Event != "message.delta" {
			// Handle other events (error, done, etc.)
			if event.Data == "[DONE]" {
				chunkCh <- &ai.Chunk{
					FinishReason: ptr("stop"),
				}
				return
			}
			continue
		}

		// Parse delta data
		var delta struct {
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
			FinishReason string `json:"finish_reason"`
			Usage       *ai.Usage `json:"usage"`
		}

		if err := json.Unmarshal([]byte(event.Data), &delta); err != nil {
			continue
		}

		chunk := &ai.Chunk{
			Delta: delta.Delta.Content,
		}

		if delta.FinishReason != "" {
			chunk.FinishReason = &delta.FinishReason
		}
		if delta.Usage != nil {
			chunk.Usage = delta.Usage
		}

		chunkCh <- chunk
	}
}

// ptr returns a pointer to the given value
func ptr[T any](v T) *T {
	return &v
}

// convertToolCalls converts tool calls to the expected format
func convertToolCalls(calls []ai.ToolCall) []toolcall.Call {
	result := make([]toolcall.Call, len(calls))
	for i, call := range calls {
		result[i] = toolcall.Call{
			ID:    call.ID,
			Tool:  call.Tool,
			Input: call.Input,
		}
	}
	return result
}

// getAPIKey retrieves the API key from environment or config
func getAPIKey() string {
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		return key
	}
	if key := os.Getenv("LIBCODE_OPENAI_API_KEY"); key != "" {
		return key
	}
	return ""
}

// NewProviderFromEnv creates a provider from environment variables
func NewProviderFromEnv() *Provider {
	cfg := Config{
		APIKey: getAPIKey(),
	}
	if cfg.APIKey == "" {
		return nil // No API key available
	}
	return NewProvider(cfg)
}
