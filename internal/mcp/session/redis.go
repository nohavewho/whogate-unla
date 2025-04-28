package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RedisStore implements Store using Redis
type RedisStore struct {
	logger *zap.Logger
	client *redis.Client
	prefix string
	topic  string
	pubsub *redis.PubSub
	// Add a map to track active connections
	connections map[string]*RedisConnection
	mu          sync.RWMutex
}

var _ Store = (*RedisStore)(nil)

// NewRedisStore creates a new Redis-based session store
func NewRedisStore(logger *zap.Logger, addr, password string, db int, topic string) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	store := &RedisStore{
		logger:      logger.Named("session.store.redis"),
		client:      client,
		prefix:      "session:",
		topic:       topic,
		connections: make(map[string]*RedisConnection),
	}

	// Subscribe to session updates
	store.pubsub = client.Subscribe(context.Background(), topic)
	go store.handleUpdates()

	return store, nil
}

// handleUpdates handles session update notifications
func (s *RedisStore) handleUpdates() {
	ch := s.pubsub.Channel()
	for msg := range ch {
		var update struct {
			Action  string   `json:"action"` // "create", "update", "delete", "event"
			Meta    *Meta    `json:"meta"`
			Message *Message `json:"message,omitempty"`
		}
		if err := json.Unmarshal([]byte(msg.Payload), &update); err != nil {
			s.logger.Error("failed to unmarshal session update",
				zap.Error(err),
				zap.String("payload", msg.Payload))
			continue
		}

		// Update local cache if needed
		switch update.Action {
		case "create", "update":
			// Update local cache
			s.logger.Debug("received session update",
				zap.String("action", update.Action),
				zap.String("id", update.Meta.ID))
		case "delete":
			// Remove from local cache
			s.logger.Debug("received session delete",
				zap.String("id", update.Meta.ID))
		case "event":
			// Handle event and send to appropriate connection
			s.mu.RLock()
			conn, exists := s.connections[update.Meta.ID]
			s.mu.RUnlock()

			if exists {
				select {
				case conn.queue <- update.Message:
					s.logger.Info("sent message to connection queue",
						zap.String("id", update.Meta.ID),
						zap.String("event", update.Message.Event))
				default:
					s.logger.Warn("connection queue is full, dropping message",
						zap.String("id", update.Meta.ID),
						zap.String("event", update.Message.Event))
				}
			} else {
				s.logger.Warn("received event for non-existent connection",
					zap.String("id", update.Meta.ID),
					zap.String("event", update.Message.Event))
			}
		}
	}
}

// publishUpdate publishes a session update to the topic
func (s *RedisStore) publishUpdate(ctx context.Context, action string, meta *Meta, msg *Message) error {
	update := struct {
		Action  string   `json:"action"`
		Meta    *Meta    `json:"meta"`
		Message *Message `json:"message,omitempty"`
	}{
		Action:  action,
		Meta:    meta,
		Message: msg,
	}

	data, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to marshal session update: %w", err)
	}

	return s.client.Publish(ctx, s.topic, data).Err()
}

// Register implements Store.Register
func (s *RedisStore) Register(ctx context.Context, meta *Meta) (Connection, error) {
	// Store metadata
	data, err := json.Marshal(meta)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session metadata: %w", err)
	}

	key := s.prefix + meta.ID
	if err := s.client.Set(ctx, key, data, 0).Err(); err != nil {
		return nil, fmt.Errorf("failed to store session metadata in Redis: %w", err)
	}

	// Add session ID to the list of valid sessions
	if err := s.client.SAdd(ctx, s.prefix+"ids", meta.ID).Err(); err != nil {
		return nil, fmt.Errorf("failed to add session ID to list: %w", err)
	}

	// Create connection
	conn := &RedisConnection{
		store: s,
		meta:  meta,
		queue: make(chan *Message, 100),
	}

	// Add to active connections
	s.mu.Lock()
	s.connections[meta.ID] = conn
	s.mu.Unlock()

	// Publish update
	if err := s.publishUpdate(ctx, "create", meta, nil); err != nil {
		return nil, fmt.Errorf("failed to publish session creation: %w", err)
	}

	return conn, nil
}

// Get implements Store.Get
func (s *RedisStore) Get(ctx context.Context, id string) (Connection, error) {
	// Check if session ID is valid
	exists, err := s.client.SIsMember(ctx, s.prefix+"ids", id).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to check session ID: %w", err)
	}
	if !exists {
		return nil, ErrSessionNotFound
	}

	key := s.prefix + id
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session metadata from Redis: %w", err)
	}

	var meta Meta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session metadata: %w", err)
	}

	return &RedisConnection{
		store: s,
		meta:  &meta,
		queue: make(chan *Message, 100),
	}, nil
}

// Unregister implements Store.Unregister
func (s *RedisStore) Unregister(ctx context.Context, id string) error {
	// Remove from active connections
	s.mu.Lock()
	delete(s.connections, id)
	s.mu.Unlock()

	// Check if session ID is valid
	exists, err := s.client.SIsMember(ctx, s.prefix+"ids", id).Result()
	if err != nil {
		return fmt.Errorf("failed to check session ID: %w", err)
	}
	if !exists {
		return ErrSessionNotFound
	}

	key := s.prefix + id
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete session metadata from Redis: %w", err)
	}

	// Remove session ID from the list
	if err := s.client.SRem(ctx, s.prefix+"ids", id).Err(); err != nil {
		return fmt.Errorf("failed to remove session ID from list: %w", err)
	}

	// Publish delete
	meta := &Meta{ID: id}
	return s.publishUpdate(ctx, "delete", meta, nil)
}

// List implements Store.List
func (s *RedisStore) List(ctx context.Context) ([]Connection, error) {
	// Get all valid session IDs
	ids, err := s.client.SMembers(ctx, s.prefix+"ids").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get session IDs: %w", err)
	}

	connections := make([]Connection, 0, len(ids))
	for _, id := range ids {
		key := s.prefix + id
		data, err := s.client.Get(ctx, key).Bytes()
		if err != nil {
			s.logger.Error("failed to get session metadata",
				zap.String("key", key),
				zap.Error(err))
			continue
		}

		var meta Meta
		if err := json.Unmarshal(data, &meta); err != nil {
			s.logger.Error("failed to unmarshal session metadata",
				zap.String("key", key),
				zap.Error(err))
			continue
		}

		connections = append(connections, &RedisConnection{
			store: s,
			meta:  &meta,
			queue: make(chan *Message, 100),
		})
	}

	return connections, nil
}

// Close closes the Redis store
func (s *RedisStore) Close() error {
	if s.pubsub != nil {
		if err := s.pubsub.Close(); err != nil {
			return fmt.Errorf("failed to close pubsub: %w", err)
		}
	}
	return s.client.Close()
}

// RedisConnection implements Connection using Redis
type RedisConnection struct {
	store *RedisStore
	meta  *Meta
	queue chan *Message
}

var _ Connection = (*RedisConnection)(nil)

// EventQueue implements Connection.EventQueue
func (c *RedisConnection) EventQueue() <-chan *Message {
	return c.queue
}

// Send implements Connection.Send
func (c *RedisConnection) Send(ctx context.Context, msg *Message) error {
	return c.store.publishUpdate(ctx, "event", c.meta, msg)
}

// Close implements Connection.Close
func (c *RedisConnection) Close(ctx context.Context) error {
	return c.store.Unregister(ctx, c.meta.ID)
}

// Meta implements Connection.Meta
func (c *RedisConnection) Meta() *Meta {
	return c.meta
}
