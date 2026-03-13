// Package mcp provides the Model Context Protocol client implementation
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/gemone/libcode/internal/mcp/protocol"
)

// Status represents the connection status
type Status string

const (
	StatusConnected                 Status = "connected"
	StatusDisabled                 Status = "disabled"
	StatusFailed                   Status = "failed"
	StatusNeedsAuth                Status = "needs_auth"
	StatusNeedsClientRegistration  Status = "needs_client_registration"
)

// Transport is the interface for MCP transport implementations
type Transport interface {
	// Start starts the transport and begins processing messages
	Start(ctx context.Context) error
	// Send sends a JSON-RPC message over the transport
	Send(message *protocol.JSONRPC) error
	// Close closes the transport and releases resources
	Close() error
	// SetMessageHandler sets the handler for incoming messages
	SetMessageHandler(handler func(*protocol.JSONRPC))
}

// Client represents an MCP client
type Client struct {
	name       string
	version    string
	transport  Transport
	requestID  atomic.Int64
	capabilities *protocol.ServerCapabilities

	mu              sync.RWMutex
	pendingRequests map[any]chan *protocol.JSONRPC
	closed          bool
}

// ClientConfig holds configuration for creating a new client
type ClientConfig struct {
	Name    string
	Version string
}

// NewClient creates a new MCP client
func NewClient(config *ClientConfig) *Client {
	return &Client{
		name:            config.Name,
		version:         config.Version,
		pendingRequests: make(map[any]chan *protocol.JSONRPC),
	}
}

// Connect connects the client to a server using the given transport
func (c *Client) Connect(ctx context.Context, transport Transport) error {
	c.mu.Lock()
	if c.transport != nil {
		c.mu.Unlock()
		return fmt.Errorf("already connected")
	}
	c.transport = transport
	c.mu.Unlock()

	// Set message handler
	transport.SetMessageHandler(c.handleMessage)

	// Start the transport
	if err := transport.Start(ctx); err != nil {
		c.mu.Lock()
		c.transport = nil
		c.mu.Unlock()
		return fmt.Errorf("transport start failed: %w", err)
	}

	// Send initialize request
	initParams := &protocol.InitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities: protocol.ClientCapabilities{
			Resources: &protocol.ResourcesCapability{
				Subscribe:   false,
				ListChanged: true,
			},
			Tools: &protocol.ToolsCapability{
				ListChanged: true,
			},
			Prompts: &protocol.PromptsCapability{
				ListChanged: true,
			},
		},
		ClientInfo: protocol.ClientInfo{
			Name:    c.name,
			Version: c.version,
		},
	}

	initResult := &protocol.InitializeResult{}
	if err := c.callRequest(ctx, "initialize", initParams, initResult); err != nil {
		c.Close()
		return fmt.Errorf("initialize failed: %w", err)
	}

	c.capabilities = &initResult.Capabilities

	// Send initialized notification
	if err := c.sendNotification("notifications/initialized", nil); err != nil {
		c.Close()
		return fmt.Errorf("initialized notification failed: %w", err)
	}

	return nil
}

// Close closes the client connection
func (c *Client) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true

	transport := c.transport
	c.transport = nil
	c.mu.Unlock()

	// Close all pending requests
	c.mu.Lock()
	for _, ch := range c.pendingRequests {
		close(ch)
	}
	c.pendingRequests = make(map[any]chan *protocol.JSONRPC)
	c.mu.Unlock()

	if transport != nil {
		return transport.Close()
	}
	return nil
}

// ListTools lists all available tools from the server
func (c *Client) ListTools(ctx context.Context) (*protocol.ListToolsResult, error) {
	result := &protocol.ListToolsResult{}
	if err := c.callRequest(ctx, "tools/list", nil, result); err != nil {
		return nil, err
	}
	return result, nil
}

// CallTool calls a tool on the server
func (c *Client) CallTool(ctx context.Context, params *protocol.CallToolParams) (*protocol.CallToolResult, error) {
	result := &protocol.CallToolResult{}
	if err := c.callRequest(ctx, "tools/call", params, result); err != nil {
		return nil, err
	}
	return result, nil
}

// ListResources lists all available resources from the server
func (c *Client) ListResources(ctx context.Context) (*protocol.ListResourcesResult, error) {
	result := &protocol.ListResourcesResult{}
	if err := c.callRequest(ctx, "resources/list", nil, result); err != nil {
		return nil, err
	}
	return result, nil
}

// ReadResource reads a resource from the server
func (c *Client) ReadResource(ctx context.Context, params *protocol.ReadResourceParams) (*protocol.ReadResourceResult, error) {
	result := &protocol.ReadResourceResult{}
	if err := c.callRequest(ctx, "resources/read", params, result); err != nil {
		return nil, err
	}
	return result, nil
}

