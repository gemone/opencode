package event

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBusPublishSubscribe(t *testing.T) {
	bus := New()
	defer bus.Close()

	received := false
	var mu sync.Mutex

	handler := func(e Event) {
		mu.Lock()
		received = true
		mu.Unlock()
	}

	// Subscribe to event type
	bus.Subscribe("test.event", handler)

	// Publish event
	bus.Publish(Event{
		Type:   "test.event",
		Source: "test",
		Data:   "test data",
	})

	// Wait for async handler
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	assert.True(t, received)
	mu.Unlock()
}

func TestBusPublishSync(t *testing.T) {
	bus := New()
	defer bus.Close()

	received := false

	handler := func(e Event) {
		received = true
	}

	bus.Subscribe("test.sync", handler)

	// Publish synchronously
	bus.PublishSync(Event{
		Type:   "test.sync",
		Source: "test",
		Data:   "test data",
	})

	// Should be received immediately
	assert.True(t, received)
}

func TestBusUnsubscribe(t *testing.T) {
	bus := New()
	defer bus.Close()

	count := 0
	var mu sync.Mutex

	handler := func(e Event) {
		mu.Lock()
		count++
		mu.Unlock()
	}

	unsub := bus.Subscribe("test.unsub", handler)

	// Publish first event
	bus.PublishSync(Event{Type: "test.unsub"})
	mu.Lock()
	assert.Equal(t, 1, count)
	mu.Unlock()

	// Unsubscribe
	unsub()

	// Publish second event - should not increment
	bus.PublishSync(Event{Type: "test.unsub"})
	mu.Lock()
	assert.Equal(t, 1, count) // Still 1
	mu.Unlock()
}

func TestBusMultipleHandlers(t *testing.T) {
	bus := New()
	defer bus.Close()

	var results []string
	var mu sync.Mutex

	handler1 := func(e Event) {
		mu.Lock()
		results = append(results, "handler1")
		mu.Unlock()
	}

	handler2 := func(e Event) {
		mu.Lock()
		results = append(results, "handler2")
		mu.Unlock()
	}

	bus.Subscribe("test.multi", handler1)
	bus.Subscribe("test.multi", handler2)

	bus.PublishSync(Event{Type: "test.multi"})

	mu.Lock()
	assert.Len(t, results, 2)
	assert.Contains(t, results, "handler1")
	assert.Contains(t, results, "handler2")
	mu.Unlock()
}

func TestBusClosed(t *testing.T) {
	bus := New()

	handler := func(e Event) {}

	// Close the bus
	bus.Close()

	// Should panic when subscribing to closed bus
	assert.Panics(t, func() {
		bus.Subscribe("test.closed", handler)
	})

	// Publish should be safe (no-op)
	received := false
	bus.Publish(Event{Type: "test.closed"})
	time.Sleep(50 * time.Millisecond)
	assert.False(t, received)
}

func TestGlobalBus(t *testing.T) {
	received := false

	handler := func(e Event) {
		if e.Type == "test.global" {
			received = true
		}
	}

	Subscribe("test.global", handler)
	defer Unsubscribe("test.global", handler)

	Publish("test.global", "test", nil)
	time.Sleep(100 * time.Millisecond)

	assert.True(t, received)
}

func TestEventTypes(t *testing.T) {
	tests := []struct {
		name string
		typ  string
	}{
		{"session start", EventSessionStart},
		{"session end", EventSessionEnd},
		{"tool start", EventToolStart},
		{"file create", EventFileCreate},
		{"lsp launch", EventLSPLaunch},
		{"mcp connect", EventMCPConnect},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.typ)
			assert.Contains(t, tt.typ, ".")
		})
	}
}
