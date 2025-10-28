package goqueue

import "context"

// Handler defines the interface that message handlers must implement.
// Handlers declare which queue they process and how to handle messages.
type Handler interface {
	// QueueName returns the name of the queue this handler processes
	QueueName() string

	// Handle processes a single message envelope.
	// Return an error to trigger retry logic, or nil on success.
	Handle(ctx context.Context, envelope *Envelope) error
}

// HandlerFunc is a function adapter that allows using functions as Handlers.
// This provides backward compatibility and convenience for simple use cases.
type HandlerFunc struct {
	queueName string
	fn        func(ctx context.Context, envelope *Envelope) error
}

// NewHandlerFunc creates a new HandlerFunc for the given queue
func NewHandlerFunc(queueName string, fn func(ctx context.Context, envelope *Envelope) error) HandlerFunc {
	return HandlerFunc{
		queueName: queueName,
		fn:        fn,
	}
}

// QueueName implements the Handler interface
func (f HandlerFunc) QueueName() string {
	return f.queueName
}

// Handle implements the Handler interface for HandlerFunc
func (f HandlerFunc) Handle(ctx context.Context, envelope *Envelope) error {
	return f.fn(ctx, envelope)
}
