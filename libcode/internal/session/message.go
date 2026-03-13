// Package session provides session management types
package session

// Message represents a message in a session
type Message struct {
	ID        string      `json:"id"`
	Role      string      `json:"role"`
	Content   string      `json:"content"`
	Timestamp int64       `json:"timestamp"`
	Metadata  Metadata    `json:"metadata,omitempty"`
	Parts     []Part      `json:"parts,omitempty"`
}

// Metadata represents message metadata
type Metadata map[string]any

// Part represents a part of a message (for multimodal content)
type Part struct {
	Type   string `json:"type"`
	Text   string `json:"text,omitempty"`
	Data   []byte `json:"data,omitempty"`
	MIME   string `json:"mime,omitempty"`
}

// WithParts is a marker interface for messages with parts
type WithParts interface {
	HasParts() bool
}

// HasParts returns true if the message has parts
func (m *Message) HasParts() bool {
	return len(m.Parts) > 0
}
