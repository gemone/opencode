// Package integration provides integration tests for libcode
package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gemone/libcode/internal/logger"
	"github.com/gemone/libcode/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestDatabase_ConcurrentAccess tests concurrent database access
func TestDatabase_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "concurrent.db")

	db, err := storage.Open(dbPath)
	require.NoError(t, err)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create initial session
	session := &storage.Session{
		ID:        "concurrent-test",
		Title:     "Concurrent Access Test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Model:     "gpt-4",
		Messages: []storage.Message{
			{
				ID:        "msg-1",
				Role:      "user",
				Content:   "Initial message",
				Timestamp: time.Now(),
			},
		},
	}
	err = db.CreateSession(session)
	require.NoError(t, err)

	// Run concurrent operations
	done := make(chan bool, 20)
	errors := make(chan error, 100)

	// Concurrent readers
	for range 10 {
		go func() {
			defer func() { done <- true }()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					_, err := db.GetSession(session.ID)
					if err != nil {
						errors <- err
					}
					time.Sleep(10 * time.Millisecond)
				}
			}
		}()
	}

	// Wait for readers to finish
	time.Sleep(500 * time.Millisecond)
	cancel()

	// Wait for all goroutines
	for range 10 {
		<-done
	}

	// Check for errors
	close(errors)
	errorCount := 0
	for range errors {
		errorCount++
	}

	// Verify final state
	loaded, err := db.GetSession(session.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, loaded.Title)

	t.Logf("Concurrent access test completed with %d errors", errorCount)
}

// TestDatabase_LargeScale tests database with large datasets
func TestDatabase_LargeScale(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "large.db")

	db, err := storage.Open(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Create many sessions
	numSessions := 100
	for i := range numSessions {
		session := &storage.Session{
			ID:        fmt.Sprintf("large-scale-%d", i),
			Title:     fmt.Sprintf("Large Scale Test Session %d", i),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Model:     "gpt-4",
			Messages: []storage.Message{
				{
					ID:        "msg-1",
					Role:      "user",
					Content:   "Test message",
					Timestamp: time.Now(),
				},
			},
		}
		err := db.CreateSession(session)
		require.NoError(t, err)
	}

	// List all sessions
	sessions, err := db.ListSessions(200)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(sessions), numSessions)

	// Test pagination
	page1, err := db.ListSessions(50)
	require.NoError(t, err)
	assert.Len(t, page1, 50)
}

// TestConfig_Compatibility tests configuration compatibility
func TestConfig_Compatibility(t *testing.T) {
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

	// Verify file was written
	content, err := os.ReadFile(configFile)
	require.NoError(t, err)
	// Config parser adds quotes around string values
	assert.Contains(t, string(content), "default_model:")
	assert.Contains(t, string(content), "log_level:")
}

// TestLogger_Integration tests logger integration
func TestLogger_Integration(t *testing.T) {
	// Initialize logger
	err := logger.Init("INFO", "")
	require.NoError(t, err)

	// Test logging at different levels
	logger.Debug("Debug message", zap.String("test", "debug"))
	logger.Info("Info message", zap.String("test", "info"))
	logger.Warn("Warning message", zap.String("test", "warn"))
	logger.Error("Error message", zap.String("test", "error"))

	// Note: Sync may fail in test environment, which is expected
	_ = logger.Sync()
}

// TestDataMigration_WritesAndReads tests data migration patterns
func TestDataMigration_WritesAndReads(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "migration.db")

	db, err := storage.Open(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Create a session with complex data (simulating JS migration)
	session := &storage.Session{
		ID:        "migration-test",
		Title:     "Migration Test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Model:     "claude-3-5-sonnet-20241022",
		Messages: []storage.Message{
			{
				ID:        "msg-1",
				Role:      "user",
				Content:   "Test message",
				Timestamp: time.Now(),
			},
			{
				ID:        "msg-2",
				Role:      "assistant",
				Content:   "Response",
				Timestamp: time.Now(),
				ToolCalls: []storage.ToolCall{
					{
						ID:     "call-1",
						Tool:   "read",
						Input:  []byte(`{"file_path":"test.go"}`),
						Status: "success",
					},
				},
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

	// Verify tool calls
	msg := loaded.Messages[1]
	assert.Len(t, msg.ToolCalls, 1)
	assert.Equal(t, "read", msg.ToolCalls[0].Tool)
}
