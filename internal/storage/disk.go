package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

// DiskStorage implements Storage interface using local disk
type DiskStorage struct {
	logger  *zap.Logger
	baseDir string
}

// NewDiskStorage creates a new disk storage
func NewDiskStorage(logger *zap.Logger, baseDir string) (*DiskStorage, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}

	return &DiskStorage{
		logger:  logger,
		baseDir: baseDir,
	}, nil
}

// Save saves a configuration to disk
func (s *DiskStorage) Save(ctx context.Context, name string, content io.Reader) error {
	// Create file path
	filePath := filepath.Join(s.baseDir, name)

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy content to file
	if _, err := io.Copy(file, content); err != nil {
		return err
	}

	return nil
}

// Load loads a configuration from disk
func (s *DiskStorage) Load(ctx context.Context, name string) (io.ReadCloser, error) {
	// Create file path
	filePath := filepath.Join(s.baseDir, name)

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// List lists all configurations in disk
func (s *DiskStorage) List(ctx context.Context) ([]string, error) {
	// Read directory
	files, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, err
	}

	// Get file names
	names := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			names = append(names, file.Name())
		}
	}

	return names, nil
}

// Delete deletes a configuration from disk
func (s *DiskStorage) Delete(ctx context.Context, name string) error {
	// Create file path
	filePath := filepath.Join(s.baseDir, name)

	// Delete file
	return os.Remove(filePath)
}
