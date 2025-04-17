package config

import (
	"github.com/mcp-ecosystem/mcp-gateway/pkg/mcp"
)

// Config represents the root configuration structure
type Config struct {
	Routers []RouterConfig `yaml:"routers"`
	Servers []ServerConfig `yaml:"servers"`
	Tools   []ToolConfig   `yaml:"tools"`
}

// GlobalConfig represents the global configuration
type GlobalConfig struct {
	Namespace string `yaml:"namespace"`
	Prefix    string `yaml:"prefix"`
}

// RouterConfig represents the router configuration
type RouterConfig struct {
	Server string `yaml:"server"`
	Prefix string `yaml:"prefix"`
}

// ServerConfig represents the server configuration
type ServerConfig struct {
	Name           string     `yaml:"name"`
	Namespace      string     `yaml:"namespace"`
	Description    string     `yaml:"description"`
	Auth           AuthConfig `yaml:"auth"`
	AllowedTools   []string   `yaml:"allowedTools"`
	AllowedOrigins []string   `yaml:"allowedOrigins"`
}

// AuthConfig represents the authentication configuration
type AuthConfig struct {
	Mode   string `yaml:"mode"`   // bearer / apikey / none
	Header string `yaml:"header"` // header name for auth
	ArgKey string `yaml:"argKey"` // parameter key for auth
}

// ToolConfig represents a tool configuration
type ToolConfig struct {
	Name         string            `yaml:"name"`
	Description  string            `yaml:"description,omitempty"`
	Method       string            `yaml:"method"`
	Endpoint     string            `yaml:"endpoint"`
	Headers      map[string]string `yaml:"headers"`
	Args         []ArgConfig       `yaml:"args"`
	RequestBody  string            `yaml:"requestBody"`
	ResponseBody string            `yaml:"responseBody"`
	InputSchema  map[string]any    `yaml:"inputSchema,omitempty"`
}

// ArgConfig represents the argument configuration
type ArgConfig struct {
	Name        string `yaml:"name"`
	Position    string `yaml:"position"` // header, query, path, body
	Required    bool   `yaml:"required"`
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
	Default     string `yaml:"default"`
}

// ToToolSchema converts a ToolConfig to a ToolSchema
func (t *ToolConfig) ToToolSchema() mcp.ToolSchema {
	// Create properties map for input schema
	properties := make(map[string]any)
	required := make([]string, 0)
	for _, arg := range t.Args {
		property := map[string]any{
			"type":        arg.Type,
			"description": arg.Description,
			"required":    arg.Required,
		}
		if arg.Description != "" {
			property["title"] = arg.Description
		}
		properties[arg.Name] = property
		if arg.Required {
			required = append(required, arg.Name)
		}
	}

	// Merge with existing input schema if any
	if t.InputSchema != nil {
		for k, v := range t.InputSchema {
			properties[k] = v
		}
	}

	return mcp.ToolSchema{
		Name:        t.Name,
		Description: t.Description,
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: properties,
			Required:   required,
		},
	}
}
