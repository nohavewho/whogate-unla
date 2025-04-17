package storage

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mcp-ecosystem/mcp-gateway/internal/server"
	"go.uber.org/zap"
)

// DiskStorage implements the Storage interface using the local filesystem
type DiskStorage struct {
	logger  *zap.Logger
	baseDir string
	mu      sync.RWMutex
}

// NewDiskStorage creates a new disk-based storage
func NewDiskStorage(logger *zap.Logger, baseDir string) (*DiskStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}

	return &DiskStorage{
		logger:  logger,
		baseDir: baseDir,
	}, nil
}

// SaveTool implements Storage.SaveTool
func (s *DiskStorage) SaveTool(ctx context.Context, tool *server.Tool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tool.UpdatedAt = time.Now()
	if tool.CreatedAt.IsZero() {
		tool.CreatedAt = tool.UpdatedAt
	}

	data, err := json.Marshal(tool)
	if err != nil {
		return err
	}

	path := filepath.Join(s.baseDir, "tools", tool.Name+".json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetTool implements Storage.GetTool
func (s *DiskStorage) GetTool(ctx context.Context, name string) (*server.Tool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.baseDir, "tools", name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var tool server.Tool
	if err := json.Unmarshal(data, &tool); err != nil {
		return nil, err
	}

	return &tool, nil
}

// ListTools implements Storage.ListTools
func (s *DiskStorage) ListTools(ctx context.Context) ([]*server.Tool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := filepath.Join(s.baseDir, "tools")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var tools []*server.Tool
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			s.logger.Error("failed to read tool file",
				zap.String("file", entry.Name()),
				zap.Error(err))
			continue
		}

		var tool server.Tool
		if err := json.Unmarshal(data, &tool); err != nil {
			s.logger.Error("failed to unmarshal tool",
				zap.String("file", entry.Name()),
				zap.Error(err))
			continue
		}

		tools = append(tools, &tool)
	}

	return tools, nil
}

// DeleteTool implements Storage.DeleteTool
func (s *DiskStorage) DeleteTool(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.baseDir, "tools", name+".json")
	return os.Remove(path)
}

// SaveServer implements Storage.SaveServer
func (s *DiskStorage) SaveServer(ctx context.Context, server *server.StoredServer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	server.UpdatedAt = time.Now()
	if server.CreatedAt.IsZero() {
		server.CreatedAt = server.UpdatedAt
	}

	data, err := json.Marshal(server)
	if err != nil {
		return err
	}

	path := filepath.Join(s.baseDir, "servers", server.Name+".json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetServer implements Storage.GetServer
func (s *DiskStorage) GetServer(ctx context.Context, name string) (*server.StoredServer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.baseDir, "servers", name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var server server.StoredServer
	if err := json.Unmarshal(data, &server); err != nil {
		return nil, err
	}

	return &server, nil
}

// ListServers implements Storage.ListServers
func (s *DiskStorage) ListServers(ctx context.Context) ([]*server.StoredServer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := filepath.Join(s.baseDir, "servers")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var servers []*server.StoredServer
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			s.logger.Error("failed to read server file",
				zap.String("file", entry.Name()),
				zap.Error(err))
			continue
		}

		var server server.StoredServer
		if err := json.Unmarshal(data, &server); err != nil {
			s.logger.Error("failed to unmarshal server",
				zap.String("file", entry.Name()),
				zap.Error(err))
			continue
		}

		servers = append(servers, &server)
	}

	return servers, nil
}

// DeleteServer implements Storage.DeleteServer
func (s *DiskStorage) DeleteServer(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.baseDir, "servers", name+".json")
	return os.Remove(path)
}
