package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewApp(t *testing.T) {
	app := NewApp()
	assert.NotNil(t, app)
	assert.Nil(t, app.ctx)
	assert.Nil(t, app.cfg)
}

func TestAppError(t *testing.T) {
	err := &AppError{Message: "test error"}
	assert.Equal(t, "test error", err.Error())
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  *AppError
		msg  string
	}{
		{"AIProviderNotInitialized", ErrAIProviderNotInitialized, "AI provider not initialized"},
		{"ConfigNotInitialized", ErrConfigNotInitialized, "Config not initialized"},
		{"DBNotInitialized", ErrDBNotInitialized, "Database not initialized"},
		{"ToolRegistryNotInitialized", ErrToolRegistryNotInitialized, "Tool registry not initialized"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.msg, tt.err.Error())
		})
	}
}
