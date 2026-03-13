package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseCRUD(t *testing.T) {
	// Create temp database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Test config operations
	t.Run("Config", func(t *testing.T) {
		// Set config
		err := db.SetConfig("test_key", "test_value")
		require.NoError(t, err)

		// Get config
		value, err := db.GetConfig("test_key")
		require.NoError(t, err)
		assert.Equal(t, "test_value", value)

		// Update config
		err = db.SetConfig("test_key", "new_value")
		require.NoError(t, err)

		value, err = db.GetConfig("test_key")
		require.NoError(t, err)
		assert.Equal(t, "new_value", value)

		// Delete config
		err = db.DeleteConfig("test_key")
		require.NoError(t, err)

		value, err = db.GetConfig("test_key")
		require.NoError(t, err)
		assert.Empty(t, value)
	})

	// Test session operations
	t.Run("Sessions", func(t *testing.T) {
		session := &Session{
			ID:        "test-session-1",
			Title:     "Test Session",
			Model:     "openai/gpt-4",
			Messages:  []Message{},
			Metadata:  []byte(`{"test": true}`),
		}
		session.CreatedAt = session.UpdatedAt
		session.UpdatedAt = session.CreatedAt

		// Create session
		err := db.CreateSession(session)
		require.NoError(t, err)

		// Get session
		retrieved, err := db.GetSession("test-session-1")
		require.NoError(t, err)
		assert.Equal(t, session.ID, retrieved.ID)
		assert.Equal(t, session.Title, retrieved.Title)
		assert.Equal(t, session.Model, retrieved.Model)

		// Update session
		retrieved.Title = "Updated Title"
		err = db.UpdateSession(retrieved)
		require.NoError(t, err)

		updated, err := db.GetSession("test-session-1")
		require.NoError(t, err)
		assert.Equal(t, "Updated Title", updated.Title)

		// List sessions
		sessions, err := db.ListSessions(10)
		require.NoError(t, err)
		assert.Len(t, sessions, 1)
		assert.Equal(t, "test-session-1", sessions[0].ID)

		// Delete session
		err = db.DeleteSession("test-session-1")
		require.NoError(t, err)

		_, err = db.GetSession("test-session-1")
		assert.Error(t, err)
	})
}

func TestDatabaseFileCreation(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "subdir", "test.db")

	// Should create directory
	db, err := Open(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Verify file exists
	_, err = os.Stat(dbPath)
	require.NoError(t, err)
}
