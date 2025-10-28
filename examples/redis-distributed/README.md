# Redis Distributed Queue Example

This example demonstrates how to use GoQueue with Redis in a distributed architecture where the **publisher** (main service) and **worker** (background processor) run as separate services.

## Architecture

```
┌─────────────────┐         ┌──────────┐         ┌─────────────────┐
│  Publisher      │────────▶│  Redis   │◀────────│  Worker         │
│  Service        │ Publish │  Queues  │ Consume │  Service        │
│  (API/Main)     │         │          │         │  (Background)   │
└─────────────────┘         └──────────┘         └─────────────────┘
```

## Components

### 1. **Shared** (`shared/messages.go`)
- Common message types used by both publisher and worker
- Implements `QueueMessage` interface
- Three job types:
  - `EmailJob` - Email sending tasks
  - `ImageProcessingJob` - Image processing tasks
  - `NotificationJob` - Push notification tasks

### 2. **Publisher** (`publisher/main.go`)
- Main service that creates and publishes jobs
- Only publishes, does not process
- Simulates continuous job creation every 2 seconds
- Can run multiple instances for high-throughput

### 3. **Worker** (`worker/main.go`)
- Background service that processes jobs
- Registers handlers for each job type
- Runs multiple concurrent workers per queue
- Can scale horizontally by running multiple instances

## Prerequisites

1. **Redis Server**

```bash
# Install Redis (macOS)
brew install redis

# Start Redis
redis-server

# Or with password:
redis-server --requirepass "@passWORD1"
```

2. **Go Dependencies**

```bash
cd examples/redis-distributed
go mod tidy
```

## Running the Example

### Terminal 1: Start the Worker Service

```bash
cd examples/redis-distributed/worker
go run main.go
```

**Output:**
```
=== Worker Service ===
This service consumes and processes jobs from Redis queues

✓ Connected to Redis
✓ Registered EmailHandler for 'email-jobs' queue
✓ Registered ImageHandler for 'image-jobs' queue
✓ Registered NotificationHandler for 'notification-jobs' queue

✓ All workers started (3 concurrent workers per queue)
✓ Listening for jobs... (Ctrl+C to stop)
```

### Terminal 2: Start the Publisher Service

```bash
cd examples/redis-distributed/publisher
go run main.go
```

**Output:**
```
=== Publisher Service ===
This service publishes jobs to Redis queues

✓ Connected to Redis
✓ Publisher ready

Publishing jobs every 2 seconds (Ctrl+C to stop)...

📧 Published EmailJob #1 to queue 'email-jobs'
🖼️  Published ImageProcessingJob #2 to queue 'image-jobs'
🔔 Published NotificationJob #3 to queue 'notification-jobs'
...
```

### You'll see the Worker processing jobs:

```
[EmailWorker] 📧 Sending email to user1@example.com: Welcome Email #1
[EmailWorker] ✓ Email #1 sent successfully
[ImageWorker] 🖼️  Processing https://example.com/image2.jpg: resize (800x600)
[ImageWorker] ✓ Image #2 processed successfully
[NotificationWorker] 🔔 Sending push notification to user 1003: You have a new message #3
[NotificationWorker] ✓ Notification #3 sent successfully
```

## Scaling

### Scale Workers Horizontally

Run multiple worker instances in separate terminals:

```bash
# Terminal 1
cd examples/redis-distributed/worker
go run main.go

# Terminal 2
cd examples/redis-distributed/worker
go run main.go

# Terminal 3
cd examples/redis-distributed/worker
go run main.go
```

Each worker instance will compete for jobs from the same Redis queues, automatically distributing the workload.

### Scale Publishers

You can also run multiple publisher instances to generate more load:

```bash
# Terminal 4
cd examples/redis-distributed/publisher
go run main.go

# Terminal 5
cd examples/redis-distributed/publisher
go run main.go
```

## Configuration

### Redis Connection

Edit the Redis connection in both services:

```go
// With password
backend := goqueue.NewRedisBackend("localhost:6379",
    goqueue.WithRedisPassword("your-password"))

// With password and database
backend := goqueue.NewRedisBackend("localhost:6379",
    goqueue.WithRedisPassword("your-password"),
    goqueue.WithRedisDB(1))
```

### Worker Configuration

Adjust worker settings in `worker/main.go`:

```go
gq := goqueue.New(backend,
    goqueue.WithWorkerCount(5),     // More concurrent workers
    goqueue.WithRetryCount(3),      // Retry attempts
    goqueue.WithDLQ("failed-jobs"), // Dead letter queue
)
```

## Real-World Use Cases

This architecture is perfect for:

1. **API + Background Jobs**
   - API server publishes jobs
   - Separate worker processes handle time-consuming tasks

2. **Microservices**
   - Service A publishes events
   - Service B consumes and processes them

3. **Scheduled Tasks**
   - Cron job publishes tasks
   - Worker pool processes them concurrently

4. **High Availability**
   - Multiple workers ensure fault tolerance
   - Redis persistence ensures no job loss

## Monitoring

### Check Queue Sizes in Redis

```bash
redis-cli

# Check queue lengths
LLEN goqueue:queue:email-jobs
LLEN goqueue:queue:image-jobs
LLEN goqueue:queue:notification-jobs

# View pending messages
KEYS goqueue:pending:*
```

## Graceful Shutdown

Both services handle `Ctrl+C` gracefully:

- **Publisher**: Stops publishing and closes Redis connection
- **Worker**: Finishes current jobs, then shuts down

## Next Steps

- Add health check endpoints
- Implement metrics/monitoring (Prometheus)
- Add structured logging
- Deploy with Docker/Kubernetes
- Use Redis Sentinel or Cluster for HA
