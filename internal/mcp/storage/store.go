package storage

import (
	"context"
	"github.com/mcp-ecosystem/mcp-gateway/internal/mcp/storage/notifier"

	"github.com/mcp-ecosystem/mcp-gateway/internal/common/config"
)

// Store defines the interface for MCP server storage operations
type Store interface {
	// Create creates a new MCP server record
	Create(ctx context.Context, cfg *config.MCPConfig) error

	// Get retrieves an MCP server by its name
	Get(ctx context.Context, name string) (*config.MCPConfig, error)

	// List returns all MCP servers
	List(ctx context.Context) ([]*config.MCPConfig, error)

	// Update updates an existing MCP server
	Update(ctx context.Context, cfg *config.MCPConfig) error

	// Delete removes an MCP server by its name
	Delete(ctx context.Context, name string) error

	// GetNotifier returns the notifier implementation
	GetNotifier(ctx context.Context) notifier.Notifier
}
