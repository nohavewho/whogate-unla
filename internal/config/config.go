package config

import (
	"os"
	"regexp"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type MCPGatewayConfig struct {
	Port       int    `yaml:"port"`
	InnerPort  int    `yaml:"inner_port"`
	ReloadPort int    `yaml:"reload_port"`
	PID        string `yaml:"pid"`
}

type Type interface {
	MCPGatewayConfig | APIServerConfig
}

// LoadConfig loads configuration from a YAML file with environment variable support
func LoadConfig[T Type](path string) (*T, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Resolve environment variables
	data = resolveEnv(data)

	var cfg T
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