// ListPrompts lists all available prompts from the server
func (c *Client) ListPrompts(ctx context.Context) (*protocol.ListPromptsResult, error) {
	result := &protocol.ListPromptsResult{}
	if err := c.callRequest(ctx, "prompts/list", nil, result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetPrompt gets a prompt from the server
func (c *Client) GetPrompt(ctx context.Context, params *protocol.GetPromptParams) (*protocol.GetPromptResult, error) {
	result := &protocol.GetPromptResult{}
	if err := c.callRequest(ctx, "prompts/get", params, result); err != nil {
		return nil, err
	}
	return result, nil
}

// SetLevel sets the logging level for the server
func (c *Client) SetLevel(ctx context.Context, level string) error {
	params := &protocol.SetLevelParams{Level: level}
	return c.sendNotification("logging/setLevel", params)
}

// callRequest sends a request and waits for the response
func (c *Client) callRequest(ctx context.Context, method string, params any, result any) error {
	// Generate request ID
	id := c.requestID.Add(1)

	// Create response channel
	responseCh := make(chan *protocol.JSONRPC, 1)
	c.mu.Lock()
	c.pendingRequests[id] = responseCh
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.pendingRequests, id)
		c.mu.Unlock()
		close(responseCh)
	}()

	// Create and send request
	req, err := protocol.NewRequest(id, method, params)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	jrpc := &protocol.JSONRPC{
		JSONRPC: req.JSONRPC,
		ID:      req.ID,
		Method:  req.Method,
		Params:  req.Params,
	}

	c.mu.RLock()
	transport := c.transport
	c.mu.RUnlock()

	if transport == nil {
		return fmt.Errorf("not connected")
	}

	if err := transport.Send(jrpc); err != nil {
		return fmt.Errorf("send failed: %w", err)
	}

	// Wait for response
	select {
	case resp := <-responseCh:
		if resp.Error != nil {
			return fmt.Errorf("rpc error: %s", resp.Error.Message)
		}
		if resp.Result == nil {
			return fmt.Errorf("empty result")
		}
		// Unmarshal result into provided result pointer
		if result != nil {
			resultBytes, err := json.Marshal(resp.Result)
			if err != nil {
				return fmt.Errorf("failed to remarshal result: %w", err)
			}
			if err := json.Unmarshal(resultBytes, result); err != nil {
				return fmt.Errorf("failed to unmarshal result: %w", err)
			}
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// sendNotification sends a notification (no response expected)
func (c *Client) sendNotification(method string, params any) error {
	notif, err := protocol.NewNotification(method, params)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	jrpc := &protocol.JSONRPC{
		JSONRPC: notif.JSONRPC,
		Method:  notif.Method,
		Params:  notif.Params,
	}

	c.mu.RLock()
	transport := c.transport
	c.mu.RUnlock()

	if transport == nil {
		return fmt.Errorf("not connected")
	}

	return transport.Send(jrpc)
}

// handleMessage handles an incoming message from the transport
func (c *Client) handleMessage(msg *protocol.JSONRPC) {
	// If this is a response, route to pending request
	if msg.ID != nil && !msg.IsNotification() {
		c.mu.RLock()
		ch, ok := c.pendingRequests[msg.ID]
		c.mu.RUnlock()

		if ok {
			select {
			case ch <- msg:
			default:
				// Channel full or closed, drop
			}
		}
		return
	}

	// Handle notifications
	// For now, we just log notifications - they don't need responses
	// In a full implementation, you'd have a notification handler registry
}

// SetNotificationHandler registers a handler for notifications
// This is a simplified version - a full implementation would have typed handlers
func (c *Client) SetNotificationHandler(method string, handler func(*protocol.JSONRPC)) {
	// TODO: Implement notification handler registry
	_ = method
	_ = handler
}

// GetCapabilities returns the server's capabilities
func (c *Client) GetCapabilities() *protocol.ServerCapabilities {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.capabilities
}

// NewJSONRPCReader creates a reader that parses JSON-RPC messages from a stream
type JSONRPCReader struct {
	decoder *json.Decoder
}

// NewJSONRPCReader creates a new JSON-RPC reader
func NewJSONRPCReader(r io.Reader) *JSONRPCReader {
	return &JSONRPCReader{
		decoder: json.NewDecoder(r),
	}
}

// ReadMessage reads a single JSON-RPC message
func (r *JSONRPCReader) ReadMessage() (*protocol.JSONRPC, error) {
	var msg protocol.JSONRPC
	if err := r.decoder.Decode(&msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// NewJSONRPCWriter creates a writer that writes JSON-RPC messages to a stream
type JSONRPCWriter struct {
	encoder *json.Encoder
}

// NewJSONRPCWriter creates a new JSON-RPC writer
func NewJSONRPCWriter(w io.Writer) *JSONRPCWriter {
	return &JSONRPCWriter{
		encoder: json.NewEncoder(w),
	}
}

// WriteMessage writes a single JSON-RPC message
func (w *JSONRPCWriter) WriteMessage(msg *protocol.JSONRPC) error {
	return w.encoder.Encode(msg)
}
