package notifier

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/mcp-ecosystem/mcp-gateway/internal/common/config"
	"go.uber.org/zap"
)

// SignalNotifier implements Notifier using system signals
type SignalNotifier struct {
	logger   *zap.Logger
	watchers map[chan<- *config.MCPConfig]struct{}
	mu       sync.RWMutex
}

// NewSignalNotifier creates a new signal-based notifier
func NewSignalNotifier(ctx context.Context, logger *zap.Logger) *SignalNotifier {
	n := &SignalNotifier{
		logger:   logger.Named("notifier.signal"),
		watchers: make(map[chan<- *config.MCPConfig]struct{}),
	}

	// Start signal handler
	go n.handleSignals(ctx)

	return n
}

func (n *SignalNotifier) handleSignals(ctx context.Context) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)

	for {
		select {
		case sig := <-sigChan:
			n.logger.Info("Received reload signal", zap.String("signal", sig.String()))
			// When signal is received, notify all watchers with nil server
			// The actual server config will be reloaded by the watcher
			_ = n.NotifyUpdate(ctx, nil)
		}
	}
}

// Watch implements Notifier.Watch
func (n *SignalNotifier) Watch(ctx context.Context) (<-chan *config.MCPConfig, error) {
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
func (n *SignalNotifier) NotifyUpdate(ctx context.Context, server *config.MCPConfig) error {
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
