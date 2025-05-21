package core

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mcp-ecosystem/mcp-gateway/internal/common/config"
	"github.com/mcp-ecosystem/mcp-gateway/internal/core/mcpproxy"
	"github.com/mcp-ecosystem/mcp-gateway/internal/mcp/session"
	"github.com/mcp-ecosystem/mcp-gateway/internal/mcp/storage/helper"
	"github.com/mcp-ecosystem/mcp-gateway/pkg/mcp"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type (
	// Server represents the MCP server
	Server struct {
		logger *zap.Logger
		cfg    *config.MCPGatewayConfig
		router *gin.Engine
		// state contains all the read-only shared state
		state *State
		// sessions manages all active sessions
		sessions session.Store
		// shutdownCh is used to signal shutdown to all SSE connections
		shutdownCh chan struct{}
		// toolRespHandler is a chain of response handlers
		toolRespHandler ResponseHandler
	}
)

// NewServer creates a new MCP server
func NewServer(logger *zap.Logger, cfg *config.MCPGatewayConfig) (*Server, error) {
	// Initialize session store
	sessionStore, err := session.NewStore(logger, &cfg.Session)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize session store: %w", err)
	}

	router := gin.Default()
	router.GET("/health_check", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Health check passed.",
		})
	})

	return &Server{
		logger:          logger,
		cfg:             cfg,
		router:          router,
		state:           NewState(),
		sessions:        sessionStore,
		shutdownCh:      make(chan struct{}),
		toolRespHandler: CreateResponseHandlerChain(),
	}, nil
}

// RegisterRoutes registers routes with the given router for MCP servers
func (s *Server) RegisterRoutes(ctx context.Context, cfgs []*config.MCPConfig) error {
	// Validate configuration before registering routes
	if err := config.ValidateMCPConfigs(cfgs); err != nil {
		s.logger.Error("invalid configuration during route registration",
			zap.Error(err))
		return fmt.Errorf("invalid configuration: %w", err)
	}

	s.logger.Info("registering middleware")
	s.router.Use(s.loggerMiddleware())
	s.router.Use(s.recoveryMiddleware())

	// Create new state and load configuration
	s.logger.Debug("initializing server state")
	newState, err := BuildStateFromConfig(ctx, cfgs, s.state, s.logger)
	if err != nil {
		s.logger.Error("failed to initialize server state",
			zap.Error(err))
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	s.logger.Info("server configuration loaded",
		zap.Int("server_count", newState.GetServerCount()),
		zap.Int("tool_count", newState.GetToolCount()),
		zap.Int("router_count", newState.GetRouterCount()))

	// Atomically replace the state
	s.state = newState

	// Register all routes under root path
	s.logger.Debug("registering root handler")
	s.router.NoRoute(s.handleRoot)

	return nil
}

// handleRoot handles all requests and routes them based on the path
func (s *Server) handleRoot(c *gin.Context) {
	path := c.Request.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		s.logger.Debug("invalid path format",
			zap.String("path", path),
			zap.String("remote_addr", c.Request.RemoteAddr))
		s.sendProtocolError(c, nil, "Invalid path", http.StatusBadRequest, mcp.ErrorCodeInvalidRequest)
		return
	}
	endpoint := parts[len(parts)-1]
	prefix := "/" + strings.Join(parts[:len(parts)-1], "/")

	s.logger.Debug("routing request",
		zap.String("path", path),
		zap.String("prefix", prefix),
		zap.String("endpoint", endpoint),
		zap.String("remote_addr", c.Request.RemoteAddr))

	// Dynamically set CORS
	if cors := s.state.GetCORS(prefix); cors != nil {
		s.logger.Debug("applying CORS middleware",
			zap.String("prefix", prefix))
		s.corsMiddleware(cors)(c)
		if c.IsAborted() {
			s.logger.Debug("request aborted by CORS middleware",
				zap.String("prefix", prefix),
				zap.String("remote_addr", c.Request.RemoteAddr))
			return
		}
	}

	protoType := s.state.GetProtoType(prefix)
	if protoType == "" {
		s.logger.Warn("invalid prefix",
			zap.String("prefix", prefix),
			zap.String("remote_addr", c.Request.RemoteAddr))
		s.sendProtocolError(c, nil, "Invalid prefix", http.StatusNotFound, mcp.ErrorCodeInvalidRequest)
		return
	}

	c.Status(http.StatusOK)
	switch endpoint {
	case "sse":
		s.logger.Debug("handling SSE endpoint",
			zap.String("prefix", prefix))
		s.handleSSE(c)
	case "message":
		s.logger.Debug("handling message endpoint",
			zap.String("prefix", prefix))
		s.handleMessage(c)
	case "mcp":
		s.logger.Debug("handling MCP endpoint",
			zap.String("prefix", prefix))
		s.handleMCP(c)
	default:
		s.logger.Warn("invalid endpoint",
			zap.String("endpoint", endpoint),
			zap.String("prefix", prefix),
			zap.String("remote_addr", c.Request.RemoteAddr))
		s.sendProtocolError(c, nil, "Invalid endpoint", http.StatusNotFound, mcp.ErrorCodeInvalidRequest)
	}
}

