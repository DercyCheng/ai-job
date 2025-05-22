package models

import (
	"time"

	"github.com/google/uuid"
)

// TaskStatus represents the current status of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusScheduled TaskStatus = "scheduled"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// TaskPriority represents the priority level of a task
type TaskPriority int

const (
	TaskPriorityLow      TaskPriority = 1
	TaskPriorityNormal   TaskPriority = 2
	TaskPriorityHigh     TaskPriority = 3
	TaskPriorityCritical TaskPriority = 4
)

// Task represents an LLM task to be processed
type Task struct {
	ID          string       `json:"id" db:"id"`
	Name        string       `json:"name" db:"name"`
	Description string       `json:"description" db:"description"`
	ModelName   string       `json:"model_name" db:"model_name"`
	Status      TaskStatus   `json:"status" db:"status"`
	Priority    TaskPriority `json:"priority" db:"priority"`
	Input       []byte       `json:"input" db:"input"`
	Output      []byte       `json:"output" db:"output"`
	Error       string       `json:"error" db:"error"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
	StartedAt   *time.Time   `json:"started_at" db:"started_at"`
	CompletedAt *time.Time   `json:"completed_at" db:"completed_at"`
	WorkerID    *string      `json:"worker_id" db:"worker_id"`
	UserID      string       `json:"user_id" db:"user_id"`
	Timeout     int          `json:"timeout" db:"timeout"`
	RetryCount  int          `json:"retry_count" db:"retry_count"`
	MaxRetries  int          `json:"max_retries" db:"max_retries"`
}

// NewTask creates a new task with default values
func NewTask(name, modelName, userID string, priority TaskPriority, input []byte) *Task {
	return &Task{
		ID:         uuid.New().String(),
		Name:       name,
		ModelName:  modelName,
		Status:     TaskStatusPending,
		Priority:   priority,
		Input:      input,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		UserID:     userID,
		Timeout:    1800, // Default timeout of 30 minutes
		RetryCount: 0,
		MaxRetries: 3,
	}
}

// Worker represents a worker node in the system
type Worker struct {
	ID                string    `json:"id" db:"id"`
	Name              string    `json:"name" db:"name"`
	Status            string    `json:"status" db:"status"`
	Capabilities      []string  `json:"capabilities" db:"capabilities"`
	CurrentTaskID     *string   `json:"current_task_id" db:"current_task_id"`
	LastHeartbeat     time.Time `json:"last_heartbeat" db:"last_heartbeat"`
	RegisteredAt      time.Time `json:"registered_at" db:"registered_at"`
	AvailableMemory   int64     `json:"available_memory" db:"available_memory"`
	AvailableCPU      float64   `json:"available_cpu" db:"available_cpu"`
	AvailableGPU      float64   `json:"available_gpu" db:"available_gpu"`
	TotalTasksHandled int       `json:"total_tasks_handled" db:"total_tasks_handled"`
}

// NewWorker creates a new worker with default values
func NewWorker(name string, capabilities []string) *Worker {
	return &Worker{
		ID:                uuid.New().String(),
		Name:              name,
		Status:            "available",
		Capabilities:      capabilities,
		LastHeartbeat:     time.Now(),
		RegisteredAt:      time.Now(),
		TotalTasksHandled: 0,
	}
}

// Model represents an LLM model configuration
type Model struct {
	Name             string `json:"name" db:"name"`
	Provider         string `json:"provider" db:"provider"`
	ModelPath        string `json:"model_path" db:"model_path"`
	MaxContextLength int    `json:"max_context_length" db:"max_context_length"`
	Quantization     string `json:"quantization" db:"quantization"`
	RequiredMemory   int64  `json:"required_memory" db:"required_memory"`
	RequiresGPU      bool   `json:"requires_gpu" db:"requires_gpu"`
}
