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
		tools                   []mcp.ToolSchema
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
		tools:                   make([]mcp.ToolSchema, 0),
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
			newState.tools = append(newState.tools, tool.ToToolSchema())
		}

		// Build prefix to tools mapping for MCP servers
		prefixMap := make(map[string]string)
		for i, routerCfg := range cfg.Routers {
			prefixMap[routerCfg.Server] = routerCfg.Prefix
			newState.prefixToRouterConfig[routerCfg.Prefix] = &cfg.Routers[i]
		}

		// Process regular HTTP servers
		for _, serverCfg := range cfg.Servers {
			prefix, exists := prefixMap[serverCfg.Name]
			if !exists {
				return nil, fmt.Errorf("no router prefix found for MCP server: %s", serverCfg.Name)
			}

			// Filter tools based on MCP server's allowed tools
			var allowedTools []mcp.ToolSchema
			for _, toolName := range serverCfg.AllowedTools {
				if tool, ok := newState.toolMap[toolName]; ok {
					allowedTools = append(allowedTools, tool.ToToolSchema())
				}
			}
			newState.prefixToTools[prefix] = allowedTools
			newState.prefixToServerConfig[prefix] = &serverCfg
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
				go func(prefix string, mcpServer config.MCPServerConfig, transport mcpproxy.Transport) {
					if transport.IsRunning() {
						logger.Info("server already started",
							zap.String("prefix", prefix),
							zap.String("command", mcpServer.Command),
							zap.Strings("args", mcpServer.Args))
						return
					}

					if err := transport.Start(ctx, template.NewContext()); err != nil {
						logger.Error("failed to start server",
							zap.String("prefix", prefix),
							zap.String("command", mcpServer.Command),
							zap.Strings("args", mcpServer.Args),
							zap.Error(err))
					} else {
						logger.Info("server started",
							zap.String("prefix", prefix),
							zap.String("command", mcpServer.Command),
							zap.Strings("args", mcpServer.Args))
					}
				}(prefix, mcpServer, transport)
			} else if mcpServer.Preinstalled {
				// If Preinstalled is set but not PolicyOnStart, verify installation by starting and stopping
				go func(prefix string, mcpServer config.MCPServerConfig, transport mcpproxy.Transport) {
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
				}(prefix, mcpServer, transport)
			}
			newState.prefixToTransport[prefix] = transport

			// Map protocol type based on server type
			switch mcpServer.Type {
			case "stdio":
				newState.prefixToProtoType[prefix] = cnst.BackendProtoStdio
			case "sse":
				newState.prefixToProtoType[prefix] = cnst.BackendProtoSSE
			case "streamable-http":
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
