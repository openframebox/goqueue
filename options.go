package goqueue

import "time"

// Config holds the configuration for a GoQueue instance
type Config struct {
	// WorkerCount is the number of concurrent workers processing messages
	WorkerCount int

	// RetryCount is the maximum number of times a failed message will be retried
	RetryCount int

	// RetryDelay is the initial delay between retries (uses exponential backoff)
	RetryDelay time.Duration

	// DLQEnabled enables the dead letter queue for permanently failed messages
	DLQEnabled bool

	// DLQName is the name of the dead letter queue
	DLQName string
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() Config {
	return Config{
		WorkerCount: 1,
		RetryCount:  3,
		RetryDelay:  1 * time.Second,
		DLQEnabled:  false,
		DLQName:     "dead-letter-queue",
	}
}

// Option is a function that modifies a Config
type Option func(*Config)

// WithWorkerCount sets the number of concurrent workers
func WithWorkerCount(count int) Option {
	return func(c *Config) {
		if count > 0 {
			c.WorkerCount = count
		}
	}
}

// WithRetryCount sets the maximum number of retries for failed messages
func WithRetryCount(count int) Option {
	return func(c *Config) {
		if count >= 0 {
			c.RetryCount = count
		}
	}
}

// WithRetryDelay sets the initial retry delay (exponential backoff is applied)
func WithRetryDelay(delay time.Duration) Option {
	return func(c *Config) {
		if delay > 0 {
			c.RetryDelay = delay
		}
	}
}

// WithDLQ enables the dead letter queue with the specified name
func WithDLQ(queueName string) Option {
	return func(c *Config) {
		c.DLQEnabled = true
		c.DLQName = queueName
	}
}
