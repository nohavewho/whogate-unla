package notifier

import (
	"context"
	"fmt"
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
}

// NewAPINotifier creates a new API-based notifier
func NewAPINotifier(logger *zap.Logger, port int) *APINotifier {
	n := &APINotifier{
		logger:   logger.Named("notifier.api"),
		watchers: make(map[chan<- *config.MCPConfig]struct{}),
		router:   gin.Default(),
	}

	// Setup API routes
	n.router.POST("/_reload", func(c *gin.Context) {
		n.NotifyUpdate(c.Request.Context(), nil)
		c.JSON(http.StatusOK, gin.H{"status": "reload triggered"})
	})

	// Start HTTP server
	n.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: n.router,
	}

	go func() {
		if err := n.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			n.logger.Fatal("failed to start API server", zap.Error(err))
		}
	}()

	return n
}

// Watch implements Notifier.Watch
func (n *APINotifier) Watch(ctx context.Context) (<-chan *config.MCPConfig, error) {
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
	return n.server.Shutdown(ctx)
}
