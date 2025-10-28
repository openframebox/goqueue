package goqueue

import (
	"context"
	"encoding/json"
	"sync"
)

// GoQueue is the main queue manager that coordinates workers and the backend
type GoQueue struct {
	backend  Backend
	config   Config
	handlers map[string]Handler
	workers  map[string]*workerPool
	mu       sync.RWMutex
	stopped  bool
}

// New creates a new GoQueue instance with the given backend and options
func New(backend Backend, opts ...Option) *GoQueue {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(&config)
	}

	return &GoQueue{
		backend:  backend,
		config:   config,
		handlers: make(map[string]Handler),
		workers:  make(map[string]*workerPool),
		stopped:  false,
	}
}

// Publish publishes a message to its designated queue.
// The message declares which queue it belongs to via QueueName().
// The message will be automatically serialized to JSON.
func (gq *GoQueue) Publish(ctx context.Context, message QueueMessage) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}

	queue := message.QueueName()
	return gq.PublishBytes(ctx, queue, jsonData)
}

// PublishBytes publishes raw bytes to the specified queue.
func (gq *GoQueue) PublishBytes(ctx context.Context, queue string, data []byte) error {
	gq.mu.RLock()
	if gq.stopped {
		gq.mu.RUnlock()
		return ErrQueueStopped
	}
	gq.mu.RUnlock()

	env := NewEnvelope(queue, data)
	return gq.backend.Publish(ctx, queue, env)
}

// Register registers a handler for its designated queue.
// Must be called before Start().
func (gq *GoQueue) Register(handler Handler) error {
	gq.mu.Lock()
	defer gq.mu.Unlock()

	if gq.stopped {
		return ErrQueueStopped
	}

	queue := handler.QueueName()
	if _, exists := gq.handlers[queue]; exists {
		return ErrHandlerAlreadyRegistered
	}

	gq.handlers[queue] = handler
	return nil
}

// Start starts workers for all registered queues.
func (gq *GoQueue) Start(ctx context.Context) error {
	gq.mu.Lock()
	defer gq.mu.Unlock()

	if gq.stopped {
		return ErrQueueStopped
	}

	if len(gq.handlers) == 0 {
		return ErrNoHandlersRegistered
	}

	for queue, handler := range gq.handlers {
		if _, exists := gq.workers[queue]; exists {
			continue
		}

		wp := newWorkerPool(gq.backend, gq.config, queue, handler)
		gq.workers[queue] = wp

		go wp.start(ctx)
	}

	return nil
}

// Stop gracefully stops all workers and closes the backend connection
func (gq *GoQueue) Stop() error {
	gq.mu.Lock()
	defer gq.mu.Unlock()

	if gq.stopped {
		return nil
	}

	gq.stopped = true

	for _, wp := range gq.workers {
		wp.stop()
	}

	return gq.backend.Close()
}
