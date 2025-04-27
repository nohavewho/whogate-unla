package notifier

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/mcp-ecosystem/mcp-gateway/internal/common/cnst"

	"github.com/gin-gonic/gin"
	"github.com/mcp-ecosystem/mcp-gateway/internal/common/config"
	"go.uber.org/zap"
)

// APINotifier implements Notifier using HTTP API
type APINotifier struct {
	logger    *zap.Logger
	watchers  map[chan<- *config.MCPConfig]struct{}
	mu        sync.RWMutex
	router    *gin.Engine
	server    *http.Server
	role      config.NotifierRole
	targetURL string
}

// NewAPINotifier creates a new API-based notifier
func NewAPINotifier(logger *zap.Logger, port int, role config.NotifierRole, targetURL string) *APINotifier {
	n := &APINotifier{
		logger:    logger.Named("notifier.api"),
		watchers:  make(map[chan<- *config.MCPConfig]struct{}),
		router:    gin.Default(),
		role:      role,
		targetURL: targetURL,
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
func (n *APINotifier) NotifyUpdate(ctx context.Context, server *config.MCPConfig) error {
	if !n.CanSend() {
		return cnst.ErrNotSender
	}

	if n.targetURL == "" {
		return fmt.Errorf("target URL is not configured")
	}

	// Send HTTP POST request to target URL
	req, err := http.NewRequestWithContext(ctx, "POST", n.targetURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
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
