package goqueue

import (
	"context"
	"sync"
)

// MemoryBackend is an in-memory queue backend using Go channels.
// This is ideal for testing, development, and single-process applications.
type MemoryBackend struct {
	queues      map[string]chan *Envelope
	pendingMsgs map[string]*Envelope // messageID -> envelope
	mu          sync.RWMutex
	closed      bool
	bufferSize  int
}

// MemoryOption is a function that configures the memory backend
type MemoryOption func(*MemoryBackend)

// WithBufferSize sets the channel buffer size for each queue
func WithBufferSize(size int) MemoryOption {
	return func(b *MemoryBackend) {
		if size > 0 {
			b.bufferSize = size
		}
	}
}

// NewMemoryBackend creates a new in-memory backend
func NewMemoryBackend(opts ...MemoryOption) *MemoryBackend {
	b := &MemoryBackend{
		queues:      make(map[string]chan *Envelope),
		pendingMsgs: make(map[string]*Envelope),
		bufferSize:  100, // default buffer size
	}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

// Publish sends a message envelope to the specified queue
func (b *MemoryBackend) Publish(ctx context.Context, queue string, envelope *Envelope) error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return ErrQueueStopped
	}

	// Get or create the queue channel
	queueChan, exists := b.queues[queue]
	if !exists {
		queueChan = make(chan *Envelope, b.bufferSize)
		b.queues[queue] = queueChan
	}

	// Store the envelope as pending
	b.pendingMsgs[envelope.ID] = envelope
	b.mu.Unlock()

	// Send to channel (non-blocking with context)
	select {
	case queueChan <- envelope:
		return nil
	case <-ctx.Done():
		// Remove from pending if context cancelled
		b.mu.Lock()
		delete(b.pendingMsgs, envelope.ID)
		b.mu.Unlock()
		return ctx.Err()
	}
}

// Subscribe creates a subscription to the specified queue
func (b *MemoryBackend) Subscribe(ctx context.Context, queue string) (<-chan *Envelope, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil, ErrQueueStopped
	}

	// Get or create the queue channel
	queueChan, exists := b.queues[queue]
	if !exists {
		queueChan = make(chan *Envelope, b.bufferSize)
		b.queues[queue] = queueChan
	}

	return queueChan, nil
}

// Ack acknowledges successful processing of a message
func (b *MemoryBackend) Ack(ctx context.Context, messageID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return ErrQueueStopped
	}

	// Remove from pending messages
	if _, exists := b.pendingMsgs[messageID]; !exists {
		return ErrMessageNotFound
	}

	delete(b.pendingMsgs, messageID)
	return nil
}

// Nack indicates that a message failed to process
func (b *MemoryBackend) Nack(ctx context.Context, messageID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return ErrQueueStopped
	}

	// For in-memory backend, we just remove it from pending
	// The retry logic is handled by the worker pool
	if _, exists := b.pendingMsgs[messageID]; !exists {
		return ErrMessageNotFound
	}

	delete(b.pendingMsgs, messageID)
	return nil
}

// Close releases resources held by the backend
func (b *MemoryBackend) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	b.closed = true

	// Close all queue channels
	for _, queueChan := range b.queues {
		close(queueChan)
	}

	// Clear pending messages
	b.pendingMsgs = make(map[string]*Envelope)

	return nil
}
