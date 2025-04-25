package helper

import "github.com/mcp-ecosystem/mcp-gateway/internal/common/config"

// MergeConfigs merges all configurations
func MergeConfigs(configs []*config.MCPConfig) (*config.MCPConfig, error) {
	mergedConfig := &config.MCPConfig{}

	for _, cfg := range configs {
		if err := mergeConfig(mergedConfig, cfg); err != nil {
			return nil, err
		}
	}

	return mergedConfig, nil
}

// mergeConfig merges two configurations
func mergeConfig(base, override *config.MCPConfig) error {
	// Merge routers
	routerMap := make(map[string]config.RouterConfig)
	for _, router := range base.Routers {
		routerMap[router.Server] = router
	}
	for _, router := range override.Routers {
		routerMap[router.Server] = router
	}
	base.Routers = make([]config.RouterConfig, 0, len(routerMap))
	for _, router := range routerMap {
		base.Routers = append(base.Routers, router)
	}

	// Merge servers
	serverMap := make(map[string]config.ServerConfig)
	for _, server := range base.Servers {
		serverMap[server.Name] = server
	}
	for _, server := range override.Servers {
		serverMap[server.Name] = server
	}
	base.Servers = make([]config.ServerConfig, 0, len(serverMap))
	for _, server := range serverMap {
		base.Servers = append(base.Servers, server)
	}

	// Merge tools
	toolMap := make(map[string]config.ToolConfig)
	for _, tool := range base.Tools {
		toolMap[tool.Name] = tool
	}
	for _, tool := range override.Tools {
		toolMap[tool.Name] = tool
	}
	base.Tools = make([]config.ToolConfig, 0, len(toolMap))
	for _, tool := range toolMap {
		base.Tools = append(base.Tools, tool)
	}

	return nil
}
