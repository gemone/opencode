// Package storage provides database layer with SQLite
package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
	"github.com/gemone/libcode/internal/errors"
)

// DB wraps the SQLite database connection
type DB struct {
	db *sql.DB
}

// Session represents a stored session
type Session struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Model       string     `json:"model"`
	Messages    []Message  `json:"messages"`
	Metadata    json.RawMessage `json:"metadata"`
}

// Message represents a message in a session
type Message struct {
	ID        string          `json:"id"`
	Role      string          `json:"role"` // "user", "assistant", "system"
	Content   string          `json:"content"`
	Timestamp time.Time       `json:"timestamp"`
	ToolCalls []ToolCall      `json:"tool_calls,omitempty"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
}

// ToolCall represents a tool call in a message
type ToolCall struct {
	ID       string                 `json:"id"`
	Tool     string                 `json:"tool"`
	Input    json.RawMessage        `json:"input"`
	Output   json.RawMessage        `json:"output,omitempty"`
	Status   string                 `json:"status"` // "pending", "running", "success", "error"
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Config represents stored configuration
type Config struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Open opens or creates the SQLite database at the given path
func Open(dbPath string) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, errors.SessionError("", "failed to create database directory", err)
	}

	// Open database with SQLite connection string
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, errors.SessionError("", "failed to open database", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, errors.SessionError("", "failed to ping database", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	storage := &DB{db: db}

	// Initialize schema
	if err := storage.initSchema(); err != nil {
		db.Close()
		return nil, errors.SessionError("", "failed to initialize schema", err)
	}

	return storage, nil
}

// initSchema creates the database schema if it doesn't exist
func (db *DB) initSchema() error {
	schema := `
	-- Sessions table
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		model TEXT NOT NULL,
		messages TEXT,
		metadata TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_sessions_created ON sessions(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_sessions_updated ON sessions(updated_at DESC);

	-- Config table
	CREATE TABLE IF NOT EXISTS config (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);

	-- MCP OAuth tokens table
	CREATE TABLE IF NOT EXISTS mcp_tokens (
		server TEXT PRIMARY KEY,
		access_token TEXT NOT NULL,
		refresh_token TEXT,
		expires_at TEXT,
		updated_at TEXT NOT NULL
	);

	-- Tool cache table
	CREATE TABLE IF NOT EXISTS tool_cache (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		expires_at TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_tool_cache_expires ON tool_cache(expires_at);
	`

	_, err := db.db.Exec(schema)
	return err
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.db.Close()
}

// Session operations

// CreateSession creates a new session
func (db *DB) CreateSession(session *Session) error {
	messagesJSON, err := json.Marshal(session.Messages)
	if err != nil {
		return errors.SessionError(session.ID, "failed to marshal messages", err)
	}

	metadataJSON, err := json.Marshal(session.Metadata)
	if err != nil {
		return errors.SessionError(session.ID, "failed to marshal metadata", err)
	}

	query := `
		INSERT INTO sessions (id, title, created_at, updated_at, model, messages, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = db.db.Exec(query,
		session.ID,
		session.Title,
		session.CreatedAt.Format(time.RFC3339),
		session.UpdatedAt.Format(time.RFC3339),
		session.Model,
		string(messagesJSON),
		string(metadataJSON),
	)

	return err
}

// GetSession retrieves a session by ID
func (db *DB) GetSession(id string) (*Session, error) {
	query := `
		SELECT id, title, created_at, updated_at, model, messages, metadata
		FROM sessions WHERE id = ?
	`

	row := db.db.QueryRow(query, id)

	var session Session
	var messagesJSON, metadataJSON string
	var createdAtStr, updatedAtStr string

	err := row.Scan(
		&session.ID,
		&session.Title,
		&createdAtStr,
		&updatedAtStr,
		&session.Model,
		&messagesJSON,
		&metadataJSON,
	)

	if err == sql.ErrNoRows {
		return nil, errors.SessionError(id, "session not found", nil)
	}
	if err != nil {
		return nil, errors.SessionError(id, "failed to scan session", err)
	}

	session.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, errors.SessionError(id, "failed to parse created_at", err)
	}

	session.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, errors.SessionError(id, "failed to parse updated_at", err)
	}

	if err := json.Unmarshal([]byte(messagesJSON), &session.Messages); err != nil {
		return nil, errors.SessionError(id, "failed to unmarshal messages", err)
	}

	session.Metadata = json.RawMessage(metadataJSON)

	return &session, nil
}

// UpdateSession updates an existing session
func (db *DB) UpdateSession(session *Session) error {
	session.UpdatedAt = time.Now()

	messagesJSON, err := json.Marshal(session.Messages)
	if err != nil {
		return errors.SessionError(session.ID, "failed to marshal messages", err)
	}

	metadataJSON, err := json.Marshal(session.Metadata)
	if err != nil {
		return errors.SessionError(session.ID, "failed to marshal metadata", err)
	}

	query := `
		UPDATE sessions
		SET title = ?, updated_at = ?, model = ?, messages = ?, metadata = ?
		WHERE id = ?
	`

	_, err = db.db.Exec(query,
		session.Title,
		session.UpdatedAt.Format(time.RFC3339),
		session.Model,
		string(messagesJSON),
		string(metadataJSON),
		session.ID,
	)

	return err
}

// DeleteSession deletes a session
func (db *DB) DeleteSession(id string) error {
	query := `DELETE FROM sessions WHERE id = ?`
	_, err := db.db.Exec(query, id)
	return err
}

// ListSessions returns all sessions, ordered by updated_at DESC
func (db *DB) ListSessions(limit int) ([]*Session, error) {
	query := `
		SELECT id, title, created_at, updated_at, model
		FROM sessions
		ORDER BY updated_at DESC
		LIMIT ?
	`

	rows, err := db.db.Query(query, limit)
	if err != nil {
		return nil, errors.SessionError("", "failed to list sessions", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var session Session
		var createdAtStr, updatedAtStr string

		err := rows.Scan(
			&session.ID,
			&session.Title,
			&createdAtStr,
			&updatedAtStr,
			&session.Model,
		)

		if err != nil {
			return nil, errors.SessionError("", "failed to scan session row", err)
		}

		session.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, errors.SessionError("", "failed to parse created_at", err)
		}

		session.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, errors.SessionError("", "failed to parse updated_at", err)
		}

		sessions = append(sessions, &session)
	}

	return sessions, rows.Err()
}

// Config operations

// GetConfig retrieves a configuration value
func (db *DB) GetConfig(key string) (string, error) {
	query := `SELECT value FROM config WHERE key = ?`
	var value string
	err := db.db.QueryRow(query, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// SetConfig stores a configuration value
func (db *DB) SetConfig(key, value string) error {
	query := `
		INSERT INTO config (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = ?
	`

	now := time.Now().Format(time.RFC3339)
	_, err := db.db.Exec(query, key, value, now, value, now)
	return err
}

// DeleteConfig removes a configuration value
func (db *DB) DeleteConfig(key string) error {
	query := `DELETE FROM config WHERE key = ?`
	_, err := db.db.Exec(query, key)
	return err
}
