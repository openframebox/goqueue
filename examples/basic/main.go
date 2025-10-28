package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/openframebox/goqueue"
)

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
	fmt.Println("=== GoQueue Basic Example ===")
	fmt.Println()

	// 1. Create an in-memory backend
	backend := goqueue.NewMemoryBackend()

	// 2. Create a GoQueue instance
	gq := goqueue.New(backend)

	// 3. Register the handler (handler declares its queue via QueueName())
	handler := &TaskHandler{}
	if err := gq.Register(handler); err != nil {
		log.Fatalf("Failed to register handler: %v", err)
	}

	fmt.Println("Handler registered for 'tasks' queue")

	// 4. Start all registered workers
	ctx := context.Background()
	if err := gq.Start(ctx); err != nil {
		log.Fatalf("Failed to start workers: %v", err)
	}

	fmt.Println("Workers started. Publishing tasks...")
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
	time.Sleep(2 * time.Second)

	// 6. Stop the queue
	fmt.Println()
	fmt.Println("Stopping queue...")
	if err := gq.Stop(); err != nil {
		log.Printf("Error stopping queue: %v", err)
	}

	fmt.Println("Done!")
}
