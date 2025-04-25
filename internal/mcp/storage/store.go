package storage

import (
	"context"
	"github.com/mcp-ecosystem/mcp-gateway/internal/common/config"
)

// Store defines the interface for MCP server storage operations
type Store interface {
	// Create creates a new MCP server record
	Create(ctx context.Context, server *config.MCPConfig) error

	// Get retrieves an MCP server by its ID
	Get(ctx context.Context, id string) (*config.MCPConfig, error)

	// List returns all MCP servers
	List(ctx context.Context) ([]*config.MCPConfig, error)

	// Update updates an existing MCP server
	Update(ctx context.Context, server *config.MCPConfig) error

	// Delete removes an MCP server by its ID
	Delete(ctx context.Context, id string) error

	// Watch returns a channel that receives notifications when servers are updated
	Watch(ctx context.Context) (<-chan *config.MCPConfig, error)
}
