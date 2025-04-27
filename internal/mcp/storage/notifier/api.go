package notifier

import (
	"context"
	"errors"
	"fmt"
	"github.com/mcp-ecosystem/mcp-gateway/internal/common/cnst"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/mcp-ecosystem/mcp-gateway/internal/common/config"
	"go.uber.org/zap"
)

// APINotifier implements Notifier using HTTP API
type APINotifier struct {
	logger   *zap.Logger
	watchers map[chan<- *config.MCPConfig]struct{}
	mu       sync.RWMutex
	router   *gin.Engine
	server   *http.Server
	role     config.NotifierRole
}

// NewAPINotifier creates a new API-based notifier
func NewAPINotifier(logger *zap.Logger, port int, role config.NotifierRole) *APINotifier {
	n := &APINotifier{
		logger:   logger.Named("notifier.api"),
		watchers: make(map[chan<- *config.MCPConfig]struct{}),
		router:   gin.Default(),
		role:     role,
	}

	// Setup API routes if can receive
	if n.CanReceive() {
		n.router.POST("/_reload", func(c *gin.Context) {
			_ = n.NotifyUpdate(c.Request.Context(), nil)
			c.JSON(http.StatusOK, gin.H{"status": "reload triggered"})
		})

		// Start HTTP server
		n.server = &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: n.router,
		}

		go func() {
			if err := n.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				n.logger.Fatal("failed to start API server", zap.Error(err))
			}
		}()
	}

	return n
}

// Watch implements Notifier.Watch
func (n *APINotifier) Watch(ctx context.Context) (<-chan *config.MCPConfig, error) {
	if !n.CanReceive() {
		return nil, cnst.ErrNotReceiver
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	ch := make(chan *config.MCPConfig, 10)
	n.watchers[ch] = struct{}{}

	// Start a goroutine to handle context cancellation and cleanup
	go func() {
		<-ctx.Done()
		n.mu.Lock()
		defer n.mu.Unlock()
		delete(n.watchers, ch)
		close(ch)
	}()

	return ch, nil
}

// NotifyUpdate implements Notifier.NotifyUpdate
func (n *APINotifier) NotifyUpdate(_ context.Context, server *config.MCPConfig) error {
	if !n.CanSend() {
		return cnst.ErrNotSender
	}

	n.mu.RLock()
	defer n.mu.RUnlock()

	for ch := range n.watchers {
		select {
		case ch <- server:
		default:
			n.logger.Warn("watcher channel is full, skipping notification",
				zap.String("server", server.Name))
		}
	}
	return nil
}

// Shutdown gracefully shuts down the API server
func (n *APINotifier) Shutdown(ctx context.Context) error {
	if n.server != nil {
		return n.server.Shutdown(ctx)
	}
	return nil
}

// CanReceive returns true if the notifier can receive updates
func (n *APINotifier) CanReceive() bool {
	return n.role == config.RoleReceiver || n.role == config.RoleBoth
}

// CanSend returns true if the notifier can send updates
func (n *APINotifier) CanSend() bool {
	return n.role == config.RoleSender || n.role == config.RoleBoth
}
