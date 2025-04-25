package database

import (
	"context"
	"fmt"
	"github.com/mcp-ecosystem/mcp-gateway/internal/common/config"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SQLiteDB implements the Database interface using SQLite
type SQLiteDB struct {
	db  *gorm.DB
	cfg *config.DatabaseConfig
}

// NewSQLiteDB creates a new SQLiteDB instance
func NewSQLiteDB(cfg *config.DatabaseConfig) *SQLiteDB {
	return &SQLiteDB{
		cfg: cfg,
	}
}

func (db *SQLiteDB) Init(_ context.Context) error {
	dir := filepath.Dir(db.cfg.DBName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	gormDB, err := gorm.Open(sqlite.Open(db.cfg.DBName), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := gormDB.AutoMigrate(&Message{}, &Session{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	db.db = gormDB
	return nil
}

func (db *SQLiteDB) Close() error {
	sqlDB, err := db.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (db *SQLiteDB) SaveMessage(ctx context.Context, message *Message) error {
	return db.db.WithContext(ctx).Create(message).Error
}

func (db *SQLiteDB) GetMessages(ctx context.Context, sessionID string) ([]*Message, error) {
	var messages []*Message
	err := db.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("timestamp asc").
		Find(&messages).Error
	return messages, err
}

func (db *SQLiteDB) GetMessagesWithPagination(ctx context.Context, sessionID string, page, pageSize int) ([]*Message, error) {
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

func (db *SQLiteDB) CreateSession(ctx context.Context, sessionId string) error {
	session := &Session{
		ID:        sessionId,
		CreatedAt: time.Now(),
	}
	return db.db.WithContext(ctx).Create(session).Error
}

func (db *SQLiteDB) UpdateSessionTitle(ctx context.Context, sessionID string, title string) error {
	return db.db.WithContext(ctx).
		Model(&Session{}).
		Where("id = ?", sessionID).
		Update("title", title).Error
}

func (db *SQLiteDB) SessionExists(ctx context.Context, sessionID string) (bool, error) {
	var count int64
	err := db.db.WithContext(ctx).
		Model(&Session{}).
		Where("id = ?", sessionID).
		Count(&count).Error
	return count > 0, err
}

func (db *SQLiteDB) GetSessions(ctx context.Context) ([]*Session, error) {
	var sessions []*Session
	err := db.db.WithContext(ctx).
		Order("created_at desc").
		Find(&sessions).Error
	return sessions, err
}
