// Package lsp provides the Language Server Protocol client implementation
package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"

	"github.com/gemone/libcode/internal/mcp/protocol"
)

// Client represents an LSP client
type Client struct {
	name       string
	rootURI    string
	command    string
	args       []string
	env        []string
	process    *exec.Cmd
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	stderr     io.ReadCloser

	mu              sync.RWMutex
	requestID       atomic.Int64
	pendingRequests map[any]chan *protocol.JSONRPC

	capabilities *ServerCapabilities
	running     bool
}

// ClientConfig holds configuration for creating a new client
type ClientConfig struct {
	Name      string
	Language string
	RootURI   string
	Command   string
	Args      []string
	Env       []string
}

// NewClient creates a new LSP client
func NewClient(config *ClientConfig) *Client {
	return &Client{
		name:            config.Name,
		rootURI:         config.RootURI,
		command:         config.Command,
		args:            config.Args,
		env:             config.Env,
		pendingRequests: make(map[any]chan *protocol.JSONRPC),
	}
}

// Start starts the LSP server process
func (c *Client) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return fmt.Errorf("already running")
	}

	// Create command
	cmd := exec.Command(c.command)
	if len(c.args) > 0 {
		cmd = exec.Command(c.command, c.args[0:]...)
		cmd.Args = c.args
	}

	// Set environment
	if len(c.env) > 0 {
		cmd.Env = c.env
	}

	// Create pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	c.stdin = stdin
	c.stdout = stdout
	c.stderr = stderr
	c.process = cmd

	// Start the process
	if err := c.process.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	c.running = true

	// Start reading messages
	go c.readMessages()

	// Initialize
	if err := c.initialize(ctx); err != nil {
		c.Close()
		return fmt.Errorf("initialization failed: %w", err)
	}

	return nil
}

// Close closes the LSP client
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return nil
	}

	c.running = false

	// Close all pending requests
	for _, ch := range c.pendingRequests {
		close(ch)
	}
	c.pendingRequests = make(map[any]chan *protocol.JSONRPC)

	// Close pipes
	c.stdin.Close()
	c.stdout.Close()
	c.stderr.Close()

	// Kill process
	if c.process.Process != nil {
		c.process.Process.Kill()
		c.process.Wait()
	}

	return nil
}

// DidOpen notifies the server that a document was opened
func (c *Client) DidOpen(ctx context.Context, params *DidOpenTextDocumentParams) error {
	return c.sendNotification("textDocument/didOpen", params)
}

// DidChange notifies the server that a document was changed
func (c *Client) DidChange(ctx context.Context, params *DidChangeTextDocumentParams) error {
	return c.sendNotification("textDocument/didChange", params)
}

// DidClose notifies the server that a document was closed
func (c *Client) DidClose(ctx context.Context, params *DidCloseTextDocumentParams) error {
	return c.sendNotification("textDocument/didClose", params)
}

// Diagnostics notifies the server about diagnostics
func (c *Client) Diagnostics(ctx context.Context, params *DiagnosticsParams) error {
	return c.sendNotification("textDocument/publishDiagnostics", params)
}

// DocumentSymbol retrieves symbols in a document
func (c *Client) DocumentSymbol(ctx context.Context, params *DocumentSymbolParams) ([]DocumentSymbol, error) {
	result := &[]DocumentSymbol{}
	if err := c.callRequest(ctx, "textDocument/documentSymbol", params, result); err != nil {
		return nil, err
	}
	return *result, nil
}

// Definition retrieves the definition location
func (c *Client) Definition(ctx context.Context, params *DefinitionParams) ([]Location, error) {
	// Definition can be a single Location, array of Locations, or null
	var rawResult json.RawMessage
	if err := c.callRequestRaw(ctx, "textDocument/definition", params, &rawResult); err != nil {
		return nil, err
	}

	// Try to parse as single Location first
	var singleLoc Location
	if err := json.Unmarshal(rawResult, &singleLoc); err == nil {
		return []Location{singleLoc}, nil
	}

	// Try to parse as array of Locations
	var locArray []Location
	if err := json.Unmarshal(rawResult, &locArray); err == nil {
		return locArray, nil
	}

	// Null result - no definition found
	return []Location{}, nil
}

