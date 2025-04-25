package database

import (
	"context"
	"fmt"
	"time"

	"github.com/mcp-ecosystem/mcp-gateway/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgresDB implements the Database interface using PostgreSQL
type PostgresDB struct {
	db  *gorm.DB
	cfg *config.DatabaseConfig
}

// NewPostgresDB creates a new PostgresDB instance
func NewPostgresDB(cfg *config.DatabaseConfig) *PostgresDB {
	return &PostgresDB{
		cfg: cfg,
	}
}

// Init initializes the database connection and creates necessary tables
func (db *PostgresDB) Init(_ context.Context) error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		db.cfg.Host, db.cfg.User, db.cfg.Password, db.cfg.DBName, db.cfg.Port, db.cfg.SSLMode)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := gormDB.AutoMigrate(&Message{}, &Session{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	db.db = gormDB
	return nil
}

// Close closes the database connection
func (db *PostgresDB) Close() error {
	sqlDB, err := db.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
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
	return db.db.WithContext(ctx).Create(message).Error
}

// GetMessages retrieves all messages for a session
func (db *PostgresDB) GetMessages(ctx context.Context, sessionID string) ([]*Message, error) {
	var messages []*Message
	err := db.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("timestamp asc").
		Find(&messages).Error
	return messages, err
}

// GetMessagesWithPagination retrieves messages for a specific session with pagination
func (db *PostgresDB) GetMessagesWithPagination(ctx context.Context, sessionID string, page, pageSize int) ([]*Message, error) {
	var messages []*Message
	offset := (page - 1) * pageSize
	err := db.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("timestamp asc").
		Offset(offset).
		Limit(pageSize).
		Find(&messages).Error
	return messages, err
}

// CreateSession creates a new session with the given sessionId
func (db *PostgresDB) CreateSession(ctx context.Context, sessionId string) error {
	session := &Session{
		ID:        sessionId,
		CreatedAt: time.Now(),
	}
	return db.db.WithContext(ctx).Create(session).Error
}

// UpdateSessionTitle updates the title of a session
func (db *PostgresDB) UpdateSessionTitle(ctx context.Context, sessionID string, title string) error {
	return db.db.WithContext(ctx).
		Model(&Session{}).
		Where("id = ?", sessionID).
		Update("title", title).Error
}

// SessionExists checks if a session exists
func (db *PostgresDB) SessionExists(ctx context.Context, sessionID string) (bool, error) {
	var count int64
	err := db.db.WithContext(ctx).
		Model(&Session{}).
		Where("id = ?", sessionID).
		Count(&count).Error
	return count > 0, err
}

// GetSessions retrieves all chat sessions with their latest message
func (db *PostgresDB) GetSessions(ctx context.Context) ([]*Session, error) {
	var sessions []*Session
	err := db.db.WithContext(ctx).
		Order("created_at desc").
		Find(&sessions).Error
	return sessions, err
}
