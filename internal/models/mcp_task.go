package models

import (
	"time"

	"github.com/google/uuid"
)

// MCPTaskType represents the type of MCP task
type MCPTaskType string

const (
	MCPTaskTypeCreateContext MCPTaskType = "create_context"
	MCPTaskTypeAddPrompt     MCPTaskType = "add_prompt"
	MCPTaskTypeAddNode       MCPTaskType = "add_node"
	MCPTaskTypeDeleteNode    MCPTaskType = "delete_node"
	MCPTaskTypeDeleteContext MCPTaskType = "delete_context"
)

// MCPTask represents a Model Context Protocol task
type MCPTask struct {
	ID          string       `json:"id" db:"id"`
	Name        string       `json:"name" db:"name"`
	Description string       `json:"description" db:"description"`
	ModelID     string       `json:"model_id" db:"model_id"`
	ContextID   *string      `json:"context_id" db:"context_id"`
	Type        MCPTaskType  `json:"type" db:"type"`
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

// NewMCPTask creates a new MCP task with default values
func NewMCPTask(name string, taskType MCPTaskType, modelID string, userID string, priority TaskPriority, input []byte) *MCPTask {
	return &MCPTask{
		ID:         uuid.New().String(),
		Name:       name,
		ModelID:    modelID,
		Type:       taskType,
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

// MCPCreateContextInput represents input for creating a new context
type MCPCreateContextInput struct {
	ModelID       string                 `json:"model_id"`
	Nodes         []MCPContextNode       `json:"nodes"`
	Metadata      map[string]interface{} `json:"metadata"`
	ReturnContext bool                   `json:"return_context"`
}

// MCPContextNode represents a node in the MCP context
type MCPContextNode struct {
	ID          string                 `json:"id"`
	Content     string                 `json:"content"`
	ContentType string                 `json:"content_type"`
	Metadata    map[string]interface{} `json:"metadata"`
	Parent      *string                `json:"parent"`
	Children    []string               `json:"children"`
}

// MCPAddPromptInput represents input for adding a prompt to a context
type MCPAddPromptInput struct {
	ContextID string                 `json:"context_id"`
	Prompt    string                 `json:"prompt"`
	PromptID  *string                `json:"prompt_id,omitempty"`
	ParentID  *string                `json:"parent_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
	Stream    bool                   `json:"stream"`
}

// MCPAddNodeInput represents input for adding a node to a context
type MCPAddNodeInput struct {
	ContextID string         `json:"context_id"`
	Node      MCPContextNode `json:"node"`
}

// MCPDeleteNodeInput represents input for deleting a node from a context
type MCPDeleteNodeInput struct {
	ContextID string `json:"context_id"`
	NodeID    string `json:"node_id"`
}

// MCPDeleteContextInput represents input for deleting a context
type MCPDeleteContextInput struct {
	ContextID string `json:"context_id"`
}

// MCPCreateContextOutput represents output from creating a new context
type MCPCreateContextOutput struct {
	ContextID string                 `json:"context_id"`
	Model     MCPModel               `json:"model"`
	Nodes     []MCPContextNode       `json:"nodes"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// MCPModel represents model information
type MCPModel struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Provider     string                 `json:"provider"`
	Capabilities []string               `json:"capabilities"`
	Config       map[string]interface{} `json:"config"`
}

// MCPAddPromptOutput represents output from adding a prompt to a context
type MCPAddPromptOutput struct {
	ContextID    string                 `json:"context_id"`
	PromptID     string                 `json:"prompt_id"`
	CompletionID string                 `json:"completion_id"`
	Completion   string                 `json:"completion"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// MCPAddNodeOutput represents output from adding a node to a context
type MCPAddNodeOutput struct {
	ContextID string         `json:"context_id"`
	Node      MCPContextNode `json:"node"`
}

// MCPDeleteNodeOutput represents output from deleting a node from a context
type MCPDeleteNodeOutput struct {
	ContextID string `json:"context_id"`
	Deleted   bool   `json:"deleted"`
}

// MCPDeleteContextOutput represents output from deleting a context
type MCPDeleteContextOutput struct {
	ContextID string `json:"context_id"`
	Deleted   bool   `json:"deleted"`
}
