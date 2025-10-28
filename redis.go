package goqueue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisBackend is a Redis-based queue backend.
// This is suitable for production use with distributed systems.
type RedisBackend struct {
	client      *redis.Client
	subscribers map[string]chan *Envelope // queue name -> channel
	mu          sync.RWMutex
	closed      bool
	stopChans   map[string]chan struct{} // queue name -> stop channel
	pollTimeout time.Duration

	// Redis connection options
	password string
	db       int
}

// RedisOption is a function that configures the Redis backend
type RedisOption func(*RedisBackend)

// WithPollTimeout sets the timeout for polling messages from Redis
func WithPollTimeout(timeout time.Duration) RedisOption {
	return func(b *RedisBackend) {
		if timeout > 0 {
			b.pollTimeout = timeout
		}
	}
}

// WithRedisPassword sets the password for Redis authentication
func WithRedisPassword(password string) RedisOption {
	return func(b *RedisBackend) {
		b.password = password
	}
}

// WithRedisDB sets the Redis database number (0-15)
func WithRedisDB(db int) RedisOption {
	return func(b *RedisBackend) {
		if db >= 0 {
			b.db = db
		}
	}
}

// NewRedisBackend creates a new Redis backend
// addr is the Redis server address (e.g., "localhost:6379")
func NewRedisBackend(addr string, opts ...RedisOption) *RedisBackend {
	b := &RedisBackend{
		subscribers: make(map[string]chan *Envelope),
		stopChans:   make(map[string]chan struct{}),
		pollTimeout: 5 * time.Second, // default poll timeout
		db:          0,                // default database
	}

	// Apply options first to get password and db
	for _, opt := range opts {
		opt(b)
	}

	// Create client with options
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: b.password,
		DB:       b.db,
	})

	b.client = client

	return b
}

// NewRedisBackendWithClient creates a new Redis backend with an existing Redis client
func NewRedisBackendWithClient(client *redis.Client, opts ...RedisOption) *RedisBackend {
	b := &RedisBackend{
		client:      client,
		subscribers: make(map[string]chan *Envelope),
		stopChans:   make(map[string]chan struct{}),
		pollTimeout: 5 * time.Second,
	}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

// queueKey returns the Redis key for a queue
func (b *RedisBackend) queueKey(queue string) string {
	return fmt.Sprintf("goqueue:queue:%s", queue)
}

// pendingKey returns the Redis key for a pending message
func (b *RedisBackend) pendingKey(messageID string) string {
	return fmt.Sprintf("goqueue:pending:%s", messageID)
}

// Publish sends a message envelope to the specified queue
func (b *RedisBackend) Publish(ctx context.Context, queue string, envelope *Envelope) error {
	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return ErrQueueStopped
	}
	b.mu.RUnlock()

	// Serialize envelope to JSON
	data, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("failed to serialize envelope: %w", err)
	}

	// Store in pending messages
	pendingKey := b.pendingKey(envelope.ID)
	if err := b.client.Set(ctx, pendingKey, data, 0).Err(); err != nil {
		return fmt.Errorf("failed to store pending message: %w", err)
	}

	// Push to queue (LPUSH adds to the left/head of the list)
	queueKey := b.queueKey(queue)
	if err := b.client.LPush(ctx, queueKey, envelope.ID).Err(); err != nil {
		// Clean up pending message if queue push fails
		b.client.Del(ctx, pendingKey)
		return fmt.Errorf("failed to push to queue: %w", err)
	}

	return nil
}

// Subscribe creates a subscription to the specified queue
func (b *RedisBackend) Subscribe(ctx context.Context, queue string) (<-chan *Envelope, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil, ErrQueueStopped
	}

	// Check if already subscribed
	if ch, exists := b.subscribers[queue]; exists {
		return ch, nil
	}

	// Create channel for this queue
	ch := make(chan *Envelope, 100) // buffer size
	stopChan := make(chan struct{})

	b.subscribers[queue] = ch
	b.stopChans[queue] = stopChan

	// Start polling goroutine
	go b.pollQueue(ctx, queue, ch, stopChan)

	return ch, nil
}

// pollQueue continuously polls Redis for new messages
func (b *RedisBackend) pollQueue(ctx context.Context, queue string, ch chan<- *Envelope, stopChan <-chan struct{}) {
	queueKey := b.queueKey(queue)

	for {
		select {
		case <-stopChan:
			close(ch)
			return
		case <-ctx.Done():
			close(ch)
			return
		default:
			// BRPOP blocks until a message is available or timeout occurs
			// Returns [queueKey, messageID]
			result, err := b.client.BRPop(ctx, b.pollTimeout, queueKey).Result()
			if err != nil {
				if err == redis.Nil {
					// Timeout, continue polling
					continue
				}
				// Log error and continue
				continue
			}

			if len(result) != 2 {
				continue
			}

			messageID := result[1]

			// Retrieve envelope from pending messages
			pendingKey := b.pendingKey(messageID)
			data, err := b.client.Get(ctx, pendingKey).Result()
			if err != nil {
				// Message not found in pending, skip
				continue
			}

			// Deserialize envelope
			var envelope Envelope
			if err := json.Unmarshal([]byte(data), &envelope); err != nil {
				// Failed to deserialize, clean up and skip
				b.client.Del(ctx, pendingKey)
				continue
			}

			// Send to channel
			select {
			case ch <- &envelope:
			case <-stopChan:
				close(ch)
				return
			case <-ctx.Done():
				close(ch)
				return
			}
		}
	}
}

// Ack acknowledges successful processing of a message
func (b *RedisBackend) Ack(ctx context.Context, messageID string) error {
	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return ErrQueueStopped
	}
	b.mu.RUnlock()

	// Remove from pending messages
	pendingKey := b.pendingKey(messageID)
	result := b.client.Del(ctx, pendingKey)
	if result.Err() != nil {
		return fmt.Errorf("failed to ack message: %w", result.Err())
	}

	if result.Val() == 0 {
		return ErrMessageNotFound
	}

	return nil
}

// Nack indicates that a message failed to process
func (b *RedisBackend) Nack(ctx context.Context, messageID string) error {
	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return ErrQueueStopped
	}
	b.mu.RUnlock()

	// For Redis backend, we just remove it from pending
	// The retry logic is handled by the worker pool which will re-publish
	pendingKey := b.pendingKey(messageID)
	result := b.client.Del(ctx, pendingKey)
	if result.Err() != nil {
		return fmt.Errorf("failed to nack message: %w", result.Err())
	}

	if result.Val() == 0 {
		return ErrMessageNotFound
	}

	return nil
}

// Close releases resources held by the backend
func (b *RedisBackend) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	b.closed = true

	// Stop all polling goroutines
	for _, stopChan := range b.stopChans {
		close(stopChan)
	}

	// Close Redis client
	return b.client.Close()
}

// Ping checks if the Redis connection is alive
func (b *RedisBackend) Ping(ctx context.Context) error {
	return b.client.Ping(ctx).Err()
}
