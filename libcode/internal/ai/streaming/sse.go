// Package streaming provides streaming infrastructure for AI responses
package streaming

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gemone/libcode/internal/ai"
)

// SSEParser parses Server-Sent Events from AI providers
type SSEParser struct {
	scanner *bufio.Scanner
}

// NewSSEParser creates a new SSE parser
func NewSSEParser(r io.Reader) *SSEParser {
	return &SSEParser{
		scanner: bufio.NewScanner(r),
	}
}

// Parse reads SSE events and returns them through a channel
func (p *SSEParser) Parse(ctx context.Context) <-chan *SSEEvent {
	ch := make(chan *SSEEvent)

	go func() {
		defer close(ch)

		var currentEvent *SSEEvent

		for p.scanner.Scan() {
			line := p.scanner.Text()

			// Check for context cancellation
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Empty line marks the end of an event
			if line == "" {
				if currentEvent != nil && currentEvent.Data != "" {
					ch <- currentEvent
				}
				currentEvent = nil
				continue
			}

			// Skip comments
			if strings.HasPrefix(line, ":") {
				continue
			}

			// Parse field
			if idx := strings.IndexByte(line, ':'); idx >= 0 {
				field := line[:idx]
				value := line[idx+1:]
				if len(value) > 0 && value[0] == ' ' {
					value = value[1:]
				}

				if currentEvent == nil {
					currentEvent = &SSEEvent{}
				}

				switch field {
				case "data":
					currentEvent.Data = value
				case "event":
					currentEvent.Event = value
				case "id":
					currentEvent.ID = value
				case "retry":
					currentEvent.Retry = value
				}
			}
		}

		// Send final event if exists
		if currentEvent != nil && currentEvent.Data != "" {
			ch <- currentEvent
		}

		if err := p.scanner.Err(); err != nil {
			ch <- &SSEEvent{Error: fmt.Errorf("scan error: %w", err)}
		}
	}()

	return ch
}

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
	// Event is the event type
	Event string

	// Data is the event data
	Data string

	// ID is the event ID
	ID string

	// Retry specifies the reconnection delay
	Retry string

	// Error contains any parsing error
	Error error
}

// ChunkBuffer assembles streaming chunks into complete responses
type ChunkBuffer struct {
	chunks    []*ai.Chunk
	content   strings.Builder
	toolCalls map[int]*ToolCallBuilder
	index     int
}

// ToolCallBuilder builds tool calls incrementally
type ToolCallBuilder struct {
	ID    string
	Tool  string
	Input strings.Builder
}

// NewChunkBuffer creates a new chunk buffer
func NewChunkBuffer() *ChunkBuffer {
	return &ChunkBuffer{
		chunks:    make([]*ai.Chunk, 0),
		toolCalls: make(map[int]*ToolCallBuilder),
	}
}

// Add adds a chunk to the buffer
func (b *ChunkBuffer) Add(chunk *ai.Chunk) {
	b.chunks = append(b.chunks, chunk)
	b.content.WriteString(chunk.Delta)
	b.index++

	// Handle tool call chunks
	for _, tc := range chunk.ToolCalls {
		if existing, ok := b.toolCalls[tc.Index]; ok {
			// Append to existing tool call
			if tc.Tool != "" {
				existing.Tool = tc.Tool
			}
			if len(tc.Input) > 0 {
				existing.Input.WriteString(string(tc.Input))
			}
		} else {
			// New tool call
			b.toolCalls[tc.Index] = &ToolCallBuilder{
				ID:   tc.ID,
				Tool: tc.Tool,
			}
		}
	}
}

// Content returns the accumulated content
func (b *ChunkBuffer) Content() string {
	return b.content.String()
}

// ToolCalls returns the accumulated tool calls
func (b *ChunkBuffer) ToolCalls() []ai.ToolCall {
	calls := make([]ai.ToolCall, 0, len(b.toolCalls))
	for _, tc := range b.toolCalls {
		calls = append(calls, ai.ToolCall{
			ID:    tc.ID,
			Tool:  tc.Tool,
			Input: []byte(tc.Input.String()),
		})
	}
	return calls
}

// Chunks returns all chunks
func (b *ChunkBuffer) Chunks() []*ai.Chunk {
	return b.chunks
}

// Reset clears the buffer
func (b *ChunkBuffer) Reset() {
	b.chunks = make([]*ai.Chunk, 0)
	b.content.Reset()
	b.toolCalls = make(map[int]*ToolCallBuilder)
	b.index = 0
}

// StreamManager manages active streaming connections
type StreamManager struct {
	streams map[string]*Stream
}

// Stream represents an active streaming connection
type Stream struct {
	ID        string
	Chunks    <-chan *ai.Chunk
	StartedAt time.Time
	Cancel    context.CancelFunc
}

// NewStreamManager creates a new stream manager
func NewStreamManager() *StreamManager {
	return &StreamManager{
		streams: make(map[string]*Stream),
	}
}

// Add adds a new stream
func (m *StreamManager) Add(id string, chunks <-chan *ai.Chunk, cancel context.CancelFunc) {
	m.streams[id] = &Stream{
		ID:        id,
		Chunks:    chunks,
		StartedAt: time.Now(),
		Cancel:    cancel,
	}
}

// Get retrieves a stream by ID
func (m *StreamManager) Get(id string) (*Stream, bool) {
	s, ok := m.streams[id]
	return s, ok
}

// Cancel cancels a stream
func (m *StreamManager) Cancel(id string) {
	if s, ok := m.streams[id]; ok {
		if s.Cancel != nil {
			s.Cancel()
		}
		delete(m.streams, id)
	}
}

// ActiveCount returns the number of active streams
func (m *StreamManager) ActiveCount() int {
	return len(m.streams)
}

// Cleanup removes inactive streams
func (m *StreamManager) Cleanup(timeout time.Duration) {
	cutoff := time.Now().Add(-timeout)
	for id, s := range m.streams {
		if s.StartedAt.Before(cutoff) {
			m.Cancel(id)
		}
	}
}
