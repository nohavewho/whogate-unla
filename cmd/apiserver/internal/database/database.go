package database

import (
	"context"
	"time"
)

// Message represents a chat message
type Message struct {
	ID        string    `json:"id"`
	SessionID string    `json:"sessionId"`
	Content   string    `json:"content"`
	Sender    string    `json:"sender"`
	Timestamp time.Time `json:"timestamp"`
}

// Database interface defines the methods for database operations
type Database interface {
	// Initialize the database connection
	Init(ctx context.Context) error

	// Close the database connection
	Close() error

	// Save a message to the database
	SaveMessage(ctx context.Context, message *Message) error

	// Get messages for a specific session
	GetMessages(ctx context.Context, sessionID string) ([]*Message, error)

	// Get messages for a specific session with pagination
	GetMessagesWithPagination(ctx context.Context, sessionID string, page, pageSize int) ([]*Message, error)

	// Create a new session with the given sessionId
	CreateSession(ctx context.Context, sessionId string) error

	// Check if a session exists
	SessionExists(ctx context.Context, sessionID string) (bool, error)

	// Get all chat sessions with their latest message
	GetSessions(ctx context.Context) ([]*Session, error)
}

// Session represents a chat session
type Session struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	Title       string    `json:"title"`
	LastMessage *Message  `json:"lastMessage,omitempty"`
}
