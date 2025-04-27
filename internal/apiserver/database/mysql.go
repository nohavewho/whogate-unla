package database

import (
	"context"
	"fmt"
	"time"

	"github.com/mcp-ecosystem/mcp-gateway/internal/common/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MySQL implements the Database interface using MySQL
type MySQL struct {
	db  *gorm.DB
	cfg *config.DatabaseConfig
}

// NewMySQL creates a new MySQL instance
func NewMySQL(cfg *config.DatabaseConfig) (Database, error) {
	db := &MySQL{
		cfg: cfg,
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		db.cfg.User, db.cfg.Password, db.cfg.Host, db.cfg.Port, db.cfg.DBName)

	gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := gormDB.AutoMigrate(&Message{}, &Session{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	db.db = gormDB
	return db, nil
}

// Close closes the database connection
func (db *MySQL) Close() error {
	sqlDB, err := db.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (db *MySQL) SaveMessage(ctx context.Context, message *Message) error {
	return db.db.WithContext(ctx).Create(message).Error
}

func (db *MySQL) GetMessages(ctx context.Context, sessionID string) ([]*Message, error) {
	var messages []*Message
	err := db.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("timestamp asc").
		Find(&messages).Error
	return messages, err
}

func (db *MySQL) GetMessagesWithPagination(ctx context.Context, sessionID string, page, pageSize int) ([]*Message, error) {
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

func (db *MySQL) CreateSession(ctx context.Context, sessionId string) error {
	session := &Session{
		ID:        sessionId,
		CreatedAt: time.Now(),
	}
	return db.db.WithContext(ctx).Create(session).Error
}

func (db *MySQL) UpdateSessionTitle(ctx context.Context, sessionID string, title string) error {
	return db.db.WithContext(ctx).
		Model(&Session{}).
		Where("id = ?", sessionID).
		Update("title", title).Error
}

func (db *MySQL) SessionExists(ctx context.Context, sessionID string) (bool, error) {
	var count int64
	err := db.db.WithContext(ctx).
		Model(&Session{}).
		Where("id = ?", sessionID).
		Count(&count).Error
	return count > 0, err
}

func (db *MySQL) GetSessions(ctx context.Context) ([]*Session, error) {
	var sessions []*Session
	err := db.db.WithContext(ctx).
		Order("created_at desc").
		Find(&sessions).Error
	return sessions, err
}
