package main

import (
	"context"
	"fmt"
	"log"
	"os"
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

func main() {
	fmt.Println("=== Publisher Service ===")
	fmt.Println("This service publishes jobs to Redis queues")
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

	// 2. Create GoQueue instance (no need to register handlers or start workers)
	gq := goqueue.New(backend)

	fmt.Println("✓ Publisher ready")
	fmt.Println()

	// 3. Simulate publishing jobs continuously
	fmt.Println("Publishing jobs every 2 seconds (Ctrl+C to stop)...")
	fmt.Println()

	jobID := 1

	// Create a ticker for periodic publishing
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Publish different types of jobs
			switch jobID % 3 {
			case 0:
				// Email job
				job := &shared.EmailJob{
					ID:       jobID,
					To:       fmt.Sprintf("user%d@example.com", jobID),
					Subject:  fmt.Sprintf("Welcome Email #%d", jobID),
					Body:     "Thank you for signing up!",
					Priority: "normal",
				}
				if err := gq.Publish(ctx, job); err != nil {
					log.Printf("❌ Failed to publish email job: %v", err)
				} else {
					fmt.Printf("📧 Published EmailJob #%d to queue '%s'\n", job.ID, job.QueueName())
				}

			case 1:
				// Image processing job
				job := &shared.ImageProcessingJob{
					ID:        jobID,
					ImageURL:  fmt.Sprintf("https://example.com/image%d.jpg", jobID),
					Operation: "resize",
					Width:     800,
					Height:    600,
				}
				if err := gq.Publish(ctx, job); err != nil {
					log.Printf("❌ Failed to publish image job: %v", err)
				} else {
					fmt.Printf("🖼️  Published ImageProcessingJob #%d to queue '%s'\n", job.ID, job.QueueName())
				}

			case 2:
				// Notification job
				job := &shared.NotificationJob{
					ID:      jobID,
					UserID:  1000 + jobID,
					Message: fmt.Sprintf("You have a new message #%d", jobID),
					Type:    "push",
				}
				if err := gq.Publish(ctx, job); err != nil {
					log.Printf("❌ Failed to publish notification job: %v", err)
				} else {
					fmt.Printf("🔔 Published NotificationJob #%d to queue '%s'\n", job.ID, job.QueueName())
				}
			}

			jobID++
		}
	}
}
