// Package mcp provides transport implementations for MCP
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/gemone/libcode/internal/mcp/protocol"
)

// SSETransport implements MCP transport over Server-Sent Events
type SSETransport struct {
	url        string
	client     *http.Client
	handler    func(*protocol.JSONRPC)
	mu         sync.Mutex
	running    bool
	lastEventID string
	headers    map[string]string
}

// SSEConfig holds configuration for SSE transport
type SSEConfig struct {
	URL     string
	Headers map[string]string
}

// NewSSETransport creates a new SSE transport
func NewSSETransport(config *SSEConfig) *SSETransport {
	return &SSETransport{
		url:     config.URL,
		client:  &http.Client{},
		headers: config.Headers,
	}
}

// Start starts the SSE transport
func (t *SSETransport) Start(ctx context.Context) error {
	t.mu.Lock()
	if t.running {
		t.mu.Unlock()
		return fmt.Errorf("transport already running")
	}
	t.mu.Unlock()

	// Start listening for SSE messages
	go t.listen(ctx)
	return nil
}

// Send sends a JSON-RPC message via HTTP POST
func (t *SSETransport) Send(message *protocol.JSONRPC) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return fmt.Errorf("transport not running")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", t.url, strings.NewReader(string(data)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Close closes the SSE transport
func (t *SSETransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return nil
	}

	t.running = false
	t.client.CloseIdleConnections()
	return nil
}

// SetMessageHandler sets the handler for incoming messages
func (t *SSETransport) SetMessageHandler(handler func(*protocol.JSONRPC)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.handler = handler
}

// listen for SSE messages
func (t *SSETransport) listen(ctx context.Context) {
	t.mu.Lock()
	t.running = true
	t.mu.Unlock()

	req, err := http.NewRequest("GET", t.url, nil)
	if err != nil {
		t.callHandlerError(fmt.Errorf("failed to create request: %w", err))
		return
	}

	// Set last event ID for resuming
	if t.lastEventID != "" {
		req.Header.Set("Last-Event-ID", t.lastEventID)
	}

	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := t.client.Do(req.WithContext(ctx))
	if err != nil {
		t.callHandlerError(fmt.Errorf("failed to connect: %w", err))
		t.mu.Lock()
		t.running = false
		t.mu.Unlock()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.callHandlerError(fmt.Errorf("connection failed with status %d", resp.StatusCode))
		t.mu.Lock()
		t.running = false
		t.mu.Unlock()
		return
	}

	// Parse SSE messages
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// SSE format: "data: <json>"
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// Parse JSON-RPC message
			var msg protocol.JSONRPC
			if err := json.Unmarshal([]byte(data), &msg); err != nil {
				// Invalid JSON, skip
				continue
			}

			t.mu.Lock()
			handler := t.handler
			t.mu.Unlock()

			if handler != nil {
				handler(&msg)
			}
		} else if strings.HasPrefix(line, "id: ") {
			t.lastEventID = strings.TrimPrefix(line, "id: ")
		}
	}

	t.mu.Lock()
	t.running = false
	t.mu.Unlock()
}

// callHandlerError calls the handler with an error message
func (t *SSETransport) callHandlerError(err error) {
	t.mu.Lock()
	handler := t.handler
	t.mu.Unlock()

	if handler != nil {
		errorMsg, _ := json.Marshal(map[string]any{
			"jsonrpc": "2.0",
			"error": map[string]any{
				"code":    -32000,
				"message": err.Error(),
			},
		})
		var msg protocol.JSONRPC
		json.Unmarshal(errorMsg, &msg)
		handler(&msg)
	}
}

// HTTPTransport implements MCP transport over HTTP/2 streaming
type HTTPTransport struct {
	url        string
	client     *http.Client
	handler    func(*protocol.JSONRPC)
	mu         sync.Mutex
	running    bool
	headers    map[string]string
}

// HTTPConfig holds configuration for HTTP transport
type HTTPConfig struct {
	URL     string
	Headers map[string]string
}

// NewHTTPTransport creates a new HTTP streaming transport
func NewHTTPTransport(config *HTTPConfig) *HTTPTransport {
	return &HTTPTransport{
		url:     config.URL,
		client:  &http.Client{},
		headers: config.Headers,
	}
}

// Start starts the HTTP transport
func (t *HTTPTransport) Start(ctx context.Context) error {
	t.mu.Lock()
	if t.running {
		t.mu.Unlock()
		return fmt.Errorf("transport already running")
	}
	t.mu.Unlock()

	// For HTTP streaming, we mainly use Send() for requests
	// Server pushes come via a separate SSE endpoint
	t.mu.Lock()
	t.running = true
	t.mu.Unlock()

	return nil
}

// Send sends a JSON-RPC message via HTTP POST
func (t *HTTPTransport) Send(message *protocol.JSONRPC) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return fmt.Errorf("transport not running")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", t.url, strings.NewReader(string(data)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read and handle response immediately
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var msg protocol.JSONRPC
	if err := json.Unmarshal(body, &msg); err == nil {
		if t.handler != nil {
			t.handler(&msg)
		}
	}

	return nil
}

// Close closes the HTTP transport
func (t *HTTPTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return nil
	}

	t.running = false
	t.client.CloseIdleConnections()
	return nil
}

// SetMessageHandler sets the handler for incoming messages
func (t *HTTPTransport) SetMessageHandler(handler func(*protocol.JSONRPC)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.handler = handler
}
