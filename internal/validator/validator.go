package validator

import (
	"io"

	"github.com/mcp-ecosystem/mcp-gateway/internal/config"
	"gopkg.in/yaml.v3"
)

// Validator validates YAML configurations
type Validator struct {
	loader *config.Loader
}

// NewValidator creates a new validator
func NewValidator(loader *config.Loader) *Validator {
	return &Validator{
		loader: loader,
	}
}

// Validate validates a YAML configuration
func (v *Validator) Validate(content io.Reader) error {
	// Read content
	data, err := io.ReadAll(content)
	if err != nil {
		return err
	}

	// Unmarshal YAML
	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}

	// Validate configuration
	if err := v.loader.Validate(&cfg); err != nil {
		return err
	}

	return nil
}
