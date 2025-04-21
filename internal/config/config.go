package config

import (
	"fmt"
	"os"
	"regexp"

	"github.com/joho/godotenv"
	"github.com/mcp-ecosystem/mcp-gateway/pkg/mcp"
	"gopkg.in/yaml.v3"
)

// Config represents the root configuration structure
type Config struct {
	Routers  []RouterConfig `yaml:"routers"`
	Servers  []ServerConfig `yaml:"servers"`
	Tools    []ToolConfig   `yaml:"tools"`
	Database DatabaseConfig `yaml:"database"`
	OpenAI   struct {
		APIKey string `yaml:"api_key" env:"OPENAI_API_KEY"`
		Model  string `yaml:"model" env:"OPENAI_MODEL" env-default:"gpt-3.5-turbo"`
	} `yaml:"openai"`
}

// GlobalConfig represents the global configuration
type GlobalConfig struct {
	Namespace string `yaml:"namespace"`
	Prefix    string `yaml:"prefix"`
}

// RouterConfig represents the router configuration
type RouterConfig struct {
	Server string      `yaml:"server"`
	Prefix string      `yaml:"prefix"`
	CORS   *CORSConfig `yaml:"cors,omitempty"`
}

// ServerConfig represents the server configuration
type ServerConfig struct {
	Name         string   `yaml:"name"`
	Namespace    string   `yaml:"namespace"`
	Description  string   `yaml:"description"`
	AllowedTools []string `yaml:"allowedTools"`
}

// AuthConfig represents the authentication configuration
type AuthConfig struct {
	Mode   string `yaml:"mode"`   // bearer / apikey / none
	Header string `yaml:"header"` // header name for auth
	ArgKey string `yaml:"argKey"` // parameter key for auth
}

// ToolConfig represents the tool configuration
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

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowOrigins     []string `yaml:"allowOrigins"`
	AllowMethods     []string `yaml:"allowMethods"`
	AllowHeaders     []string `yaml:"allowHeaders"`
	ExposeHeaders    []string `yaml:"exposeHeaders"`
	AllowCredentials bool     `yaml:"allowCredentials"`
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

// Server represents a single server configuration
type Server struct {
	Name string `yaml:"name"`
	// Add other server fields as needed
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Type     string `yaml:"type"`     // postgres, mysql, etc.
	Host     string `yaml:"host"`     // localhost
	Port     int    `yaml:"port"`     // 5432
	User     string `yaml:"user"`     // postgres
	Password string `yaml:"password"` // postgres
	DBName   string `yaml:"dbname"`   // mcp_gateway
	SSLMode  string `yaml:"sslmode"`  // disable
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	switch c.Type {
	case "postgres":
		return c.getPostgresDSN()
	default:
		return ""
	}
}

// getPostgresDSN returns PostgreSQL connection string
func (c *DatabaseConfig) getPostgresDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}

// LoadConfig loads configuration from a YAML file with environment variable support
func LoadConfig(path string) (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Resolve environment variables
	data = resolveEnv(data)

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// resolveEnv replaces environment variable placeholders in YAML content
func resolveEnv(content []byte) []byte {
	regex := regexp.MustCompile(`\$\{(\w+)(?::([^}]+))?}`)

	return regex.ReplaceAllFunc(content, func(match []byte) []byte {
		matches := regex.FindSubmatch(match)
		envKey := string(matches[1])
		var defaultValue string

		if len(matches) > 2 {
			defaultValue = string(matches[2])
		}

		if value, exists := os.LookupEnv(envKey); exists {
			return []byte(value)
		}
		return []byte(defaultValue)
	})
}
