package goqueue

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// QueueMessage is the interface that all queue messages must implement.
type QueueMessage interface {
	QueueName() string
}

// Envelope represents the internal wrapper around queue messages with metadata.
type Envelope struct {
	ID          string            `json:"id"`
	Queue       string            `json:"queue"`
	Data        []byte            `json:"data"`
	RetryCount  int               `json:"retry_count"`
	CreatedAt   time.Time         `json:"created_at"`
	ProcessedAt *time.Time        `json:"processed_at,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// NewEnvelope creates a new envelope with the given queue name and data
func NewEnvelope(queue string, data []byte) *Envelope {
	return &Envelope{
		ID:         uuid.New().String(),
		Queue:      queue,
		Data:       data,
		RetryCount: 0,
		CreatedAt:  time.Now(),
		Metadata:   make(map[string]string),
	}
}

// Unmarshal unmarshals the envelope data into the provided value
func (e *Envelope) Unmarshal(v any) error {
	return json.Unmarshal(e.Data, v)
}

// String returns a string representation of the envelope
func (e *Envelope) String() string {
	return "Envelope{ID: " + e.ID + ", Queue: " + e.Queue + ", RetryCount: " + string(rune(e.RetryCount)) + "}"
}
