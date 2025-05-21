package core

import (
	"context"
	"fmt"

	"github.com/mcp-ecosystem/mcp-gateway/internal/common/cnst"
	"github.com/mcp-ecosystem/mcp-gateway/internal/common/config"
	"github.com/mcp-ecosystem/mcp-gateway/internal/core/mcpproxy"
	"github.com/mcp-ecosystem/mcp-gateway/internal/template"
	"github.com/mcp-ecosystem/mcp-gateway/pkg/mcp"
	"go.uber.org/zap"
)

type (
	// State contains all the read-only shared state
	State struct {
		rawConfigs              []*config.MCPConfig
		toolMap                 map[string]*config.ToolConfig
		prefixToTools           map[string][]mcp.ToolSchema
		prefixToServerConfig    map[string]*config.ServerConfig
		prefixToRouterConfig    map[string]*config.RouterConfig
		prefixToMCPServerConfig map[string]config.MCPServerConfig
		prefixToProtoType       map[string]cnst.ProtoType
		prefixToTransport       map[string]mcpproxy.Transport
	}
)

func NewState() *State {
	return &State{
		rawConfigs:              make([]*config.MCPConfig, 0),
		toolMap:                 make(map[string]*config.ToolConfig),
		prefixToTools:           make(map[string][]mcp.ToolSchema),
		prefixToServerConfig:    make(map[string]*config.ServerConfig),
		prefixToRouterConfig:    make(map[string]*config.RouterConfig),
		prefixToMCPServerConfig: make(map[string]config.MCPServerConfig),
		prefixToProtoType:       make(map[string]cnst.ProtoType),
		prefixToTransport:       make(map[string]mcpproxy.Transport),
	}
}

// BuildStateFromConfig creates a new State from the given configuration
func BuildStateFromConfig(ctx context.Context, cfgs []*config.MCPConfig, oldState *State, logger *zap.Logger) (*State, error) {
	// Create new state
	newState := NewState()
	newState.rawConfigs = cfgs

	for _, cfg := range cfgs {
		// Initialize tool map and list for MCP servers
		for _, tool := range cfg.Tools {
			newState.toolMap[tool.Name] = &tool
		}

		// Build prefix to tools mapping for MCP servers
		prefixMap := make(map[string]string)
		for _, router := range cfg.Routers {
			prefixMap[router.Server] = router.Prefix
			newState.prefixToRouterConfig[router.Prefix] = &router
		}

		// Process regular HTTP servers
		for _, server := range cfg.Servers {
			prefix, ok := prefixMap[server.Name]
			if !ok {
				logger.Warn("failed to find prefix for server", zap.String("server", server.Name))
				continue
			}

			// Filter tools based on MCP server's allowed tools
			var allowedTools []mcp.ToolSchema
			for _, toolName := range server.AllowedTools {
				tool, ok := newState.toolMap[toolName]
				if ok {
					allowedTools = append(allowedTools, tool.ToToolSchema())
				} else {
					logger.Warn("failed to find allowed tool for server", zap.String("server", server.Name),
						zap.String("tool", toolName))
				}
			}
			newState.prefixToTools[prefix] = allowedTools
			newState.prefixToServerConfig[prefix] = &server
			newState.prefixToProtoType[prefix] = cnst.BackendProtoHttp
		}

		// Process MCP servers
		for _, mcpServer := range cfg.McpServers {
			prefix, exists := prefixMap[mcpServer.Name]
			if !exists {
				continue // Skip MCP servers without router prefix
			}

			// Map prefix to MCP server config
			newState.prefixToMCPServerConfig[prefix] = mcpServer

			// Check if we already have transport with the same configuration
			var transport mcpproxy.Transport
			if oldState != nil {
				if oldTransport, exists := oldState.prefixToTransport[prefix]; exists {
					// Compare configurations to see if we need to create a new transport
					oldConfig := oldState.prefixToMCPServerConfig[prefix]
					if oldConfig.Type == mcpServer.Type &&
						oldConfig.Command == mcpServer.Command &&
						oldConfig.URL == mcpServer.URL &&
						len(oldConfig.Args) == len(mcpServer.Args) {
						// Compare args
						argsMatch := true
						for i, arg := range oldConfig.Args {
							if arg != mcpServer.Args[i] {
								argsMatch = false
								break
							}
						}
						if argsMatch {
							// Reuse existing transport
							transport = oldTransport
						}
					}
				}
			}

			// Create new transport if needed
			if transport == nil {
				var err error
				transport, err = mcpproxy.NewTransport(mcpServer)
				if err != nil {
					return nil, fmt.Errorf("failed to create transport for server %s: %w", mcpServer.Name, err)
				}
			}

			// Handle server startup based on policy and preinstalled flag
			if mcpServer.Policy == cnst.PolicyOnStart {
				// If PolicyOnStart is set, just start the server and keep it running
				go startMCPServer(ctx, logger, prefix, mcpServer, transport, false)
			} else if mcpServer.Preinstalled {
				// If Preinstalled is set but not PolicyOnStart, verify installation by starting and stopping
				go startMCPServer(ctx, logger, prefix, mcpServer, transport, true)
			}
			newState.prefixToTransport[prefix] = transport

			// Map protocol type based on server type
			switch mcpServer.Type {
			case cnst.BackendProtoStdio.String():
				newState.prefixToProtoType[prefix] = cnst.BackendProtoStdio
			case cnst.BackendProtoSSE.String():
				newState.prefixToProtoType[prefix] = cnst.BackendProtoSSE
			case cnst.BackendProtoStreamable.String():
				newState.prefixToProtoType[prefix] = cnst.BackendProtoStreamable
			}
		}
	}

	if oldState != nil {
		for prefix, oldTransport := range oldState.prefixToTransport {
			if _, stillExists := newState.prefixToTransport[prefix]; !stillExists {
				mcpSvrCfg := oldState.prefixToMCPServerConfig[prefix]
				if oldTransport == nil {
					logger.Info("transport already stopped", zap.String("prefix", prefix),
						zap.String("command", mcpSvrCfg.Command), zap.Strings("args", mcpSvrCfg.Args))
					continue
				}
				logger.Info("shutting down unused transport", zap.String("prefix", prefix),
					zap.String("command", mcpSvrCfg.Command), zap.Strings("args", mcpSvrCfg.Args))
				if err := oldTransport.Stop(ctx); err != nil {
					logger.Warn("failed to close old transport", zap.String("prefix", prefix),
						zap.Error(err), zap.String("command", mcpSvrCfg.Command),
						zap.Strings("args", mcpSvrCfg.Args))
				}
			}
		}
	}

	return newState, nil
}

