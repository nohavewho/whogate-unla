package database

import (
	"context"
)

// Database defines the methods for database operations.
type Database interface {
	// Close closes the database connection.
	Close() error

	// SaveMessage saves a message to the database.
	SaveMessage(ctx context.Context, message *Message) error

	// GetMessages gets messages for a specific session.
	GetMessages(ctx context.Context, sessionID string) ([]*Message, error)

	// GetMessagesWithPagination gets messages for a specific session with pagination.
	GetMessagesWithPagination(ctx context.Context, sessionID string, page, pageSize int) ([]*Message, error)

	// CreateSession creates a new session with the given sessionId.
	CreateSession(ctx context.Context, sessionId string) error

	// SessionExists checks if a session exists.
	SessionExists(ctx context.Context, sessionID string) (bool, error)

	// GetSessions gets all chat sessions with their latest message.
	GetSessions(ctx context.Context) ([]*Session, error)

	// UpdateSessionTitle updates the title of a session.
	UpdateSessionTitle(ctx context.Context, sessionID string, title string) error
}