// References finds all references to a symbol
func (c *Client) References(ctx context.Context, params *ReferencesParams) ([]Location, error) {
	result := &[]Location{}
	if err := c.callRequest(ctx, "textDocument/references", params, result); err != nil {
		return nil, err
	}
	return *result, nil
}

// Hover retrieves hover information
func (c *Client) Hover(ctx context.Context, params *HoverParams) (*HoverResult, error) {
	result := &HoverResult{}
	if err := c.callRequest(ctx, "textDocument/hover", params, result); err != nil {
		return nil, err
	}
	return result, nil
}

// CodeAction retrieves code actions
func (c *Client) CodeAction(ctx context.Context, params *CodeActionParams) ([]CodeAction, error) {
	result := &[]CodeAction{}
	if err := c.callRequest(ctx, "textDocument/codeAction", params, result); err != nil {
		return nil, err
	}
	return *result, nil
}

// Rename renames a symbol
func (c *Client) Rename(ctx context.Context, params *RenameParams) (*WorkspaceEdit, error) {
	result := &RenameResult{}
	if err := c.callRequest(ctx, "textDocument/rename", params, result); err != nil {
		return nil, err
	}

	if result.WorkspaceEdit != nil {
		return result.WorkspaceEdit, nil
	}
	return nil, fmt.Errorf("rename returned no changes")
}

// Completion retrieves completion items
func (c *Client) Completion(ctx context.Context, params *CompletionParams) (*CompletionResult, error) {
	result := &CompletionResult{}
	if err := c.callRequest(ctx, "textDocument/completion", params, result); err != nil {
		return nil, err
	}
	return result, nil
}

// ExecuteCommand executes a command on the server
func (c *Client) ExecuteCommand(ctx context.Context, params *ExecuteCommandParams) (ExecuteCommandResult, error) {
	var result ExecuteCommandResult
	if err := c.callRequest(ctx, "workspace/executeCommand", params, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetCapabilities returns the server's capabilities
func (c *Client) GetCapabilities() *ServerCapabilities {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.capabilities
}

// initialize performs LSP initialization
func (c *Client) initialize(ctx context.Context) error {
	params := &InitializeParams{
		ProcessID: os.Getpid(),
		RootURI:   c.rootURI,
	}

	result := &InitializeResult{}
	if err := c.callRequest(ctx, "initialize", params, result); err != nil {
		return err
	}

	c.capabilities = &result.Capabilities

	// Send initialized notification
	if err := c.sendNotification("initialized", nil); err != nil {
		return err
	}

	return nil
}

// callRequest sends a request and waits for the response
func (c *Client) callRequest(ctx context.Context, method string, params any, result any) error {
	return c.callRequestRaw(ctx, method, params, result)
}

// callRequestRaw sends a request and returns the raw JSON response
func (c *Client) callRequestRaw(ctx context.Context, method string, params any, result any) error {
	id := c.requestID.Add(1)

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

	// Create request
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

	// Send
	if err := c.sendMessage(jrpc); err != nil {
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
		// Unmarshal result into provided result
		if err := json.Unmarshal(resp.Result, result); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
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

	return c.sendMessage(jrpc)
}

// sendMessage sends a JSON-RPC message
func (c *Client) sendMessage(msg *protocol.JSONRPC) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Write data and newline separately to avoid allocation from append
	if _, err = c.stdin.Write(data); err != nil {
		return err
	}
	_, err = c.stdin.Write([]byte{'\n'})
	return err
}

// readMessages reads messages from stdout
func (c *Client) readMessages() {
	decoder := json.NewDecoder(c.stdout)

	for {
		var msg protocol.JSONRPC
		if err := decoder.Decode(&msg); err != nil {
			c.mu.RLock()
			if !c.running {
				c.mu.RUnlock()
				break
			}
			c.mu.RUnlock()
			continue
		}

		c.handleMessage(&msg)
	}
}

// handleMessage handles an incoming message
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

	// Handle notifications (publishDiagnostics, etc.)
	// For now, we just log them - could be handled via callbacks
}

// IsRunning checks if the client is running
func (c *Client) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}
