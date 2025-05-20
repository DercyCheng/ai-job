package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ai-job/internal/models"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Config represents database configuration
type Config struct {
	Driver          string
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// Database represents the database connection
type Database struct {
	db *sqlx.DB
}

// New creates a new database connection
func New(cfg Config) (*Database, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)

	db, err := sqlx.Connect(cfg.Driver, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return &Database{db: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// Ping checks database connectivity
func (d *Database) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

// Transaction begins a new transaction
func (d *Database) Transaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := d.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// TaskRepository handles database operations for tasks
type TaskRepository struct {
	db *Database
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db *Database) *TaskRepository {
	return &TaskRepository{db: db}
}

// Create creates a new task
func (r *TaskRepository) Create(ctx context.Context, task *models.Task) error {
	query := `
		INSERT INTO tasks (
			id, name, description, model_name, status, priority, input, 
			created_at, updated_at, user_id, timeout, retry_count, max_retries
		) VALUES (
			:id, :name, :description, :model_name, :status, :priority, :input, 
			:created_at, :updated_at, :user_id, :timeout, :retry_count, :max_retries
		)
	`

	_, err := r.db.db.NamedExecContext(ctx, query, task)
	return err
}

// GetByID retrieves a task by ID
func (r *TaskRepository) GetByID(ctx context.Context, id string) (*models.Task, error) {
	var task models.Task
	query := `SELECT * FROM tasks WHERE id = $1`
	err := r.db.db.GetContext(ctx, &task, query, id)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// Update updates a task
func (r *TaskRepository) Update(ctx context.Context, task *models.Task) error {
	task.UpdatedAt = time.Now()
	query := `
		UPDATE tasks SET
			name = :name,
			description = :description,
			model_name = :model_name,
			status = :status,
			priority = :priority,
			input = :input,
			output = :output,
			error = :error,
			updated_at = :updated_at,
			started_at = :started_at,
			completed_at = :completed_at,
			worker_id = :worker_id,
			retry_count = :retry_count
		WHERE id = :id
	`

	_, err := r.db.db.NamedExecContext(ctx, query, task)
	return err
}

// List retrieves a list of tasks with filtering options
func (r *TaskRepository) List(ctx context.Context, status *models.TaskStatus, limit, offset int) ([]*models.Task, error) {
	var tasks []*models.Task
	var query string
	var args []interface{}

	if status != nil {
		query = `SELECT * FROM tasks WHERE status = $1 ORDER BY priority DESC, created_at ASC LIMIT $2 OFFSET $3`
		args = []interface{}{*status, limit, offset}
	} else {
		query = `SELECT * FROM tasks ORDER BY priority DESC, created_at ASC LIMIT $1 OFFSET $2`
		args = []interface{}{limit, offset}
	}

	err := r.db.db.SelectContext(ctx, &tasks, query, args...)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

// GetPendingTasks retrieves pending tasks ordered by priority and creation time
func (r *TaskRepository) GetPendingTasks(ctx context.Context, limit int) ([]*models.Task, error) {
	var tasks []*models.Task
	status := models.TaskStatusPending

	query := `
		SELECT * FROM tasks 
		WHERE status = $1 
		ORDER BY priority DESC, created_at ASC 
		LIMIT $2
	`

	err := r.db.db.SelectContext(ctx, &tasks, query, status, limit)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

// WorkerRepository handles database operations for workers
type WorkerRepository struct {
	db *Database
}

// NewWorkerRepository creates a new worker repository
func NewWorkerRepository(db *Database) *WorkerRepository {
	return &WorkerRepository{db: db}
}

// Create creates a new worker
func (r *WorkerRepository) Create(ctx context.Context, worker *models.Worker) error {
	query := `
		INSERT INTO workers (
			id, name, status, capabilities, last_heartbeat, registered_at,
			available_memory, available_cpu, available_gpu, total_tasks_handled
		) VALUES (
			:id, :name, :status, :capabilities, :last_heartbeat, :registered_at,
			:available_memory, :available_cpu, :available_gpu, :total_tasks_handled
		)
	`

	_, err := r.db.db.NamedExecContext(ctx, query, worker)
	return err
}

// GetByID retrieves a worker by ID
func (r *WorkerRepository) GetByID(ctx context.Context, id string) (*models.Worker, error) {
	var worker models.Worker
	query := `SELECT * FROM workers WHERE id = $1`
	err := r.db.db.GetContext(ctx, &worker, query, id)
	if err != nil {
		return nil, err
	}
	return &worker, nil
}

// Update updates a worker
func (r *WorkerRepository) Update(ctx context.Context, worker *models.Worker) error {
	query := `
		UPDATE workers SET
			name = :name,
			status = :status,
			capabilities = :capabilities,
			current_task_id = :current_task_id,
			last_heartbeat = :last_heartbeat,
			available_memory = :available_memory,
			available_cpu = :available_cpu,
			available_gpu = :available_gpu,
			total_tasks_handled = :total_tasks_handled
		WHERE id = :id
	`

	_, err := r.db.db.NamedExecContext(ctx, query, worker)
	return err
}

// UpdateHeartbeat updates a worker's heartbeat timestamp
func (r *WorkerRepository) UpdateHeartbeat(ctx context.Context, id string) error {
	query := `UPDATE workers SET last_heartbeat = $1 WHERE id = $2`
	_, err := r.db.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

// ListAvailable retrieves a list of available workers
func (r *WorkerRepository) ListAvailable(ctx context.Context) ([]*models.Worker, error) {
	var workers []*models.Worker
	// Find workers that are available (have no current task) and have checked in recently
	query := `
		SELECT * FROM workers 
		WHERE status = 'available' AND current_task_id IS NULL
		AND last_heartbeat > $1
		ORDER BY available_gpu DESC, available_memory DESC
	`

	// Consider workers that have checked in within the last minute
	heartbeatThreshold := time.Now().Add(-1 * time.Minute)

	err := r.db.db.SelectContext(ctx, &workers, query, heartbeatThreshold)
	if err != nil {
		return nil, err
	}

	return workers, nil
}
