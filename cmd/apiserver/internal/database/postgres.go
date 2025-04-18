package database

import (
	"context"
	"database/sql"
	"fmt"

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

	// Create tables if they don't exist
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS sessions (
			id VARCHAR(36) PRIMARY KEY,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS messages (
			id VARCHAR(36) PRIMARY KEY,
			session_id VARCHAR(36) REFERENCES sessions(id),
			content TEXT NOT NULL,
			sender VARCHAR(50) NOT NULL,
			timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("error creating tables: %w", err)
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

// SaveMessage saves a message to the database
func (db *PostgresDB) SaveMessage(ctx context.Context, message *Message) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO messages (id, session_id, content, sender, timestamp)
		VALUES ($1, $2, $3, $4, $5)
	`, message.ID, message.SessionID, message.Content, message.Sender, message.Timestamp)
	return err
}

// GetMessages retrieves all messages for a specific session
func (db *PostgresDB) GetMessages(ctx context.Context, sessionID string) ([]*Message, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, session_id, content, sender, timestamp
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
		err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Content, &msg.Sender, &msg.Timestamp)
		if err != nil {
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
		SELECT id, session_id, content, sender, timestamp
		FROM messages
		WHERE session_id = $1
		ORDER BY timestamp ASC
		LIMIT $2 OFFSET $3
	`, sessionID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Content, &msg.Sender, &msg.Timestamp)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}

// CreateSession creates a new session with the given sessionId
func (db *PostgresDB) CreateSession(ctx context.Context, sessionId string) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO sessions (id)
		VALUES ($1)
	`, sessionId)
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
			s.id,
			s.created_at,
			COALESCE(m.content, '') as title,
			m.id as message_id,
			m.content as message_content,
			m.sender as message_sender,
			m.timestamp as message_timestamp
		FROM sessions s
		LEFT JOIN LATERAL (
			SELECT id, content, sender, timestamp
			FROM messages
			WHERE session_id = s.id
			ORDER BY timestamp DESC
			LIMIT 1
		) m ON true
		ORDER BY s.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var session Session
		var messageID, messageContent, messageSender sql.NullString
		var messageTimestamp sql.NullTime

		err := rows.Scan(
			&session.ID,
			&session.CreatedAt,
			&session.Title,
			&messageID,
			&messageContent,
			&messageSender,
			&messageTimestamp,
		)
		if err != nil {
			return nil, err
		}

		if messageID.Valid {
			session.LastMessage = &Message{
				ID:        messageID.String,
				SessionID: session.ID,
				Content:   messageContent.String,
				Sender:    messageSender.String,
				Timestamp: messageTimestamp.Time,
			}
		}

		sessions = append(sessions, &session)
	}

	return sessions, nil
}