func (s *Server) Start() {
	go func() {
		if err := s.router.Run(fmt.Sprintf(":%d", s.cfg.Port)); err != nil {
			s.logger.Error("failed to start server", zap.Error(err))
		}
	}()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down server")
	close(s.shutdownCh)

	var wg sync.WaitGroup
	for prefix, transport := range s.state.GetTransports() {
		if transport.IsRunning() {
			wg.Add(1)
			go func(p string, t mcpproxy.Transport) {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := t.Stop(ctx); err != nil {
					if err.Error() == "signal: interrupt" {
						s.logger.Info("transport stopped", zap.String("prefix", p))
						return
					}
					s.logger.Error("failed to stop transport",
						zap.String("prefix", p),
						zap.Error(err))
				}
			}(prefix, transport)
		}
	}
	wg.Wait()

	return nil
}

// UpdateConfig updates the server configuration
func (s *Server) UpdateConfig(ctx context.Context, cfgs []*config.MCPConfig) error {
	// Validate configuration before updating
	s.logger.Debug("validating updated configuration")
	if err := config.ValidateMCPConfigs(cfgs); err != nil {
		s.logger.Error("invalid configuration during update",
			zap.Error(err))
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create new state and load configuration
	s.logger.Info("updating server configuration")
	newState, err := BuildStateFromConfig(ctx, cfgs, s.state, s.logger)
	if err != nil {
		s.logger.Error("failed to initialize state during update",
			zap.Error(err))
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// 记录配置更新信息
	s.logger.Info("server configuration updated",
		zap.Int("server_count", newState.GetServerCount()),
		zap.Int("tool_count", newState.GetToolCount()),
		zap.Int("router_count", newState.GetRouterCount()))

	// Atomically replace the state
	s.state = newState

	return nil
}

// MergeConfig updates the server configuration incrementally
func (s *Server) MergeConfig(ctx context.Context, cfg *config.MCPConfig) error {
	s.logger.Info("merging configuration")

	newConfig, err := helper.MergeConfigs(s.state.GetRawConfigs(), cfg)
	if err != nil {
		s.logger.Error("failed to merge configuration",
			zap.Error(err))
		return fmt.Errorf("failed to merge configuration: %w", err)
	}

	// Validate configuration after merge
	s.logger.Debug("validating merged configuration")
	if err := config.ValidateMCPConfig(cfg); err != nil {
		s.logger.Error("invalid configuration after merge",
			zap.Error(err))
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create new state and load configuration
	s.logger.Debug("initializing state with merged configuration")
	newState, err := BuildStateFromConfig(ctx, newConfig, s.state, s.logger)
	if err != nil {
		s.logger.Error("failed to initialize state with merged configuration",
			zap.Error(err))
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Record configuration merge information
	s.logger.Info("configuration merged successfully",
		zap.Int("server_count", newState.GetServerCount()),
		zap.Int("tool_count", newState.GetToolCount()),
		zap.Int("router_count", newState.GetRouterCount()))

	// Atomically replace the state
	s.state = newState

	return nil
}
