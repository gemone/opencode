package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistry(t *testing.T) {
	reg := NewRegistry()

	// Test getting non-existent provider
	_, err := reg.Get("nonexistent")
	assert.Error(t, err)

	// Test empty list
	list := reg.List()
	assert.Empty(t, list)

	// Test default registry
	defaultReg := DefaultRegistry()
	list = defaultReg.List()
	// May have openai if env var is set
	assert.NotNil(t, list)
}
