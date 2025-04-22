package config

import "fmt"

type (
	APIServerConfig struct {
		Database   DatabaseConfig `yaml:"database"`
		OpenAI     OpenAIConfig   `yaml:"openai"`
		GatewayPID string         `yaml:"gateway_pid"`
	}

	DatabaseConfig struct {
		Type     string `yaml:"type"`     // postgres, sqlite, etc.
		Host     string `yaml:"host"`     // localhost (for postgres)
		Port     int    `yaml:"port"`     // 5432 (for postgres)
		User     string `yaml:"user"`     // postgres
		Password string `yaml:"password"` // postgres
		DBName   string `yaml:"dbname"`   // mcp_gateway (for postgres) or file path (for sqlite)
		SSLMode  string `yaml:"sslmode"`  // disable (for postgres)
	}

	OpenAIConfig struct {
		APIKey string `yaml:"api_key" env:"OPENAI_API_KEY"`
		Model  string `yaml:"model" env:"OPENAI_MODEL" env-default:"gpt-3.5-turbo"`
	}
)

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	switch c.Type {
	case "postgres":
		return c.getPostgresDSN()
	case "sqlite":
		return c.DBName // For SQLite, DBName is the file path
	default:
		return ""
	}
}

// getPostgresDSN returns PostgreSQL connection string
func (c *DatabaseConfig) getPostgresDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}
