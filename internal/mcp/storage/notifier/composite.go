package notifier

import (
	"context"
	"sync"

	"github.com/mcp-ecosystem/mcp-gateway/internal/common/config"
	"go.uber.org/zap"
)

// CompositeNotifier implements Notifier by combining multiple notifiers
type CompositeNotifier struct {
	logger    *zap.Logger
	notifiers []Notifier
	mu        sync.RWMutex
	watchers  map[chan<- *config.MCPConfig]struct{}
}

// NewCompositeNotifier creates a new composite notifier
func NewCompositeNotifier(logger *zap.Logger, notifiers ...Notifier) *CompositeNotifier {
	return &CompositeNotifier{
		logger:    logger.Named("notifier.composite"),
		notifiers: notifiers,
		watchers:  make(map[chan<- *config.MCPConfig]struct{}),
	}
}

// Watch implements Notifier.Watch
func (n *CompositeNotifier) Watch(ctx context.Context) (<-chan *config.MCPConfig, error) {
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
func (n *CompositeNotifier) NotifyUpdate(ctx context.Context, server *config.MCPConfig) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// Notify all watchers
	for ch := range n.watchers {
		select {
		case ch <- server:
		default:
			n.logger.Warn("watcher channel is full, skipping notification",
				zap.String("server", server.Name))
		}
	}

	// Notify all underlying notifiers
	for _, notifier := range n.notifiers {
		if err := notifier.NotifyUpdate(ctx, server); err != nil {
			n.logger.Error("failed to notify underlying notifier",
				zap.Error(err))
		}
	}

	return nil
}
