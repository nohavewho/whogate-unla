package database

import (
	"fmt"

	"github.com/mcp-ecosystem/mcp-gateway/internal/common/config"
)

// NewDatabase creates a new database based on configuration
func NewDatabase(cfg *config.DatabaseConfig) (Database, error) {
	switch cfg.Type {
	case "postgres":
		return NewPostgres(cfg)
	case "sqlite":
		return NewSQLite(cfg)
	case "mysql":
		return NewMySQL(cfg)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}
}
