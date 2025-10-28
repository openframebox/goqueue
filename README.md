# GoQueue

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go" alt="Go Version" />
  <img src="https://img.shields.io/github/license/openframebox/goqueue?style=for-the-badge" alt="License" />
  <img src="https://img.shields.io/badge/status-active-success?style=for-the-badge" alt="Status" />
</p>

<p align="center">
  <strong>A simple, flexible, and production-ready queue library for Go</strong>
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#installation">Installation</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#documentation">Documentation</a> •
  <a href="#examples">Examples</a> •
  <a href="#contributing">Contributing</a>
</p>

---

## Overview

GoQueue is a lightweight, interface-driven queue library that makes it easy to add background job processing to your Go applications. With support for multiple backends and a clean, intuitive API, you can start simple with in-memory queues and scale to production with Redis.

### Why GoQueue?

- **🎯 Simple API** - Minimal boilerplate, maximum productivity
- **🔌 Pluggable Backends** - Swap backends without changing your code
- **🚀 Production Ready** - Built-in retry logic, DLQ, and graceful shutdown
- **⚡ High Performance** - Concurrent workers with configurable pool sizes
- **📦 Zero Dependencies** (for in-memory backend)
- **🏗️ Clean Architecture** - Interface-based design for easy testing

## Features

- ✅ **Interface-Based Handlers** - Type-safe message handlers with dependency injection
- ✅ **Multiple Backends** - In-memory (dev/test) and Redis (production)
- ✅ **Concurrent Workers** - Process messages in parallel with configurable workers
- ✅ **Automatic Retry** - Exponential backoff with configurable retry attempts
- ✅ **Dead Letter Queue** - Handle permanently failed messages
- ✅ **Graceful Shutdown** - Context-based cancellation
- ✅ **JSON Serialization** - Automatic message serialization/deserialization
- ✅ **Distributed Processing** - Scale horizontally with Redis backend

## Installation

```bash
go get github.com/openframebox/goqueue
```

### Requirements

- Go 1.21 or higher
- Redis 5.0+ (optional, for Redis backend)

## Quick Start

### 1. Define Your Message

```go
type EmailTask struct {
    To      string `json:"to"`
    Subject string `json:"subject"`
    Body    string `json:"body"`
}

func (e *EmailTask) QueueName() string {
    return "emails"
}
```

### 2. Create a Handler

```go
type EmailHandler struct {
    emailService *EmailService
}

func (h *EmailHandler) QueueName() string {
    return "emails"
}

func (h *EmailHandler) Handle(ctx context.Context, envelope *goqueue.Envelope) error {
    var task EmailTask
    if err := envelope.Unmarshal(&task); err != nil {
        return err
    }

    return h.emailService.Send(task.To, task.Subject, task.Body)
}
```

### 3. Set Up the Queue

```go
// Create backend (in-memory for development)
backend := goqueue.NewMemoryBackend()

// Or use Redis for production
// backend := goqueue.NewRedisBackend("localhost:6379")

// Create queue with options
gq := goqueue.New(backend,
    goqueue.WithWorkerCount(5),
    goqueue.WithRetryCount(3),
    goqueue.WithDLQ("failed-emails"),
)

// Register handler
gq.Register(&EmailHandler{emailService: myEmailService})

// Start processing
ctx := context.Background()
gq.Start(ctx)

// Publish messages
gq.Publish(ctx, &EmailTask{
    To:      "user@example.com",
    Subject: "Welcome!",
    Body:    "Thanks for signing up!",
})

// Graceful shutdown
defer gq.Stop()
```

## Backends

### In-Memory Backend

Perfect for development, testing, and single-process applications.

```go
backend := goqueue.NewMemoryBackend(
    goqueue.WithBufferSize(200),
)
```

**Features:**
- Zero external dependencies
- Fast and lightweight
- Great for testing
- Single-process only

### Redis Backend

Production-ready distributed queue with Redis.

```go
backend := goqueue.NewRedisBackend("localhost:6379",
    goqueue.WithRedisPassword("your-password"),
    goqueue.WithRedisDB(0),
    goqueue.WithPollTimeout(5 * time.Second),
)

// Test connection
if err := backend.Ping(ctx); err != nil {
    log.Fatal(err)
}
```

**Features:**
- Distributed processing across multiple instances
- Persistent message storage
- Horizontal scaling
- Production-ready reliability

## Configuration

Configure GoQueue behavior with functional options:

| Option | Description | Default |
|--------|-------------|---------|
| `WithWorkerCount(n)` | Number of concurrent workers per queue | 1 |
| `WithRetryCount(n)` | Maximum retry attempts for failed messages | 3 |
| `WithRetryDelay(d)` | Initial delay between retries (exponential backoff) | 1 second |
| `WithDLQ(name)` | Enable dead letter queue for failed messages | Disabled |

Example:

```go
gq := goqueue.New(backend,
    goqueue.WithWorkerCount(10),
    goqueue.WithRetryCount(5),
    goqueue.WithRetryDelay(2 * time.Second),
    goqueue.WithDLQ("failed-jobs"),
)
```

