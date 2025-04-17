package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/mcp-ecosystem/mcp-gateway/internal/config"
	"github.com/mcp-ecosystem/mcp-gateway/internal/template"
	"github.com/mcp-ecosystem/mcp-gateway/pkg/mcp"
	"go.uber.org/zap"
)

// Server represents the MCP server
type Server struct {
	logger   *zap.Logger
	store    Storage
	renderer *template.Renderer
	sessions sync.Map
	tools    []mcp.ToolSchema
	toolMap  map[string]*config.ToolConfig
}

// NewServer creates a new MCP server
func NewServer(logger *zap.Logger, store Storage) *Server {
	return &Server{
		logger:   logger,
		store:    store,
		renderer: template.NewRenderer(),
		tools:    make([]mcp.ToolSchema, 0),
		toolMap:  make(map[string]*config.ToolConfig),
	}
}

// RegisterRoutes registers routes with the given router
func (s *Server) RegisterRoutes(router *gin.Engine, cfg *config.Config) error {
	router.Use(s.loggerMiddleware())
	router.Use(s.recoveryMiddleware())

	// Initialize tool map and list
	for i := range cfg.Tools {
		tool := &cfg.Tools[i]
		s.toolMap[tool.Name] = tool
		s.tools = append(s.tools, tool.ToToolSchema())
	}

	prefixMap := make(map[string]string)
	for _, routerCfg := range cfg.Routers {
		prefixMap[routerCfg.Server] = routerCfg.Prefix
	}

	for _, serverCfg := range cfg.Servers {
		prefix, exists := prefixMap[serverCfg.Name]
		if !exists {
			return fmt.Errorf("no router prefix found for server: %s", serverCfg.Name)
		}

		group := router.Group(prefix)

		// Add SSE and message endpoints
		group.GET("/sse", s.handleSSE)
		group.POST("/message", s.handleMessage)
	}

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(_ context.Context) error {
	return nil
}

// LoadConfig loads the server configuration
func (s *Server) LoadConfig(cfg *config.Config) error {
	// Initialize tool map and list
	for i := range cfg.Tools {
		tool := &cfg.Tools[i]
		s.toolMap[tool.Name] = tool
		s.tools = append(s.tools, tool.ToToolSchema())
	}
	return nil
}
