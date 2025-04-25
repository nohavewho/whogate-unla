package database

import (
	"context"
	"fmt"
	"time"

	"github.com/mcp-ecosystem/mcp-gateway/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MySQLDB implements the Database interface using MySQL
type MySQLDB struct {
	db  *gorm.DB
	cfg *config.DatabaseConfig
}

// NewMySQLDB creates a new MySQLDB instance
func NewMySQLDB(cfg *config.DatabaseConfig) *MySQLDB {
	return &MySQLDB{
		cfg: cfg,
	}
}

func (db *MySQLDB) Init(_ context.Context) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		db.cfg.User, db.cfg.Password, db.cfg.Host, db.cfg.Port, db.cfg.DBName)

	gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := gormDB.AutoMigrate(&Message{}, &Session{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	db.db = gormDB
	return nil
}

func (db *MySQLDB) Close() error {
	sqlDB, err := db.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (db *MySQLDB) SaveMessage(ctx context.Context, message *Message) error {
	return db.db.WithContext(ctx).Create(message).Error
}

func (db *MySQLDB) GetMessages(ctx context.Context, sessionID string) ([]*Message, error) {
	var messages []*Message
	err := db.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("timestamp asc").
		Find(&messages).Error
	return messages, err
}

func (db *MySQLDB) GetMessagesWithPagination(ctx context.Context, sessionID string, page, pageSize int) ([]*Message, error) {
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

func (db *MySQLDB) CreateSession(ctx context.Context, sessionId string) error {
	session := &Session{
		ID:        sessionId,
		CreatedAt: time.Now(),
	}
	return db.db.WithContext(ctx).Create(session).Error
}

func (db *MySQLDB) UpdateSessionTitle(ctx context.Context, sessionID string, title string) error {
	return db.db.WithContext(ctx).
		Model(&Session{}).
		Where("id = ?", sessionID).
		Update("title", title).Error
}

func (db *MySQLDB) SessionExists(ctx context.Context, sessionID string) (bool, error) {
	var count int64
	err := db.db.WithContext(ctx).
		Model(&Session{}).
		Where("id = ?", sessionID).
		Count(&count).Error
	return count > 0, err
}

func (db *MySQLDB) GetSessions(ctx context.Context) ([]*Session, error) {
	var sessions []*Session
	err := db.db.WithContext(ctx).
		Order("created_at desc").
		Find(&sessions).Error
	return sessions, err
}