## Examples

See the [`examples/`](examples/) directory for complete working examples:

- **[Basic](examples/basic/main.go)** - Simple pub/sub with in-memory backend
- **[Advanced](examples/advanced/main.go)** - Retry, DLQ, concurrent workers
- **[Redis](examples/redis/main.go)** - Redis backend usage
- **[Distributed](examples/redis-distributed/)** - Production architecture with separate publisher and worker services

### Running Examples

```bash
# In-memory example
go run examples/basic/main.go

# Redis example (requires Redis)
REDIS_ADDR=localhost:6379 go run examples/redis/main.go

# With password
REDIS_ADDR=localhost:6379 REDIS_PASSWORD=secret go run examples/redis/main.go

# Distributed example (run in separate terminals)
cd examples/redis-distributed
go run worker/main.go      # Terminal 1
go run publisher/main.go   # Terminal 2
```

## Production Architecture

GoQueue supports distributed architectures where publishers and workers run as separate services:

```
┌─────────────┐         ┌────────┐         ┌─────────────┐
│  Publisher  │────────▶│ Redis  │◀────────│   Worker    │
│  (API)      │         │ Queue  │         │  (Process)  │
└─────────────┘         └────────┘         └─────────────┘
                             ▲
                             │
                        ┌────┴────┐
                        │  Worker │
                        │ (Scale) │
                        └─────────┘
```

See [examples/redis-distributed](examples/redis-distributed/) for a complete implementation.

## Documentation

### Core Interfaces

#### QueueMessage

All messages must implement this interface:

```go
type QueueMessage interface {
    QueueName() string
}
```

#### Handler

All handlers must implement this interface:

```go
type Handler interface {
    QueueName() string
    Handle(ctx context.Context, envelope *Envelope) error
}
```

#### Backend

Custom backends must implement:

```go
type Backend interface {
    Publish(ctx context.Context, queue string, envelope *Envelope) error
    Subscribe(ctx context.Context, queue string) (<-chan *Envelope, error)
    Ack(ctx context.Context, messageID string) error
    Nack(ctx context.Context, messageID string) error
    Close() error
}
```

### Envelope Structure

Internal message wrapper with metadata:

```go
type Envelope struct {
    ID          string
    Queue       string
    Data        []byte
    RetryCount  int
    CreatedAt   time.Time
    ProcessedAt *time.Time
    Metadata    map[string]string
}
```

## Error Handling and Retries

When a handler returns an error, GoQueue automatically:

1. Increments the retry count
2. Applies exponential backoff: `initial_delay * 2^(retry_count-1)`
3. Re-queues the message
4. After max retries, moves to Dead Letter Queue (if enabled)

Example retry timeline:
- 1st retry: 1 second
- 2nd retry: 2 seconds
- 3rd retry: 4 seconds
- After max retries: Move to DLQ

## Testing

GoQueue is designed to be easily testable:

```go
// Use in-memory backend for tests
func TestMyHandler(t *testing.T) {
    backend := goqueue.NewMemoryBackend()
    gq := goqueue.New(backend)

    handler := &MyHandler{}
    gq.Register(handler)
    gq.Start(context.Background())

    // Publish test message
    gq.Publish(context.Background(), &MyMessage{})

    // Assert expected behavior
}
```

## Environment Variables

Configure examples using environment variables:

```bash
# Redis connection
export REDIS_ADDR=localhost:6379
export REDIS_PASSWORD=your-password

# Run example
go run examples/redis/main.go
```

## Monitoring

### Redis Queue Monitoring

```bash
# Connect to Redis
redis-cli

# Check queue lengths
LLEN goqueue:queue:emails
LLEN goqueue:queue:notifications

# View pending messages
KEYS goqueue:pending:*

# Get message details
GET goqueue:pending:message-id
```

## Performance

- In-memory backend: 100,000+ messages/sec
- Redis backend: Depends on Redis and network latency
- Concurrent workers: Linear scaling up to CPU cores

## Roadmap

- [ ] RabbitMQ backend
- [ ] AWS SQS backend
- [ ] Kafka backend
- [ ] Message priority support
- [ ] Delayed/scheduled messages
- [ ] Metrics and monitoring (Prometheus)
- [ ] Message batching
- [ ] Web UI for queue monitoring

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development

```bash
# Clone repository
git clone https://github.com/openframebox/goqueue.git
cd goqueue

# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Run linter
golangci-lint run
```

## License

GoQueue is released under the [MIT License](LICENSE).

## Support

- 📫 **Issues**: [GitHub Issues](https://github.com/openframebox/goqueue/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/openframebox/goqueue/discussions)
- 📖 **Documentation**: [Wiki](https://github.com/openframebox/goqueue/wiki)

## Acknowledgments

Built with ❤️ by [OpenFrameBox](https://github.com/openframebox)

---

<p align="center">
  <sub>If you find GoQueue useful, please consider giving it a ⭐️</sub>
</p>
