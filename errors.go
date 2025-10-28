package goqueue

import "errors"

var (
	// ErrQueueStopped is returned when trying to use a stopped queue
	ErrQueueStopped = errors.New("queue is stopped")

	// ErrWorkersAlreadyRunning is returned when trying to start workers for a queue that already has workers running
	ErrWorkersAlreadyRunning = errors.New("workers already running for this queue")

	// ErrMessageNotFound is returned when a message with the given ID is not found
	ErrMessageNotFound = errors.New("message not found")

	// ErrHandlerAlreadyRegistered is returned when trying to register a handler for a queue that already has one
	ErrHandlerAlreadyRegistered = errors.New("handler already registered for this queue")

	// ErrNoHandlersRegistered is returned when trying to start without any registered handlers
	ErrNoHandlersRegistered = errors.New("no handlers registered")
)
