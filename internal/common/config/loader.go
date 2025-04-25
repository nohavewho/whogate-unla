package config

import (
	"github.com/mcp-ecosystem/mcp-gateway/internal/common/cnst"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// Loader is responsible for loading configuration
type Loader struct {
	logger *zap.Logger
}

// NewLoader creates a new configuration loader
func NewLoader(logger *zap.Logger) *Loader {
	return &Loader{
		logger: logger,
	}
}

// LoadFromFile loads configuration from a YAML file
func (l *Loader) LoadFromFile(path string) (*MCPConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg MCPConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if err := l.Validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LoadFromDir loads MCP server configurations from a directory and merges them
func (l *Loader) LoadFromDir(dir string) (*MCPConfig, error) {
	// Create a base config
	baseCfg := &MCPConfig{
		Routers: make([]RouterConfig, 0),
		Servers: make([]ServerConfig, 0),
		Tools:   make([]ToolConfig, 0),
	}

	// Walk through the directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-YAML files
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".yaml") {
			return nil
		}

		// Load MCP server configuration from file
		cfg, err := l.LoadFromFile(path)
		if err != nil {
			l.logger.Error("failed to load MCP server configuration file",
				zap.String("path", path),
				zap.Error(err))
			return nil // Continue with other files
		}

		// Merge MCP server configurations
		if err := l.mergeConfig(baseCfg, cfg); err != nil {
			l.logger.Error("failed to merge MCP server configuration",
				zap.String("path", path),
				zap.Error(err))
			return nil // Continue with other files
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Validate the merged MCP server configuration
	if err := l.Validate(baseCfg); err != nil {
		return nil, err
	}

	return baseCfg, nil
}

// mergeConfig merges two configurations
func (l *Loader) mergeConfig(base, override *MCPConfig) error {
	// Merge routers
	routerMap := make(map[string]RouterConfig)
	for _, router := range base.Routers {
		routerMap[router.Server] = router
	}
	for _, router := range override.Routers {
		routerMap[router.Server] = router
	}
	base.Routers = make([]RouterConfig, 0, len(routerMap))
	for _, router := range routerMap {
		base.Routers = append(base.Routers, router)
	}

	// Merge servers
	serverMap := make(map[string]ServerConfig)
	for _, server := range base.Servers {
		serverMap[server.Name] = server
	}
	for _, server := range override.Servers {
		serverMap[server.Name] = server
	}
	base.Servers = make([]ServerConfig, 0, len(serverMap))
	for _, server := range serverMap {
		base.Servers = append(base.Servers, server)
	}

	// Merge tools
	toolMap := make(map[string]ToolConfig)
	for _, tool := range base.Tools {
		toolMap[tool.Name] = tool
	}
	for _, tool := range override.Tools {
		toolMap[tool.Name] = tool
	}
	base.Tools = make([]ToolConfig, 0, len(toolMap))
	for _, tool := range toolMap {
		base.Tools = append(base.Tools, tool)
	}

	return nil
}

// Validate performs configuration validation
func (l *Loader) Validate(cfg *MCPConfig) error {
	// Validate tool names are unique
	toolNames := make(map[string]bool)
	for _, tool := range cfg.Tools {
		if toolNames[tool.Name] {
			return cnst.ErrDuplicateToolName
		}
		toolNames[tool.Name] = true
	}

	// Validate server names are unique
	serverNames := make(map[string]bool)
	for _, server := range cfg.Servers {
		if serverNames[server.Name] {
			return cnst.ErrDuplicateServerName
		}
		serverNames[server.Name] = true
	}

	// Validate router prefixes don't conflict
	prefixes := make(map[string]bool)
	for _, router := range cfg.Routers {
		if prefixes[router.Prefix] {
			return cnst.ErrDuplicateRouterPrefix
		}
		prefixes[router.Prefix] = true
	}

	return nil
}
