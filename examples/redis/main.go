package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/openframebox/goqueue"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Task represents a simple task message
type Task struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// QueueName implements the goqueue.QueueMessage interface
func (t *Task) QueueName() string {
	return "tasks"
}

// TaskHandler processes task messages
type TaskHandler struct{}

// QueueName implements the goqueue.Handler interface
func (h *TaskHandler) QueueName() string {
	return "tasks"
}

// Handle processes task messages
func (h *TaskHandler) Handle(ctx context.Context, envelope *goqueue.Envelope) error {
	var task Task
	if err := envelope.Unmarshal(&task); err != nil {
		return fmt.Errorf("failed to unmarshal task: %w", err)
	}

	fmt.Printf("Processing task: ID=%d, Name=%s\n", task.ID, task.Name)
	return nil
}

func main() {
	fmt.Println("=== GoQueue Redis Example ===")
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
		log.Fatalf("Failed to connect to Redis: %v\n", err)
	}
	fmt.Printf("✓ Connected to Redis at %s\n", redisAddr)

	// 2. Create a GoQueue instance
	gq := goqueue.New(backend)

	// 3. Register the handler (handler declares its queue via QueueName())
	handler := &TaskHandler{}
	if err := gq.Register(handler); err != nil {
		log.Fatalf("Failed to register handler: %v", err)
	}

	fmt.Println("✓ Handler registered for 'tasks' queue")

	// 4. Start all registered workers
	if err := gq.Start(ctx); err != nil {
		log.Fatalf("Failed to start workers: %v", err)
	}

	fmt.Println("✓ Workers started. Publishing tasks...")
	fmt.Println()

	// 5. Publish some tasks (each task declares its queue via QueueName())
	for i := 1; i <= 5; i++ {
		task := &Task{
			ID:   i,
			Name: fmt.Sprintf("Task #%d", i),
		}

		if err := gq.Publish(ctx, task); err != nil {
			log.Printf("Failed to publish task %d: %v", i, err)
			continue
		}

		fmt.Printf("Published: %s\n", task.Name)
	}

	// Wait a bit for messages to be processed
	fmt.Println()
	fmt.Println("Waiting for tasks to be processed...")
	time.Sleep(3 * time.Second)

	// 6. Stop the queue
	fmt.Println()
	fmt.Println("Stopping queue...")
	if err := gq.Stop(); err != nil {
		log.Printf("Error stopping queue: %v", err)
	}

	fmt.Println("Done!")
}
