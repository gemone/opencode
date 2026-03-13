// Package event provides a publish-subscribe event bus for libcode
package event

import (
	"sync"
	"unsafe"
)

// Event represents an event in the system
type Event struct {
	// Type identifies the kind of event
	Type string
	// Source is the origin of the event
	Source string
	// Data holds the event payload
	Data interface{}
}

// Handler is a function that handles events
type Handler func(Event)

// Bus implements a publish-subscribe event bus
type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
	closed   bool
}

// New creates a new event bus
func New() *Bus {
	return &Bus{
		handlers: make(map[string][]Handler),
	}
}

// Subscribe registers a handler for events of the given type
func (b *Bus) Subscribe(eventType string, handler Handler) func() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		panic("cannot subscribe to closed bus")
	}

	b.handlers[eventType] = append(b.handlers[eventType], handler)

	// Return unsubscribe function
	return func() {
		b.Unsubscribe(eventType, handler)
	}
}

// Unsubscribe removes a handler for events of the given type
func (b *Bus) Unsubscribe(eventType string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	handlers := b.handlers[eventType]
	// Filter out the handler by creating a new slice
	newHandlers := make([]Handler, 0, len(handlers))
	for _, h := range handlers {
		// Compare function pointers using their addresses
		h1, h2 := getFuncPointer(h), getFuncPointer(handler)
		if h1 != h2 {
			newHandlers = append(newHandlers, h)
		}
	}
	b.handlers[eventType] = newHandlers
}

// getFuncPointer extracts the pointer to a function for comparison
func getFuncPointer(h Handler) uintptr {
	return uintptr(*(*unsafe.Pointer)(unsafe.Pointer(&h)))
}

// Publish sends an event to all subscribed handlers
func (b *Bus) Publish(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return
	}

	handlers := b.handlers[event.Type]
	for _, handler := range handlers {
		// Call handler in goroutine to avoid blocking
		go handler(event)
	}
}

// PublishSync sends an event to all subscribed handlers synchronously
func (b *Bus) PublishSync(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return
	}

	handlers := b.handlers[event.Type]
	for _, handler := range handlers {
		handler(event)
	}
}

// Close closes the event bus
func (b *Bus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.closed = true
	b.handlers = make(map[string][]Handler)
}

// Global event bus instance
var globalBus = New()

// Subscribe to an event type on the global bus
func Subscribe(eventType string, handler Handler) func() {
	return globalBus.Subscribe(eventType, handler)
}

// Unsubscribe from an event type on the global bus
func Unsubscribe(eventType string, handler Handler) {
	globalBus.Unsubscribe(eventType, handler)
}

// Publish an event on the global bus
func Publish(eventType string, source string, data interface{}) {
	globalBus.Publish(Event{
		Type:   eventType,
		Source: source,
		Data:   data,
	})
}

// PublishSync publishes an event synchronously on the global bus
func PublishSync(eventType string, source string, data interface{}) {
	globalBus.PublishSync(Event{
		Type:   eventType,
		Source: source,
		Data:   data,
	})
}

// Event type constants
const (
	// Session events
	EventSessionStart    = "session.start"
	EventSessionEnd      = "session.end"
	EventSessionMessage  = "session.message"
	EventSessionError    = "session.error"

	// Tool events
	EventToolStart       = "tool.start"
	EventToolEnd         = "tool.end"
	EventToolError       = "tool.error"

	// File events
	EventFileCreate      = "file.create"
	EventFileChange      = "file.change"
	EventFileDelete      = "file.delete"

	// LSP events
	EventLSPLaunch      = "lsp.launch"
	EventLSPReady       = "lsp.ready"
	EventLSPError       = "lsp.error"

	// MCP events
	EventMCPConnect     = "mcp.connect"
	EventMCPDisconnect  = "mcp.disconnect"
	EventMCPError       = "mcp.error"

	// Config events
	EventConfigChange   = "config.change"
	EventConfigReload   = "config.reload"

	// PTY events
	EventPtyCreated     = "pty.created"
	EventPtyUpdated     = "pty.updated"
	EventPtyExited      = "pty.exited"
	EventPtyDeleted     = "pty.deleted"
)
