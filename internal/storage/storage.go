package storage

import (
	"context"
	"io"
)

// Storage defines the interface for configuration storage
type Storage interface {
	// Save saves a configuration to storage
	Save(ctx context.Context, name string, content io.Reader) error

	// Load loads a configuration from storage
	Load(ctx context.Context, name string) (io.ReadCloser, error)

	// List lists all configurations in storage
	List(ctx context.Context) ([]string, error)

	// Delete deletes a configuration from storage
	Delete(ctx context.Context, name string) error
}
