package streaming

import (
	"context"
	"strings"
	"testing"

	"github.com/gemone/libcode/internal/ai"
	"github.com/stretchr/testify/assert"
)

func TestSSEParser(t *testing.T) {
	input := `event: message.delta
data: {"delta":{"content":"Hello"}}

event: message.delta
data: {"delta":{"content":" world"}}

data: [DONE]

event: error
data: {"error":"rate limit"}`

	parser := NewSSEParser(strings.NewReader(input))
	events := parser.Parse(context.Background())

	var eventList []*SSEEvent
	for event := range events {
		if event.Error != nil {
			t.Logf("Parse error: %v", event.Error)
			break
		}
		// Filter out [DONE] which signals end of stream
		if event.Data == "[DONE]" {
			continue
		}
		eventList = append(eventList, event)
		t.Logf("Event: type=%s, data=%s", event.Event, event.Data)
	}

	// We expect 3 data events (2 message.delta + 1 error)
	assert.Len(t, eventList, 3, "Expected 3 events, got %d: %v", len(eventList), eventList)

	// First event - Hello
	assert.Equal(t, "message.delta", eventList[0].Event)
	assert.Equal(t, `{"delta":{"content":"Hello"}}`, eventList[0].Data)

	// Second event - world
	assert.Equal(t, "message.delta", eventList[1].Event)
	assert.Equal(t, `{"delta":{"content":" world"}}`, eventList[1].Data)

	// Third event - error
	assert.Equal(t, "error", eventList[2].Event)
	assert.Contains(t, eventList[2].Data, "rate limit")
}

func TestChunkBuffer(t *testing.T) {
	buffer := NewChunkBuffer()

	// Add content chunks
	buffer.Add(&ai.Chunk{Delta: "Hello"})
	buffer.Add(&ai.Chunk{Delta: " "})
	buffer.Add(&ai.Chunk{Delta: "World"})

	assert.Equal(t, "Hello World", buffer.Content())
	assert.Equal(t, 3, buffer.index)

	// Reset
	buffer.Reset()
	assert.Equal(t, "", buffer.Content())
	assert.Equal(t, 0, buffer.index)
}

func TestStreamManager(t *testing.T) {
	manager := NewStreamManager()

	cancel := func() {}
	chunks := make(chan *ai.Chunk, 1)
	close(chunks)

	manager.Add("test-1", chunks, cancel)

	// Check stream exists
	stream, ok := manager.Get("test-1")
	assert.True(t, ok)
	assert.NotNil(t, stream)
	assert.Equal(t, "test-1", stream.ID)

	// Check active count
	assert.Equal(t, 1, manager.ActiveCount())

	// Cancel stream
	manager.Cancel("test-1")

	// Check removed
	_, ok = manager.Get("test-1")
	assert.False(t, ok)
	assert.Equal(t, 0, manager.ActiveCount())
}
