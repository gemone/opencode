// Package ai provides the AI abstraction layer for libcode
package ai

import (
	"context"
	"encoding/json"
	"io"

	"github.com/gemone/libcode/internal/ai/toolcall"
)

// Provider defines the interface for AI providers
type Provider interface {
	// Stream sends a request and returns a streaming response
	Stream(ctx context.Context, req *Request) (<-chan *Chunk, error)

	// Complete sends a request and returns a complete response
	Complete(ctx context.Context, req *Request) (*Response, error)

	// SupportsToolCalling returns true if the provider supports tool calling
	SupportsToolCalling() bool

	// SupportsStreaming returns true if the provider supports streaming
	SupportsStreaming() bool

	// Name returns the provider name
	Name() string
}

// Request represents an AI request
type Request struct {
	// Model specifies the model to use
	Model string `json:"model"`

	// Messages contains the conversation history
	Messages []Message `json:"messages"`

	// Tools available for the model to call
	Tools []toolcall.Tool `json:"tools,omitempty"`

	// ToolChoice controls how tools are used
	ToolChoice ToolChoice `json:"tool_choice,omitempty"`

	// MaxTokens limits the response length
	MaxTokens int `json:"max_tokens,omitempty"`

	// Temperature controls randomness (0.0 - 2.0)
	Temperature float64 `json:"temperature,omitempty"`

	// TopP controls nucleus sampling
	TopP float64 `json:"top_p,omitempty"`

	// ReasoningEffort controls thinking budget (for Claude)
	ReasoningEffort string `json:"reasoning_effort,omitempty"`

	// Metadata for provider-specific options
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Message represents a message in the conversation
type Message struct {
	Role    string     `json:"role"`    // "user", "assistant", "system"
	Content string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall represents a tool call in a message
type ToolCall struct {
	ID       string         `json:"id"`
	Tool     string         `json:"tool"`
	Input    json.RawMessage `json:"input"`
}

// Response represents a complete AI response
type Response struct {
	// Message is the assistant's response message
	Message Message `json:"message"`

	// ToolCalls contains any tools the model wants to call
	ToolCalls []toolcall.Call `json:"tool_calls,omitempty"`

	// Usage contains token usage information
	Usage Usage `json:"usage"`

	// Model used for the response
	Model string `json:"model"`

	// FinishReason indicates why the response ended
	FinishReason string `json:"finish_reason"`
}

// Chunk represents a streaming response chunk
type Chunk struct {
	// Delta contains the incremental content
	Delta string `json:"delta"`

	// ToolCalls contains incremental tool call information
	ToolCalls []toolcall.CallChunk `json:"tool_calls,omitempty"`

	// Usage contains token usage (final chunk only)
	Usage *Usage `json:"usage,omitempty"`

	// FinishReason indicates why the response ended (final chunk only)
	FinishReason *string `json:"finish_reason,omitempty"`

	// Error contains any error that occurred
	Error error `json:"error,omitempty"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ToolChoice controls how tools are used
type ToolChoice string

const (
	ToolChoiceAuto   ToolChoice = "auto"   // Let the model decide
	ToolChoiceNone   ToolChoice = "none"   // Don't use tools
	ToolChoiceRequired ToolChoice = "required" // Must use a tool
)

// Reader converts a chunk channel to an io.Reader
type Reader struct {
	chunks <-chan *Chunk
	buffer string
}

// NewReader creates a new Reader from a chunk channel
func NewReader(chunks <-chan *Chunk) *Reader {
	return &Reader{chunks: chunks}
}

// Read implements io.Reader
func (r *Reader) Read(p []byte) (n int, err error) {
	if r.buffer != "" {
		n = copy(p, r.buffer)
		r.buffer = r.buffer[n:]
		return n, nil
	}

	chunk, ok := <-r.chunks
	if !ok {
		return 0, io.EOF
	}
	if chunk.Error != nil {
		return 0, chunk.Error
	}
	r.buffer = chunk.Delta
	n = copy(p, r.buffer)
	r.buffer = r.buffer[n:]
	return n, nil
}

// StreamConfig holds configuration for streaming
type StreamConfig struct {
	// Timeout for the entire streaming operation
	Timeout int

	// ChunkTimeout is the maximum time to wait for each chunk
	ChunkTimeout int

	// RetryConfig controls retry behavior
	RetryConfig *RetryConfig
}

// RetryConfig controls retry behavior
type RetryConfig struct {
	// MaxRetries is the maximum number of retries
	MaxRetries int

	// InitialBackoff is the initial backoff duration in milliseconds
	InitialBackoff int

	// MaxBackoff is the maximum backoff duration in milliseconds
	MaxBackoff int

	// RetryableFunc determines if an error is retryable
	RetryableFunc func(error) bool
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1000,
		MaxBackoff:     10000,
		RetryableFunc:  DefaultRetryableFunc,
	}
}

// DefaultRetryableFunc returns true for common retryable errors
func DefaultRetryableFunc(err error) bool {
	if err == nil {
		return false
	}
	// TODO: Add specific error type checking
	// For now, retry all errors
	return true
}
