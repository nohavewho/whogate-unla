package storage

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mcp-ecosystem/mcp-gateway/internal/common/config"
	"github.com/mcp-ecosystem/mcp-gateway/internal/mcp/storage/notifier"
	"gopkg.in/yaml.v3"

	"go.uber.org/zap"
)

// DiskStore implements the Store interface using the local filesystem
type DiskStore struct {
	logger   *zap.Logger
	baseDir  string
	mu       sync.RWMutex
	notifier notifier.Notifier
}

var _ Store = (*DiskStore)(nil)

// NewDiskStore creates a new disk-based store
func NewDiskStore(ctx context.Context, logger *zap.Logger, baseDir string) (*DiskStore, error) {
	logger = logger.Named("mcp.store")

	baseDir = getConfigPath(baseDir)
	logger.Info("Using configuration directory", zap.String("path", baseDir))

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}

	return &DiskStore{
		logger:   logger,
		baseDir:  baseDir,
		notifier: notifier.NewSignalNotifier(ctx, logger),
	}, nil
}

// Create implements Store.Create
func (s *DiskStore) Create(_ context.Context, server *config.MCPConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	server.UpdatedAt = time.Now()
	if server.CreatedAt.IsZero() {
		server.CreatedAt = server.UpdatedAt
	}

	data, err := yaml.Marshal(server)
	if err != nil {
		return err
	}

	path := filepath.Join(s.baseDir, server.Name+".yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Get implements Store.Get
func (s *DiskStore) Get(_ context.Context, name string) (*config.MCPConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.baseDir, name+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var server config.MCPConfig
	if err := yaml.Unmarshal(data, &server); err != nil {
		return nil, err
	}

	return &server, nil
}

// List implements Store.List
func (s *DiskStore) List(_ context.Context) ([]*config.MCPConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var servers []*config.MCPConfig
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		data, err := os.ReadFile(filepath.Join(s.baseDir, entry.Name()))
		if err != nil {
			s.logger.Error("failed to read server file",
				zap.String("file", entry.Name()),
				zap.Error(err))
			continue
		}

		var server config.MCPConfig
		if err := yaml.Unmarshal(data, &server); err != nil {
			s.logger.Error("failed to unmarshal server",
				zap.String("file", entry.Name()),
				zap.Error(err))
			continue
		}

		servers = append(servers, &server)
	}

	return servers, nil
}

// Update implements Store.Update
func (s *DiskStore) Update(_ context.Context, server *config.MCPConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	server.UpdatedAt = time.Now()
	data, err := yaml.Marshal(server)
	if err != nil {
		return err
	}

	path := filepath.Join(s.baseDir, server.Name+".yaml")
	return os.WriteFile(path, data, 0644)
}

// Delete implements Store.Delete
func (s *DiskStore) Delete(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.baseDir, name+".yaml")
	return os.Remove(path)
}

// GetNotifier implements Store.GetNotifier
func (s *DiskStore) GetNotifier(_ context.Context) notifier.Notifier {
	return s.notifier
}

func getConfigPath(baseDir string) string {
	// 1. Check command line flag
	if baseDir != "" {
		return baseDir
	}

	// 2. Check environment variable
	if envPath := os.Getenv("CONFIG_DIR"); envPath != "" {
		return envPath
	}

	// 3. Default to APPDATA/.mcp/gateway
	appData := os.Getenv("APPDATA")
	if appData == "" {
		// For non-Windows systems, use HOME
		appData = os.Getenv("HOME")
		if appData == "" {
			log.Fatal("Neither APPDATA nor HOME environment variable is set")
		}
	}
	return filepath.Join(appData, ".mcp", "gateway")
}
