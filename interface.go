package goqueue

import "context"

// Backend defines the interface that all queue backends must implement.
// This allows easy swapping between different queue implementations
// (e.g., Redis, RabbitMQ, in-memory, etc.)
type Backend interface {
	// Publish sends a message envelope to the specified queue
	Publish(ctx context.Context, queue string, envelope *Envelope) error

	// Subscribe creates a subscription to the specified queue and returns
	// a channel that will receive message envelopes
	Subscribe(ctx context.Context, queue string) (<-chan *Envelope, error)

	// Ack acknowledges successful processing of a message
	Ack(ctx context.Context, messageID string) error

	// Nack indicates that a message failed to process and should be retried
	Nack(ctx context.Context, messageID string) error

	// Close releases any resources held by the backend
	Close() error
}

