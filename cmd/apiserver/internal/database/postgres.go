package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mcp-ecosystem/mcp-gateway/internal/config"
)

// PostgresDB implements the Database interface using PostgreSQL
type PostgresDB struct {
	pool *pgxpool.Pool
	cfg  *config.DatabaseConfig
}

// NewPostgresDB creates a new PostgresDB instance
func NewPostgresDB(cfg *config.DatabaseConfig) *PostgresDB {
	return &PostgresDB{
		cfg: cfg,
	}
}

// Init initializes the database connection and creates necessary tables
func (db *PostgresDB) Init(ctx context.Context) error {
	config, err := pgxpool.ParseConfig(db.cfg.GetDSN())
	if err != nil {
		return fmt.Errorf("error parsing database config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("error creating connection pool: %w", err)
	}

	db.pool = pool

	// Create sessions table
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			title TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create sessions table: %w", err)
	}

	// Create messages table
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			session_id TEXT REFERENCES sessions(id) ON DELETE CASCADE,
			content TEXT,
			sender TEXT,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			tool_calls TEXT,
			tool_result TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create messages table: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *PostgresDB) Close() error {
	if db.pool != nil {
		db.pool.Close()
	}
	return nil
}

// Message represents a chat message
type Message struct {
	ID         string    `json:"id"`
	SessionID  string    `json:"session_id"`
	Content    string    `json:"content"`
	Sender     string    `json:"sender"`
	Timestamp  time.Time `json:"timestamp"`
	ToolCalls  string    `json:"toolCalls,omitempty"`
	ToolResult string    `json:"toolResult,omitempty"`
}

// SaveMessage saves a message to the database
func (db *PostgresDB) SaveMessage(ctx context.Context, message *Message) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO messages (id, session_id, content, sender, timestamp, tool_calls, tool_result)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, message.ID, message.SessionID, message.Content, message.Sender, message.Timestamp, message.ToolCalls, message.ToolResult)
	return err
}

// GetMessages retrieves all messages for a session
func (db *PostgresDB) GetMessages(ctx context.Context, sessionID string) ([]*Message, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, session_id, content, sender, timestamp, tool_calls, tool_result
		FROM messages
		WHERE session_id = $1
		ORDER BY timestamp ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Content, &msg.Sender, &msg.Timestamp, &msg.ToolCalls, &msg.ToolResult); err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}
	return messages, nil
}

// GetMessagesWithPagination retrieves messages for a specific session with pagination
func (db *PostgresDB) GetMessagesWithPagination(ctx context.Context, sessionID string, page, pageSize int) ([]*Message, error) {
	offset := (page - 1) * pageSize

	rows, err := db.pool.Query(ctx, `
		SELECT id, session_id, content, sender, timestamp, tool_calls
		FROM messages
		WHERE session_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`, sessionID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Content, &msg.Sender, &msg.Timestamp, &msg.ToolCalls)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}

	// Reverse the order of messages to maintain chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// CreateSession creates a new session with the given sessionId
func (db *PostgresDB) CreateSession(ctx context.Context, sessionId string) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO sessions (id, title)
		VALUES ($1, '')
	`, sessionId)
	return err
}

// UpdateSessionTitle updates the title of a session
func (db *PostgresDB) UpdateSessionTitle(ctx context.Context, sessionID string, title string) error {
	_, err := db.pool.Exec(ctx, `
		UPDATE sessions
		SET title = $1
		WHERE id = $2
	`, title, sessionID)
	return err
}

// SessionExists checks if a session exists
func (db *PostgresDB) SessionExists(ctx context.Context, sessionID string) (bool, error) {
	var exists bool
	err := db.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM sessions WHERE id = $1
		)
	`, sessionID).Scan(&exists)
	return exists, err
}

// GetSessions retrieves all chat sessions with their latest message
func (db *PostgresDB) GetSessions(ctx context.Context) ([]*Session, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT 
			id,
			created_at,
			title
		FROM sessions
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var session Session
		err := rows.Scan(
			&session.ID,
			&session.CreatedAt,
			&session.Title,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, &session)
	}

	return sessions, nil
}
