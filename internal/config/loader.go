package config

import (
	"os"

	"github.com/mcp-ecosystem/mcp-gateway/pkg/errors"
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
func (l *Loader) LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if err := l.validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// validate performs configuration validation
func (l *Loader) validate(cfg *Config) error {
	// Validate tool names are unique
	toolNames := make(map[string]bool)
	for _, tool := range cfg.Tools {
		if toolNames[tool.Name] {
			return errors.ErrDuplicateToolName(tool.Name)
		}
		toolNames[tool.Name] = true
	}

	// Validate server names are unique
	serverNames := make(map[string]bool)
	for _, server := range cfg.Servers {
		if serverNames[server.Name] {
			return errors.ErrDuplicateServerName(server.Name)
		}
		serverNames[server.Name] = true
	}

	// Validate router prefixes don't conflict
	prefixes := make(map[string]bool)
	for _, router := range cfg.Routers {
		if prefixes[router.Prefix] {
			return errors.ErrDuplicateRouterPrefix(router.Prefix)
		}
		prefixes[router.Prefix] = true
	}

	return nil
}
