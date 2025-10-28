package goqueue

import (
	"context"
	"log"
	"sync"
	"time"
)

// workerPool manages multiple workers processing messages from a queue
type workerPool struct {
	backend Backend
	config  Config
	queue   string
	handler Handler
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// newWorkerPool creates a new worker pool
func newWorkerPool(backend Backend, config Config, queue string, handler Handler) *workerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &workerPool{
		backend: backend,
		config:  config,
		queue:   queue,
		handler: handler,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// start begins processing messages with the configured number of workers
func (wp *workerPool) start(ctx context.Context) {
	// Subscribe to the queue
	envChan, err := wp.backend.Subscribe(ctx, wp.queue)
	if err != nil {
		log.Printf("Failed to subscribe to queue %s: %v", wp.queue, err)
		return
	}

	// Start worker goroutines
	for i := 0; i < wp.config.WorkerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, envChan, i)
	}

	// Wait for context cancellation
	<-wp.ctx.Done()
}

// worker is a single worker goroutine that processes messages
func (wp *workerPool) worker(ctx context.Context, envChan <-chan *Envelope, workerID int) {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.ctx.Done():
			return
		case env, ok := <-envChan:
			if !ok {
				return
			}
			wp.processMessage(ctx, env, workerID)
		}
	}
}

// processMessage handles a single message envelope with retry logic
func (wp *workerPool) processMessage(ctx context.Context, env *Envelope, workerID int) {
	now := time.Now()
	env.ProcessedAt = &now

	// Call the handler
	err := wp.handler.Handle(ctx, env)

	if err == nil {
		// Success - acknowledge the message
		if ackErr := wp.backend.Ack(ctx, env.ID); ackErr != nil {
			log.Printf("Worker %d: Failed to ack message %s: %v", workerID, env.ID, ackErr)
		}
		return
	}

	// Handler returned an error
	log.Printf("Worker %d: Error processing message %s: %v", workerID, env.ID, err)

	// Check if we should retry
	if env.RetryCount < wp.config.RetryCount {
		wp.retryMessage(ctx, env, workerID)
	} else {
		// Max retries reached
		wp.handleFailedMessage(ctx, env, workerID)
	}
}

// retryMessage re-queues a message envelope for retry with exponential backoff
func (wp *workerPool) retryMessage(ctx context.Context, env *Envelope, workerID int) {
	env.RetryCount++

	// Calculate backoff delay: initial delay * 2^(retry_count-1)
	delay := wp.config.RetryDelay * time.Duration(1<<uint(env.RetryCount-1))
	log.Printf("Worker %d: Retrying message %s after %v (attempt %d/%d)",
		workerID, env.ID, delay, env.RetryCount, wp.config.RetryCount)

	// Wait for backoff delay
	time.Sleep(delay)

	// Re-publish the envelope to the same queue
	if err := wp.backend.Publish(ctx, wp.queue, env); err != nil {
		log.Printf("Worker %d: Failed to re-publish message %s: %v", workerID, env.ID, err)
		// Nack the original message since we couldn't retry
		if nackErr := wp.backend.Nack(ctx, env.ID); nackErr != nil {
			log.Printf("Worker %d: Failed to nack message %s: %v", workerID, env.ID, nackErr)
		}
		return
	}

	// Ack the original message since we've successfully re-queued it
	if err := wp.backend.Ack(ctx, env.ID); err != nil {
		log.Printf("Worker %d: Failed to ack message %s after retry: %v", workerID, env.ID, err)
	}
}

// handleFailedMessage processes a message envelope that has exceeded max retries
func (wp *workerPool) handleFailedMessage(ctx context.Context, env *Envelope, workerID int) {
	log.Printf("Worker %d: Message %s exceeded max retries (%d), sending to DLQ",
		workerID, env.ID, wp.config.RetryCount)

	// If DLQ is enabled, send the envelope to the dead letter queue
	if wp.config.DLQEnabled {
		// Update the queue name to DLQ
		originalQueue := env.Queue
		env.Queue = wp.config.DLQName
		if env.Metadata == nil {
			env.Metadata = make(map[string]string)
		}
		env.Metadata["original_queue"] = originalQueue
		env.Metadata["failure_reason"] = "max_retries_exceeded"

		if err := wp.backend.Publish(ctx, wp.config.DLQName, env); err != nil {
			log.Printf("Worker %d: Failed to publish message %s to DLQ: %v", workerID, env.ID, err)
		}
	}

	// Nack the original message
	if err := wp.backend.Nack(ctx, env.ID); err != nil {
		log.Printf("Worker %d: Failed to nack failed message %s: %v", workerID, env.ID, err)
	}
}

// stop gracefully shuts down the worker pool
func (wp *workerPool) stop() {
	wp.cancel()
	wp.wg.Wait()
}
