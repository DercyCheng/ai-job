package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ai-job/internal/models"
)

// QueueDriver defines the interface for queue drivers
type QueueDriver interface {
	Push(ctx context.Context, task *models.Task) error
	Pop(ctx context.Context, timeout time.Duration) (*models.Task, error)
	Delete(ctx context.Context, taskID string) error
	Size(ctx context.Context) (int, error)
	Close() error
}

// Config represents queue configuration
type Config struct {
	Driver   string
	Address  string
	Password string
	MaxRetry int
	JobTTL   time.Duration
}

// Queue manages task queues
type Queue struct {
	driver QueueDriver
	config Config
}

// New creates a new queue
func New(config Config) (*Queue, error) {
	var driver QueueDriver
	var err error

	switch config.Driver {
	case "memory":
		driver = newMemoryQueue()
	case "redis":
		driver, err = newRedisQueue(config.Address, config.Password)
	default:
		return nil, fmt.Errorf("unsupported queue driver: %s", config.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize queue driver: %w", err)
	}

	return &Queue{
		driver: driver,
		config: config,
	}, nil
}

// Push adds a task to the queue
func (q *Queue) Push(ctx context.Context, task *models.Task) error {
	return q.driver.Push(ctx, task)
}

// Pop retrieves and removes a task from the queue
func (q *Queue) Pop(ctx context.Context, timeout time.Duration) (*models.Task, error) {
	return q.driver.Pop(ctx, timeout)
}

// Delete removes a task from the queue
func (q *Queue) Delete(ctx context.Context, taskID string) error {
	return q.driver.Delete(ctx, taskID)
}

// Size returns the number of tasks in the queue
func (q *Queue) Size(ctx context.Context) (int, error) {
	return q.driver.Size(ctx)
}

// Close closes the queue
func (q *Queue) Close() error {
	return q.driver.Close()
}

// MemoryQueue is an in-memory implementation of QueueDriver
type MemoryQueue struct {
	tasks []models.Task
}

// newMemoryQueue creates a new memory queue
func newMemoryQueue() *MemoryQueue {
	return &MemoryQueue{
		tasks: make([]models.Task, 0),
	}
}

// Push adds a task to the memory queue
func (q *MemoryQueue) Push(ctx context.Context, task *models.Task) error {
	q.tasks = append(q.tasks, *task)
	return nil
}

// Pop retrieves and removes a task from the memory queue
func (q *MemoryQueue) Pop(ctx context.Context, timeout time.Duration) (*models.Task, error) {
	if len(q.tasks) == 0 {
		return nil, errors.New("queue is empty")
	}

	task := q.tasks[0]
	q.tasks = q.tasks[1:]
	return &task, nil
}

// Delete removes a task from the memory queue
func (q *MemoryQueue) Delete(ctx context.Context, taskID string) error {
	for i, task := range q.tasks {
		if task.ID == taskID {
			q.tasks = append(q.tasks[:i], q.tasks[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("task not found: %s", taskID)
}

// Size returns the number of tasks in the memory queue
func (q *MemoryQueue) Size(ctx context.Context) (int, error) {
	return len(q.tasks), nil
}

// Close closes the memory queue
func (q *MemoryQueue) Close() error {
	q.tasks = nil
	return nil
}

// RedisQueue is a Redis implementation of QueueDriver
// Note: This is a simplified version. A real implementation would use a Redis client.
type RedisQueue struct {
	address  string
	password string
}

// newRedisQueue creates a new Redis queue
func newRedisQueue(address, password string) (*RedisQueue, error) {
	// In a real implementation, this would initialize a Redis client
	return &RedisQueue{
		address:  address,
		password: password,
	}, nil
}

// Push adds a task to the Redis queue
func (q *RedisQueue) Push(ctx context.Context, task *models.Task) error {
	// This is a placeholder. In a real implementation, this would serialize and store the task in Redis.
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to serialize task: %w", err)
	}
	_ = taskJSON // Use this in a real implementation

	return nil
}

// Pop retrieves and removes a task from the Redis queue
func (q *RedisQueue) Pop(ctx context.Context, timeout time.Duration) (*models.Task, error) {
	// This is a placeholder. In a real implementation, this would retrieve and deserialize a task from Redis.
	return nil, errors.New("not implemented")
}

// Delete removes a task from the Redis queue
func (q *RedisQueue) Delete(ctx context.Context, taskID string) error {
	// This is a placeholder. In a real implementation, this would remove a task from Redis.
	return nil
}

// Size returns the number of tasks in the Redis queue
func (q *RedisQueue) Size(ctx context.Context) (int, error) {
	// This is a placeholder. In a real implementation, this would count tasks in Redis.
	return 0, nil
}

// Close closes the Redis queue
func (q *RedisQueue) Close() error {
	// This is a placeholder. In a real implementation, this would close the Redis client.
	return nil
}
