package database

import (
	"context"
	"time"

	"ai-job/internal/models"
)

// MCPTaskRepository handles database operations for MCP tasks
type MCPTaskRepository struct {
	db *Database
}

// NewMCPTaskRepository creates a new MCP task repository
func NewMCPTaskRepository(db *Database) *MCPTaskRepository {
	return &MCPTaskRepository{db: db}
}

// Create creates a new MCP task
func (r *MCPTaskRepository) Create(ctx context.Context, task *models.MCPTask) error {
	query := `
		INSERT INTO mcp_tasks (
			id, name, description, model_id, context_id, type, status, priority, input, 
			created_at, updated_at, user_id, timeout, retry_count, max_retries
		) VALUES (
			:id, :name, :description, :model_id, :context_id, :type, :status, :priority, :input, 
			:created_at, :updated_at, :user_id, :timeout, :retry_count, :max_retries
		)
	`

	_, err := r.db.db.NamedExecContext(ctx, query, task)
	return err
}

// GetByID retrieves an MCP task by ID
func (r *MCPTaskRepository) GetByID(ctx context.Context, id string) (*models.MCPTask, error) {
	var task models.MCPTask
	query := `SELECT * FROM mcp_tasks WHERE id = $1`
	err := r.db.db.GetContext(ctx, &task, query, id)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetByContextID retrieves MCP tasks by context ID
func (r *MCPTaskRepository) GetByContextID(ctx context.Context, contextID string) ([]*models.MCPTask, error) {
	var tasks []*models.MCPTask
	query := `SELECT * FROM mcp_tasks WHERE context_id = $1 ORDER BY created_at ASC`
	err := r.db.db.SelectContext(ctx, &tasks, query, contextID)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// Update updates an MCP task
func (r *MCPTaskRepository) Update(ctx context.Context, task *models.MCPTask) error {
	task.UpdatedAt = time.Now()
	query := `
		UPDATE mcp_tasks SET
			name = :name,
			description = :description,
			model_id = :model_id,
			context_id = :context_id,
			type = :type,
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

// List retrieves a list of MCP tasks with filtering options
func (r *MCPTaskRepository) List(ctx context.Context, status *models.TaskStatus, taskType *models.MCPTaskType, limit, offset int) ([]*models.MCPTask, error) {
	var tasks []*models.MCPTask
	var query string
	var args []interface{}

	if status != nil && taskType != nil {
		query = `
			SELECT * FROM mcp_tasks 
			WHERE status = $1 AND type = $2 
			ORDER BY priority DESC, created_at ASC 
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{*status, *taskType, limit, offset}
	} else if status != nil {
		query = `
			SELECT * FROM mcp_tasks 
			WHERE status = $1 
			ORDER BY priority DESC, created_at ASC 
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{*status, limit, offset}
	} else if taskType != nil {
		query = `
			SELECT * FROM mcp_tasks 
			WHERE type = $1 
			ORDER BY priority DESC, created_at ASC 
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{*taskType, limit, offset}
	} else {
		query = `
			SELECT * FROM mcp_tasks 
			ORDER BY priority DESC, created_at ASC 
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	err := r.db.db.SelectContext(ctx, &tasks, query, args...)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

// GetPendingTasks retrieves pending MCP tasks ordered by priority and creation time
func (r *MCPTaskRepository) GetPendingTasks(ctx context.Context, limit int) ([]*models.MCPTask, error) {
	var tasks []*models.MCPTask
	status := models.TaskStatusPending

	query := `
		SELECT * FROM mcp_tasks 
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

// MCPContextRepository handles database operations for MCP contexts
type MCPContextRepository struct {
	db *Database
}

// MCPContext represents an MCP context stored in the database
type MCPContext struct {
	ID        string    `db:"id"`
	ModelID   string    `db:"model_id"`
	UserID    string    `db:"user_id"`
	Data      []byte    `db:"data"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// NewMCPContextRepository creates a new MCP context repository
func NewMCPContextRepository(db *Database) *MCPContextRepository {
	return &MCPContextRepository{db: db}
}

// Store stores an MCP context
func (r *MCPContextRepository) Store(ctx context.Context, contextID, modelID, userID string, data []byte) error {
	query := `
		INSERT INTO mcp_contexts (id, model_id, user_id, data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) 
		DO UPDATE SET data = $4, updated_at = $6
	`

	now := time.Now()
	_, err := r.db.db.ExecContext(ctx, query, contextID, modelID, userID, data, now, now)
	return err
}

// Get retrieves an MCP context by ID
func (r *MCPContextRepository) Get(ctx context.Context, contextID string) (*MCPContext, error) {
	var context MCPContext
	query := `SELECT * FROM mcp_contexts WHERE id = $1`
	err := r.db.db.GetContext(ctx, &context, query, contextID)
	if err != nil {
		return nil, err
	}
	return &context, nil
}

// List retrieves a list of MCP contexts
func (r *MCPContextRepository) List(ctx context.Context, userID *string, limit, offset int) ([]*MCPContext, error) {
	var contexts []*MCPContext
	var query string
	var args []interface{}

	if userID != nil {
		query = `SELECT * FROM mcp_contexts WHERE user_id = $1 ORDER BY updated_at DESC LIMIT $2 OFFSET $3`
		args = []interface{}{*userID, limit, offset}
	} else {
		query = `SELECT * FROM mcp_contexts ORDER BY updated_at DESC LIMIT $1 OFFSET $2`
		args = []interface{}{limit, offset}
	}

	err := r.db.db.SelectContext(ctx, &contexts, query, args...)
	if err != nil {
		return nil, err
	}

	return contexts, nil
}

// Delete deletes an MCP context
func (r *MCPContextRepository) Delete(ctx context.Context, contextID string) error {
	query := `DELETE FROM mcp_contexts WHERE id = $1`
	_, err := r.db.db.ExecContext(ctx, query, contextID)
	return err
}
