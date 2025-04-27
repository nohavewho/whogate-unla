package notifier

import (
	"context"

	"github.com/mcp-ecosystem/mcp-gateway/internal/common/config"
	"go.uber.org/zap"
)

// Type represents the type of notifier
type Type string

const (
	// TypeSignal represents signal-based notifier
	TypeSignal Type = "signal"
	// TypeAPI represents API-based notifier
	TypeAPI Type = "api"
	// TypeRedis represents Redis-based notifier
	TypeRedis Type = "redis"
	// TypeComposite represents composite notifier
	TypeComposite Type = "composite"
)

// NewNotifier creates a new notifier based on the configuration
func NewNotifier(ctx context.Context, logger *zap.Logger, cfg *config.NotifierConfig) (Notifier, error) {
	switch Type(cfg.Type) {
	case TypeSignal:
		return NewSignalNotifier(ctx, logger, cfg.Signal.PID), nil
	case TypeAPI:
		return NewAPINotifier(logger, cfg.API.Port), nil
	case TypeRedis:
		return NewRedisNotifier(logger, cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB, cfg.Redis.Topic)
	case TypeComposite:
		notifiers := make([]Notifier, 0)
		// Add signal notifier
		signalNotifier := NewSignalNotifier(ctx, logger, cfg.Signal.PID)
		notifiers = append(notifiers, signalNotifier)
		// Add API notifier
		apiNotifier := NewAPINotifier(logger, cfg.API.Port)
		notifiers = append(notifiers, apiNotifier)
		// Add Redis notifier if configured
		if cfg.Redis.Addr != "" {
			redisNotifier, err := NewRedisNotifier(logger, cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB, cfg.Redis.Topic)
			if err != nil {
				return nil, err
			}
			notifiers = append(notifiers, redisNotifier)
		}
		return NewCompositeNotifier(logger, notifiers...), nil
	default:
		return NewSignalNotifier(ctx, logger, cfg.Signal.PID), nil
	}
}
