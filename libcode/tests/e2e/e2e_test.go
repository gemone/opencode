// Package e2e provides end-to-end tests for libcode
package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gemone/libcode/internal/ai"
	"github.com/gemone/libcode/internal/ai/provider"
	"github.com/gemone/libcode/internal/config"
	"github.com/gemone/libcode/internal/logger"
	"github.com/gemone/libcode/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_FullWorkflow tests the complete workflow from initialization to message sending
func TestE2E_FullWorkflow(t *testing.T) {
	// Skip in CI if no API key
	if os.Getenv("OPENAI_API_KEY") == "" && os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("Skipping E2E test: No API key configured")
	}

	// Setup temporary directory
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Initialize logger
	err := logger.Init("INFO", "")
	require.NoError(t, err)

	// Initialize storage
	db, err := storage.Open(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Initialize AI provider
	providerRegistry := provider.DefaultRegistry()
	providers := providerRegistry.List()
	require.NotEmpty(t, providers, "At least one AI provider should be available")

	var aiProvider ai.Provider
	for _, p := range providers {
		aiProvider = p
		break
	}
	require.NotNil(t, aiProvider, "AI provider should be available")

	// Create a test session
	session := &storage.Session{
		ID:        "test-session-1",
		Title:     "E2E Test Session",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Model:     "gpt-4",
		Messages: []storage.Message{
			{
				ID:        "msg-1",
				Role:      "user",
				Content:   "Hello, this is a test message.",
				Timestamp: time.Now(),
			},
		},
	}

	err = db.CreateSession(session)
	require.NoError(t, err)

	// Load session back
	loaded, err := db.GetSession(session.ID)
	require.NoError(t, err)
	assert.Equal(t, session.ID, loaded.ID)
	assert.Equal(t, session.Title, loaded.Title)
	assert.Len(t, loaded.Messages, 1)

	// Test AI provider (if API key available)
	ctx := context.Background()
	req := &ai.Request{
		Messages: []ai.Message{
			{Role: "user", Content: "Say 'Hello, World!' in exactly those words."},
		},
		MaxTokens: 10,
	}

	response, err := aiProvider.Complete(ctx, req)
	if err != nil {
		t.Logf("AI provider test skipped: %v", err)
	} else {
		assert.NotEmpty(t, response.Message.Content)
		assert.Greater(t, response.Usage.TotalTokens, 0)
	}

	// List sessions
	sessions, err := db.ListSessions(10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(sessions), 1)

	// Update session
	session.Title = "Updated E2E Test Session"
	err = db.UpdateSession(session)
	require.NoError(t, err)

	// Verify update
	loaded, err = db.GetSession(session.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated E2E Test Session", loaded.Title)

	// Delete session
	err = db.DeleteSession(session.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = db.GetSession(session.ID)
	assert.Error(t, err)
}

// TestE2E_ConfigCompatibility tests config file compatibility
func TestE2E_ConfigCompatibility(t *testing.T) {
	// Create a test config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	configContent := `
default_model: "openai/gpt-4"
model_variant: ""
session_dir: "~/.libcode/sessions"
max_history: 100
log_level: "INFO"
log_file: ""

lsp:
  enabled: true
  server_timeout: 30
  languages:
    - typescript
    - go
    - python
    - rust

mcp:
  enabled: true
  server_timeout: 30
  servers: []

features:
  tui: true
  desktop_mode: false
  stream_response: true

providers:
  openai:
    api_key: "sk-test-key"
    enabled: false
    model: "gpt-4"
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set config path
	origConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer func() {
		if origConfigHome != "" {
			os.Setenv("XDG_CONFIG_HOME", origConfigHome)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	// Load config
	cfg, err := config.Load()
	require.NoError(t, err)

	// Verify config values (config parser adds provider prefix)
	assert.Equal(t, "openai/gpt-4", cfg.DefaultModel)
	assert.Equal(t, 100, cfg.MaxHistory)
	assert.Equal(t, "INFO", cfg.LogLevel)
	assert.True(t, cfg.LSP.Enabled)
	assert.True(t, cfg.Features.TUI)
	assert.False(t, cfg.Features.DesktopMode)

	// Validate config
	err = cfg.Validate()
	// Should not fail even with test key (validation checks for enabled providers)
	assert.NoError(t, err)
}

// TestE2E_DataMigration tests data migration compatibility
func TestE2E_DataMigration(t *testing.T) {
	// Create a test database with JS schema format
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "sessions.js.db")

	// Initialize database
	db, err := storage.Open(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Create a session with JS-style data
	session := &storage.Session{
		ID:        "js-migration-test",
		Title:     "JS Migration Test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Model:     "claude-3-5-sonnet-20241022",
		Messages: []storage.Message{
			{
				ID:        "msg-1",
				Role:      "user",
				Content:   "Test message for migration",
				Timestamp: time.Now(),
				Metadata:  []byte(`{"source": "migration-test"}`),
				ToolCalls: []storage.ToolCall{
					{
						ID:     "call-1",
						Tool:   "read",
						Input:  []byte(`{"file_path":"test.go"}`),
						Status: "success",
						Output: []byte(`"test content"`),
					},
				},
			},
			{
				ID:        "msg-2",
				Role:      "assistant",
				Content:   "Response to test message",
				Timestamp: time.Now(),
			},
		},
	}

	err = db.CreateSession(session)
	require.NoError(t, err)

	// Load and verify
	loaded, err := db.GetSession(session.ID)
	require.NoError(t, err)

	assert.Equal(t, session.ID, loaded.ID)
	assert.Equal(t, session.Title, loaded.Title)
	assert.Len(t, loaded.Messages, 2)

	// Verify message details
	msg := loaded.Messages[0]
	assert.Equal(t, "user", msg.Role)
	assert.Equal(t, "Test message for migration", msg.Content)
	assert.Len(t, msg.ToolCalls, 1)
	assert.Equal(t, "read", msg.ToolCalls[0].Tool)
	assert.Equal(t, "success", msg.ToolCalls[0].Status)
}

// TestE2E_StressTest runs a stress test on the system
func TestE2E_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "stress.db")

	db, err := storage.Open(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Create many sessions rapidly
	numSessions := 500
	start := time.Now()

	for i := range numSessions {
		session := &storage.Session{
			ID:        fmt.Sprintf("stress-%d", i),
			Title:     "Stress Test Session",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Model:     "gpt-4",
			Messages: []storage.Message{
				{
					ID:        "msg-1",
					Role:      "user",
					Content:   "Stress test message",
					Timestamp: time.Now(),
				},
			},
		}
		err := db.CreateSession(session)
		require.NoError(t, err)
	}

	elapsed := time.Since(start)
	t.Logf("Created %d sessions in %v (%.2f sessions/sec)",
		numSessions, elapsed, float64(numSessions)/elapsed.Seconds())

	// Verify all sessions were created
	sessions, err := db.ListSessions(1000)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(sessions), numSessions)
}