func startMCPServer(ctx context.Context, logger *zap.Logger, prefix string, mcpServer config.MCPServerConfig,
	transport mcpproxy.Transport, needStop bool) {
	// If Preinstalled is set but not PolicyOnStart, verify installation by starting and stopping
	if transport.IsRunning() {
		logger.Info("server already started, don't need to preinstall",
			zap.String("prefix", prefix),
			zap.String("command", mcpServer.Command),
			zap.Strings("args", mcpServer.Args))
		return
	}

	if err := transport.Start(ctx, template.NewContext()); err != nil {
		logger.Error("failed to start server for preinstall",
			zap.String("prefix", prefix),
			zap.String("command", mcpServer.Command),
			zap.Strings("args", mcpServer.Args),
			zap.Error(err))
	} else {
		logger.Info("server started for preinstall",
			zap.String("prefix", prefix),
			zap.String("command", mcpServer.Command),
			zap.Strings("args", mcpServer.Args))

		if needStop {
			// Stop the server after successful start
			if err := transport.Stop(ctx); err != nil {
				logger.Error("failed to stop server for preinstall",
					zap.String("prefix", prefix),
					zap.String("command", mcpServer.Command),
					zap.Strings("args", mcpServer.Args),
					zap.Error(err))
			} else {
				logger.Info("server stopped for preinstall",
					zap.String("prefix", prefix),
					zap.String("command", mcpServer.Command),
					zap.Strings("args", mcpServer.Args))
			}
		}
	}
}

func (s *State) GetCORS(prefix string) *config.CORSConfig {
	routerCfg, ok := s.prefixToRouterConfig[prefix]
	if ok {
		return routerCfg.CORS
	}
	return nil
}

func (s *State) GetRouterCount() int {
	return len(s.prefixToRouterConfig)
}

func (s *State) GetToolCount() int {
	return len(s.toolMap)
}

func (s *State) GetServerCount() int {
	return len(s.prefixToServerConfig)
}

func (s *State) GetTool(name string) *config.ToolConfig {
	return s.toolMap[name]
}

func (s *State) GetToolSchemas(prefix string) []mcp.ToolSchema {
	return s.prefixToTools[prefix]
}

func (s *State) GetServerConfig(prefix string) *config.ServerConfig {
	return s.prefixToServerConfig[prefix]
}

func (s *State) GetProtoType(prefix string) cnst.ProtoType {
	return s.prefixToProtoType[prefix]
}

func (s *State) GetTransport(prefix string) mcpproxy.Transport {
	return s.prefixToTransport[prefix]
}

func (s *State) GetTransports() map[string]mcpproxy.Transport {
	return s.prefixToTransport
}

func (s *State) GetRawConfigs() []*config.MCPConfig {
	return s.rawConfigs
}
