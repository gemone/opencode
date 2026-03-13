// Package mcp provides transport implementations for MCP
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/gemone/libcode/internal/mcp/protocol"
)

// StdioTransport implements MCP transport over stdio for local servers
type StdioTransport struct {
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	stderr     io.ReadCloser
	handler   func(*protocol.JSONRPC)
	mu        sync.Mutex
	running   bool
	processID int
}

// StdioConfig holds configuration for stdio transport
type StdioConfig struct {
	Command   string
	Args      []string
	Env       []string
	Directory string
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(config *StdioConfig) *StdioTransport {
	return &StdioTransport{
		cmd: exec.Command(config.Command, config.Args...),
	}
}

// Start starts the transport by spawning the process
func (t *StdioTransport) Start(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.running {
		return fmt.Errorf("transport already running")
	}

	// Set up pipes
	stdin, err := t.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := t.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := t.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Set environment if provided
	if len(t.cmd.Env) == 0 {
		t.cmd.Env = os.Environ()
	}

	t.stdin = stdin
	t.stdout = stdout
	t.stderr = stderr

	// Start the process
	if err := t.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	t.processID = t.cmd.Process.Pid
	t.running = true

	// Start reading from stdout in a goroutine
	go t.readMessages()

	// Start reading from stderr in a goroutine
	go t.readStderr()

	return nil
}

// Send sends a JSON-RPC message over stdin
func (t *StdioTransport) Send(message *protocol.JSONRPC) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return fmt.Errorf("transport not running")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	_, err = t.stdin.Write(append(data, '\n'))
	return err
}

// Close closes the transport and terminates the process
func (t *StdioTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return nil
	}

	t.running = false

	// Close pipes
	t.stdin.Close()
	t.stdout.Close()
	t.stderr.Close()

	// Kill the process if still running
	if t.cmd.Process != nil {
		t.cmd.Process.Kill()
		t.cmd.Wait()
	}

	return nil
}

// SetMessageHandler sets the handler for incoming messages
func (t *StdioTransport) SetMessageHandler(handler func(*protocol.JSONRPC)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.handler = handler
}

// ProcessID returns the process ID of the spawned process
func (t *StdioTransport) ProcessID() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.processID
}

// readMessages reads messages from stdout
func (t *StdioTransport) readMessages() {
	scanner := bufio.NewScanner(t.stdout)
	for scanner.Scan() {
		line := scanner.Bytes()
		var msg protocol.JSONRPC
		if err := json.Unmarshal(line, &msg); err != nil {
			// Invalid JSON, skip
			continue
		}

		t.mu.Lock()
		handler := t.handler
		t.mu.Unlock()

		if handler != nil {
			handler(&msg)
		}
	}
}

// readStderr reads from stderr and logs it
func (t *StdioTransport) readStderr() {
	scanner := bufio.NewScanner(t.stderr)
	for scanner.Scan() {
		// Log stderr output - could be sent to a logger
		_ = scanner.Text()
	}
}
