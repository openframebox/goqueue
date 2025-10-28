package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/openframebox/goqueue"
)

// Job represents a job message to be processed
type Job struct {
	ID         int    `json:"id"`
	Operation  string `json:"operation"`
	ShouldFail bool   `json:"should_fail"`
}

// QueueName implements the goqueue.QueueMessage interface
func (j *Job) QueueName() string {
	return "jobs"
}

var (
	processedCount int32
	failedCount    int32
)

// JobHandler processes job messages
type JobHandler struct{}

// QueueName implements the goqueue.Handler interface
func (h *JobHandler) QueueName() string {
	return "jobs"
}

// Handle processes job messages
func (h *JobHandler) Handle(ctx context.Context, envelope *goqueue.Envelope) error {
	var job Job
	if err := envelope.Unmarshal(&job); err != nil {
		return fmt.Errorf("failed to unmarshal job: %w", err)
	}

	fmt.Printf("[Worker] Processing job %d (retry: %d): %s\n",
		job.ID, envelope.RetryCount, job.Operation)

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	// Simulate failure for certain jobs
	if job.ShouldFail {
		atomic.AddInt32(&failedCount, 1)
		return errors.New("job failed as expected")
	}

	atomic.AddInt32(&processedCount, 1)
	fmt.Printf("[Worker] ✓ Successfully completed job %d\n", job.ID)
	return nil
}

// DLQHandler processes messages from the dead letter queue
type DLQHandler struct{}

// QueueName implements the goqueue.Handler interface
func (h *DLQHandler) QueueName() string {
	return "failed-jobs"
}

// Handle processes messages that have permanently failed
func (h *DLQHandler) Handle(ctx context.Context, envelope *goqueue.Envelope) error {
	var job Job
	if err := envelope.Unmarshal(&job); err != nil {
		return fmt.Errorf("failed to unmarshal DLQ job: %w", err)
	}

	fmt.Printf("[DLQ] Received permanently failed job %d (original queue: %s)\n",
		job.ID, envelope.Metadata["original_queue"])
	return nil
}

func main() {
	fmt.Println("=== GoQueue Advanced Example ===")
	fmt.Println("Features: Concurrent workers, Retry mechanism, Dead letter queue")
	fmt.Println()

	// 1. Create an in-memory backend with larger buffer
	backend := goqueue.NewMemoryBackend(goqueue.WithBufferSize(200))

	// 2. Create a GoQueue instance with advanced options
	gq := goqueue.New(
		backend,
		goqueue.WithWorkerCount(3),                   // 3 concurrent workers
		goqueue.WithRetryCount(3),                    // Retry up to 3 times
		goqueue.WithRetryDelay(500*time.Millisecond), // Initial retry delay
		goqueue.WithDLQ("failed-jobs"),               // Dead letter queue for failed jobs
	)

	// 3. Register handlers (each handler declares its queue via QueueName())
	jobHandler := &JobHandler{}
	if err := gq.Register(jobHandler); err != nil {
		log.Fatalf("Failed to register job handler: %v", err)
	}

	dlqHandler := &DLQHandler{}
	if err := gq.Register(dlqHandler); err != nil {
		log.Fatalf("Failed to register DLQ handler: %v", err)
	}

	fmt.Println("Registered handlers:")
	fmt.Println("  - JobHandler for 'jobs' queue")
	fmt.Println("  - DLQHandler for 'failed-jobs' queue")
	fmt.Println()

	// 4. Start all registered workers
	ctx := context.Background()
	if err := gq.Start(ctx); err != nil {
		log.Fatalf("Failed to start workers: %v", err)
	}

	fmt.Println("All workers started (3 concurrent workers per queue)")
	fmt.Println()

	// 5. Publish a mix of successful and failing jobs (each job declares its queue)
	jobs := []*Job{
		{ID: 1, Operation: "Process payment", ShouldFail: false},
		{ID: 2, Operation: "Send email", ShouldFail: false},
		{ID: 3, Operation: "Update database", ShouldFail: true},  // Will fail
		{ID: 4, Operation: "Generate report", ShouldFail: false},
		{ID: 5, Operation: "Sync data", ShouldFail: false},
		{ID: 6, Operation: "Invalid operation", ShouldFail: true}, // Will fail
		{ID: 7, Operation: "Backup files", ShouldFail: false},
		{ID: 8, Operation: "Clean cache", ShouldFail: false},
	}

	fmt.Println("Publishing jobs...")
	for _, job := range jobs {
		if err := gq.Publish(ctx, job); err != nil {
			log.Printf("Failed to publish job %d: %v", job.ID, err)
			continue
		}

		status := "✓"
		if job.ShouldFail {
			status = "✗ (will fail)"
		}
		fmt.Printf("Published job %d: %s %s\n", job.ID, job.Operation, status)
	}

	// Wait for all messages to be processed
	// Note: With 3 retries and exponential backoff, failed messages take longer
	fmt.Println()
	fmt.Println("Waiting for jobs to be processed...")
	fmt.Println("(Failed jobs will be retried 3 times before moving to DLQ)")
	fmt.Println()

	// Monitor progress
	for range 15 {
		time.Sleep(1 * time.Second)
		processed := atomic.LoadInt32(&processedCount)
		failed := atomic.LoadInt32(&failedCount)
		fmt.Printf("[Status] Processed: %d, Failed attempts: %d\n", processed, failed)
	}

	// 6. Stop the queue
	fmt.Println()
	fmt.Println("Stopping queue...")
	if err := gq.Stop(); err != nil {
		log.Printf("Error stopping queue: %v", err)
	}

	// 7. Show final statistics
	fmt.Println()
	fmt.Println("=== Final Statistics ===")
	fmt.Printf("Successfully processed: %d jobs\n", atomic.LoadInt32(&processedCount))
	fmt.Printf("Total failed attempts: %d\n", atomic.LoadInt32(&failedCount))
	fmt.Println()
	fmt.Println("Done!")
}
