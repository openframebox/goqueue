package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/openframebox/goqueue"
	"github.com/openframebox/goqueue/examples/redis-distributed/shared"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// EmailHandler processes email jobs
type EmailHandler struct{}

func (h *EmailHandler) QueueName() string {
	return "email-jobs"
}

func (h *EmailHandler) Handle(ctx context.Context, envelope *goqueue.Envelope) error {
	var job shared.EmailJob
	if err := envelope.Unmarshal(&job); err != nil {
		return fmt.Errorf("failed to unmarshal email job: %w", err)
	}

	// Simulate email sending
	fmt.Printf("[EmailWorker] 📧 Sending email to %s: %s\n", job.To, job.Subject)
	time.Sleep(500 * time.Millisecond) // Simulate work

	fmt.Printf("[EmailWorker] ✓ Email #%d sent successfully\n", job.ID)
	return nil
}

// ImageHandler processes image processing jobs
type ImageHandler struct{}

func (h *ImageHandler) QueueName() string {
	return "image-jobs"
}

func (h *ImageHandler) Handle(ctx context.Context, envelope *goqueue.Envelope) error {
	var job shared.ImageProcessingJob
	if err := envelope.Unmarshal(&job); err != nil {
		return fmt.Errorf("failed to unmarshal image job: %w", err)
	}

	// Simulate image processing
	fmt.Printf("[ImageWorker] 🖼️  Processing %s: %s (%dx%d)\n",
		job.ImageURL, job.Operation, job.Width, job.Height)
	time.Sleep(1 * time.Second) // Simulate work

	fmt.Printf("[ImageWorker] ✓ Image #%d processed successfully\n", job.ID)
	return nil
}

// NotificationHandler processes notification jobs
type NotificationHandler struct{}

func (h *NotificationHandler) QueueName() string {
	return "notification-jobs"
}

func (h *NotificationHandler) Handle(ctx context.Context, envelope *goqueue.Envelope) error {
	var job shared.NotificationJob
	if err := envelope.Unmarshal(&job); err != nil {
		return fmt.Errorf("failed to unmarshal notification job: %w", err)
	}

	// Simulate sending notification
	fmt.Printf("[NotificationWorker] 🔔 Sending %s notification to user %d: %s\n",
		job.Type, job.UserID, job.Message)
	time.Sleep(300 * time.Millisecond) // Simulate work

	fmt.Printf("[NotificationWorker] ✓ Notification #%d sent successfully\n", job.ID)
	return nil
}

func main() {
	fmt.Println("=== Worker Service ===")
	fmt.Println("This service consumes and processes jobs from Redis queues")
	fmt.Println()

	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")

	var backend *goqueue.RedisBackend
	if redisPassword != "" {
		backend = goqueue.NewRedisBackend(redisAddr,
			goqueue.WithRedisPassword(redisPassword))
	} else {
		backend = goqueue.NewRedisBackend(redisAddr)
	}

	ctx := context.Background()
	if err := backend.Ping(ctx); err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v\n", err)
	}
	fmt.Printf("✓ Connected to Redis at %s\n", redisAddr)

	// 2. Create GoQueue instance with concurrent workers
	gq := goqueue.New(backend,
		goqueue.WithWorkerCount(3),    // 3 concurrent workers per queue
		goqueue.WithRetryCount(3),     // Retry failed jobs 3 times
		goqueue.WithDLQ("failed-jobs"), // Dead letter queue for failed jobs
	)

	// 3. Register handlers for different job types
	if err := gq.Register(&EmailHandler{}); err != nil {
		log.Fatalf("Failed to register email handler: %v", err)
	}
	fmt.Println("✓ Registered EmailHandler for 'email-jobs' queue")

	if err := gq.Register(&ImageHandler{}); err != nil {
		log.Fatalf("Failed to register image handler: %v", err)
	}
	fmt.Println("✓ Registered ImageHandler for 'image-jobs' queue")

	if err := gq.Register(&NotificationHandler{}); err != nil {
		log.Fatalf("Failed to register notification handler: %v", err)
	}
	fmt.Println("✓ Registered NotificationHandler for 'notification-jobs' queue")

	// 4. Start all workers
	if err := gq.Start(ctx); err != nil {
		log.Fatalf("Failed to start workers: %v", err)
	}

	fmt.Println()
	fmt.Println("✓ All workers started (3 concurrent workers per queue)")
	fmt.Println("✓ Listening for jobs... (Ctrl+C to stop)")
	fmt.Println()

	// 5. Wait for interrupt signal for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	// 6. Graceful shutdown
	fmt.Println()
	fmt.Println("Shutting down workers...")
	if err := gq.Stop(); err != nil {
		log.Printf("Error stopping queue: %v", err)
	}

	fmt.Println("Worker service stopped")
}
