// Package pty provides cross-platform pseudo-terminal support
package pty

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/creack/pty"
	"github.com/gemone/libcode/internal/event"
	"github.com/google/uuid"
)

// SessionStatus represents the status of a PTY session
type SessionStatus string

const (
	StatusRunning SessionStatus = "running"
	StatusExited  SessionStatus = "exited"
)

// Info represents information about a PTY session
type Info struct {
	ID      string        `json:"id"`
	Title   string        `json:"title"`
	Command string        `json:"command"`
	Args    []string      `json:"args"`
	Cwd     string        `json:"cwd"`
	Status  SessionStatus `json:"status"`
	Pid     int           `json:"pid"`
}

// CreateInput holds input for creating a PTY session
type CreateInput struct {
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Cwd     string            `json:"cwd,omitempty"`
	Title   string            `json:"title,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// UpdateInput holds input for updating a PTY session
type UpdateInput struct {
	Title string        `json:"title,omitempty"`
	Size  *TerminalSize `json:"size,omitempty"`
}

// TerminalSize represents terminal dimensions
type TerminalSize struct {
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
}

// Session represents an active PTY session
type Session struct {
	mu        sync.RWMutex
	info      *Info
	process   *os.Process
	ptyFile   *os.File
	buffer    *ringBuffer
	cursor    atomic.Int64
	closed    atomic.Bool
	cancel    context.CancelFunc
	exitCode  atomic.Int32
}

// ringBuffer implements a fixed-size circular buffer
type ringBuffer struct {
	mu     sync.Mutex
	data   []byte
	size   int
	write  int
	cursor int
}

const (
	bufferLimit = 2 * 1024 * 1024 // 2MB
	bufferChunk = 64 * 1024       // 64KB
	metaPrefix  = 0x00
)

// Manager manages PTY sessions
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	bus      *event.Bus
}

// NewManager creates a new PTY manager
func NewManager(bus *event.Bus) *Manager {
	m := &Manager{
		sessions: make(map[string]*Session),
		bus:      bus,
	}
	return m
}

// Close closes all sessions
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, session := range m.sessions {
		session.Close()
	}
	m.sessions = make(map[string]*Session)
	return nil
}

// List returns all session info
func (m *Manager) List() []*Info {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Info, 0, len(m.sessions))
	for _, session := range m.sessions {
		result = append(result, session.Info())
	}
	return result
}

// Get returns session info by ID
func (m *Manager) Get(id string) (*Info, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[id]
	if !ok {
		return nil, false
	}
	return session.Info(), true
}

// Create creates a new PTY session
func (m *Manager) Create(ctx context.Context, input *CreateInput) (*Info, error) {
	id := uuid.New().String()
	command := input.Command
	if command == "" {
		command = detectShell()
	}

	args := input.Args
	if args == nil {
		args = []string{}
	}

	// Add login flag for sh/bash
	if strings.HasSuffix(command, "sh") || strings.HasSuffix(command, "bash") {
		args = append([]string{"-l"}, args...)
	}

	cwd := input.Cwd
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	// Prepare environment
	env := prepareEnvironment(input.Env)

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = cwd
	cmd.Env = env

	// Start PTY
	ptyFile, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to start pty: %w", err)
	}

	info := &Info{
		ID:      id,
		Title:   input.Title,
		Command: command,
		Args:    args,
		Cwd:     cwd,
		Status:  StatusRunning,
		Pid:     cmd.Process.Pid,
	}

	if info.Title == "" {
		info.Title = fmt.Sprintf("Terminal %s", id[len(id)-4:])
	}

	ctx, cancel := context.WithCancel(ctx)
	session := &Session{
		info:    info,
		process: cmd.Process,
		ptyFile: ptyFile,
		buffer:  newRingBuffer(bufferLimit),
		cancel:  cancel,
	}

	// Store session
	m.mu.Lock()
	m.sessions[id] = session
	m.mu.Unlock()

	// Start reading output
	go session.readOutput()
	go session.waitForExit(m, id)

	// Publish event
	m.bus.Publish(event.Event{
		Type: event.EventPtyCreated,
		Data: map[string]interface{}{
			"info": info,
		},
	})

	return info, nil
}

// Update updates a session
func (m *Manager) Update(id string, input *UpdateInput) (*Info, error) {
	m.mu.Lock()
	session, ok := m.sessions[id]
	if !ok {
		m.mu.Unlock()
		return nil, fmt.Errorf("session not found: %s", id)
	}
	m.mu.Unlock()

	session.mu.Lock()
	if input.Title != "" {
		session.info.Title = input.Title
	}
	if input.Size != nil {
		err := pty.Setsize(session.ptyFile, &pty.Winsize{
			Rows: input.Size.Rows,
			Cols: input.Size.Cols,
		})
		if err != nil {
			session.mu.Unlock()
			return nil, fmt.Errorf("failed to resize: %w", err)
		}
	}
	session.mu.Unlock()

	// Publish event
	m.bus.Publish(event.Event{
		Type: event.EventPtyUpdated,
		Data: map[string]interface{}{
			"info": session.Info(),
		},
	})

	return session.Info(), nil
}

// Remove removes a session
func (m *Manager) Remove(id string) error {
	m.mu.Lock()
	session, ok := m.sessions[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("session not found: %s", id)
	}
	delete(m.sessions, id)
	m.mu.Unlock()

	session.Close()

	// Publish event
	m.bus.Publish(event.Event{
		Type: event.EventPtyDeleted,
		Data: map[string]interface{}{
			"id": id,
		},
	})

	return nil
}

// Resize resizes a session terminal
func (m *Manager) Resize(id string, cols, rows uint16) error {
	m.mu.RLock()
	session, ok := m.sessions[id]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}

	return session.Resize(cols, rows)
}

// Write writes data to a session
func (m *Manager) Write(id string, data string) error {
	m.mu.RLock()
	session, ok := m.sessions[id]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}

	return session.Write(data)
}

// Connect connects to a session and returns output from cursor position
func (m *Manager) Connect(id string, cursor int64) (*Connection, error) {
	m.mu.RLock()
	session, ok := m.sessions[id]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found: %s", id)
	}

	return session.Connect(cursor), nil
}

// Info returns session info
func (s *Session) Info() *Info {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info := *s.info
	return &info
}

// Resize resizes the terminal
func (s *Session) Resize(cols, rows uint16) error {
	s.mu.RLock()
	if s.info.Status != StatusRunning {
		s.mu.RUnlock()
		return fmt.Errorf("session not running")
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	return pty.Setsize(s.ptyFile, &pty.Winsize{
		Rows: rows,
		Cols: cols,
	})
}

// Write writes data to the PTY
func (s *Session) Write(data string) error {
	s.mu.RLock()
	if s.info.Status != StatusRunning {
		s.mu.RUnlock()
		return fmt.Errorf("session not running")
	}
	ptf := s.ptyFile
	s.mu.RUnlock()

	_, err := io.WriteString(ptf, data)
	return err
}

// Close closes the session
func (s *Session) Close() error {
	if !s.closed.CompareAndSwap(false, true) {
		return nil // Already closed
	}

	s.cancel()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ptyFile != nil {
		s.ptyFile.Close()
	}

	if s.process != nil {
		s.process.Kill()
		s.process.Wait()
	}

	return nil
}

// readOutput reads from PTY and stores in buffer
func (s *Session) readOutput() {
	buf := make([]byte, 32*1024)

	for {
		n, err := s.ptyFile.Read(buf)
		if err != nil {
			if !s.closed.Load() {
				s.Close()
			}
			return
		}

		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buf[:n])

			s.cursor.Add(int64(n))
			s.buffer.Write(chunk)
		}
	}
}

// waitForExit waits for process to exit
func (s *Session) waitForExit(m *Manager, id string) {
	state, _ := s.process.Wait()
	exitCode := int32(0)
	if state != nil {
		exitCode = int32(state.ExitCode())
	}
	s.exitCode.Store(exitCode)

	s.mu.Lock()
	wasExited := s.info.Status == StatusExited
	s.info.Status = StatusExited
	s.mu.Unlock()

	if !wasExited {
		m.bus.Publish(event.Event{
			Type: event.EventPtyExited,
			Data: map[string]interface{}{
				"id":       id,
				"exitCode": exitCode,
			},
		})
		s.Close()
	}
}

// Connection represents a connection to a PTY session
type Connection struct {
	session   *Session
	cursor    int64
	chunkSize int
}

// ReadFromCursor returns data from cursor position
func (c *Connection) ReadFromCursor() []byte {
	return c.session.buffer.ReadFrom(c.cursor)
}

// GetCursor returns current cursor position
func (c *Connection) GetCursor() int64 {
	return c.session.cursor.Load()
}

// Write writes data to the session
func (c *Connection) Write(data string) error {
	return c.session.Write(data)
}

// Close closes the connection
func (c *Connection) Close() error {
	return nil
}

// Connect creates a new connection to the session
func (s *Session) Connect(cursor int64) *Connection {
	return &Connection{
		session:   s,
		cursor:    cursor,
		chunkSize: bufferChunk,
	}
}

// newRingBuffer creates a new ring buffer
func newRingBuffer(size int) *ringBuffer {
	return &ringBuffer{
		data: make([]byte, size),
		size: size,
	}
}

// Write writes data to the buffer
func (rb *ringBuffer) Write(data []byte) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	for _, b := range data {
		rb.data[rb.write] = b
		rb.write = (rb.write + 1) % rb.size
		if rb.write == rb.cursor {
			rb.cursor = (rb.cursor + 1) % rb.size
		}
	}
}

// ReadFrom reads data from cursor position
func (rb *ringBuffer) ReadFrom(cursor int64) []byte {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	currentCursor := rb.write

	if cursor < 0 {
		cursor = int64(currentCursor)
	}

	// Calculate how much data we have
	startOffset := int(cursor) % rb.size
	endOffset := currentCursor

	var result []byte

	if startOffset <= endOffset {
		// Simple case: no wrap
		result = make([]byte, endOffset-startOffset)
		copy(result, rb.data[startOffset:endOffset])
	} else {
		// Wrapped case
		firstLen := rb.size - startOffset
		result = make([]byte, firstLen+endOffset)
		copy(result, rb.data[startOffset:])
		copy(result[firstLen:], rb.data[:endOffset])
	}

	return result
}

// detectShell detects the default shell
func detectShell() string {
	// Check common shell locations
	shells := []string{
		"/bin/zsh",
		"/bin/bash",
		"/bin/sh",
		"/usr/bin/zsh",
		"/usr/bin/bash",
		"/usr/bin/sh",
	}

	for _, shell := range shells {
		if _, err := os.Stat(shell); err == nil {
			return shell
		}
	}

	// Fallback to sh
	return "/bin/sh"
}

// prepareEnvironment prepares environment variables for the PTY
func prepareEnvironment(extra map[string]string) []string {
	env := os.Environ()

	// Add PTY-specific environment
	env = append(env, "TERM=xterm-256color")
	env = append(env, "OPENCODE_TERMINAL=1")

	// Add extra environment
	for k, v := range extra {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// Windows-specific locale settings
	if runtime.GOOS == "windows" {
		env = append(env, "LC_ALL=C.UTF-8")
		env = append(env, "LC_CTYPE=C.UTF-8")
		env = append(env, "LANG=C.UTF-8")
	}

	return env
}

// MetaMessage creates a meta message with cursor position
func MetaMessage(cursor int64) []byte {
	data := map[string]interface{}{
		"cursor": cursor,
	}
	jsonBytes, _ := json.Marshal(data)

	result := make([]byte, len(jsonBytes)+1)
	result[0] = metaPrefix
	copy(result[1:], jsonBytes)

	return result
}
