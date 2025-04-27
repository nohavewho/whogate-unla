package storage

import (
	"context"
	"github.com/mcp-ecosystem/mcp-gateway/internal/common/config"
)

// Store defines the interface for MCP server storage operations
type Store interface {
	// Create creates a new MCP server record
	Create(ctx context.Context, cfg *config.MCPConfig) error

	// Get retrieves an MCP server record by name
	Get(ctx context.Context, name string) (*config.MCPConfig, error)

	// List retrieves all MCP server records
	List(ctx context.Context) ([]*config.MCPConfig, error)

	// Update updates an existing MCP server record
	Update(ctx context.Context, cfg *config.MCPConfig) error

	// Delete deletes an MCP server record by name
	Delete(ctx context.Context, name string) error
}
