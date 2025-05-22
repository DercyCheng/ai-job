package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"ai-job/internal/database"
	"ai-job/internal/metrics"
	"ai-job/internal/models"
	"ai-job/pkg/mcp"
)

// MCPWorker handles MCP tasks
type MCPWorker struct {
	mcpTaskRepo    *database.MCPTaskRepository
	mcpContextRepo *database.MCPContextRepository
	mcpClient      *mcp.Client
	workerID       string
	metrics        *metrics.Metrics
	lastHeartbeat  time.Time
	mu             sync.RWMutex
}

// NewMCPWorker creates a new MCP worker
func NewMCPWorker(mcpTaskRepo *database.MCPTaskRepository, mcpContextRepo *database.MCPContextRepository, mcpServerURL, workerID string) *MCPWorker {
	return &MCPWorker{
		mcpTaskRepo:    mcpTaskRepo,
		mcpContextRepo: mcpContextRepo,
		mcpClient:      mcp.NewClient(mcpServerURL),
		workerID:       workerID,
		metrics:        metrics.GetMetrics(),
		lastHeartbeat:  time.Now(),
	}
}

// ProcessTask processes an MCP task
func (w *MCPWorker) ProcessTask(ctx context.Context, task *models.MCPTask) error {
	// Update metrics
	w.metrics.TasksInProgress.Inc()
	startTime := time.Now()
	defer func() {
		w.metrics.TasksInProgress.Dec()
		w.metrics.TaskDuration.WithLabelValues(string(task.Type)).
			Observe(time.Since(startTime).Seconds())
	}()

	// Update task status to running
	task.Status = models.TaskStatusRunning
	task.StartedAt = timePtr(time.Now())
	if err := w.mcpTaskRepo.Update(ctx, task); err != nil {
		w.metrics.WorkerErrors.WithLabelValues("status_update").Inc()
		return fmt.Errorf("failed to update task status: %w", err)
	}

	var result []byte
	var err error

	// Process task based on type
	w.metrics.TaskTypeCount.WithLabelValues(string(task.Type)).Inc()
	switch task.Type {
	case models.MCPTaskTypeCreateContext:
		result, err = w.handleCreateContext(ctx, task)
	case models.MCPTaskTypeAddPrompt:
		result, err = w.handleAddPrompt(ctx, task)
	case models.MCPTaskTypeAddNode:
		result, err = w.handleAddNode(ctx, task)
	case models.MCPTaskTypeDeleteNode:
		result, err = w.handleDeleteNode(ctx, task)
	case models.MCPTaskTypeDeleteContext:
		result, err = w.handleDeleteContext(ctx, task)
	default:
		err = fmt.Errorf("unsupported MCP task type: %s", task.Type)
		w.metrics.WorkerErrors.WithLabelValues("invalid_type").Inc()
	}

	// Update task status based on result
	if err != nil {
		task.Status = models.TaskStatusFailed
		task.Error = err.Error()
		w.metrics.TasksCompleted.WithLabelValues("failed").Inc()
		w.metrics.WorkerErrors.WithLabelValues("execution").Inc()
		log.Printf("MCP task %s failed: %v", task.ID, err)
	} else {
		task.Status = models.TaskStatusCompleted
		task.Output = result
		w.metrics.TasksCompleted.WithLabelValues("completed").Inc()
		log.Printf("MCP task %s completed successfully", task.ID)
	}

	task.CompletedAt = timePtr(time.Now())
	if updateErr := w.mcpTaskRepo.Update(ctx, task); updateErr != nil {
		w.metrics.WorkerErrors.WithLabelValues("status_update").Inc()
		return fmt.Errorf("failed to update task status: %w", updateErr)
	}

	return err
}

// handleCreateContext handles creating a new context
func (w *MCPWorker) handleCreateContext(ctx context.Context, task *models.MCPTask) ([]byte, error) {
	var input models.MCPCreateContextInput
	if err := json.Unmarshal(task.Input, &input); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input: %w", err)
	}

	// Convert to MCP client request
	req := mcp.CreateContextRequest{
		ModelID:       input.ModelID,
		ReturnContext: input.ReturnContext,
		Metadata:      input.Metadata,
	}

	// Convert nodes if any
	for _, node := range input.Nodes {
		req.Nodes = append(req.Nodes, mcp.ContextNode{
			ID:          node.ID,
			Content:     node.Content,
			ContentType: node.ContentType,
			Metadata:    node.Metadata,
			Parent:      node.Parent,
			Children:    node.Children,
		})
	}

	// Call MCP server
	resp, err := w.mcpClient.CreateContext(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create context: %w", err)
	}

	// Store context ID in task
	task.ContextID = &resp.ContextID

	// Store context data in database for persistence
	contextData, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal context data: %w", err)
	}

	if err := w.mcpContextRepo.Store(ctx, resp.ContextID, input.ModelID, task.UserID, contextData); err != nil {
		return nil, fmt.Errorf("failed to store context data: %w", err)
	}

	// Return the response
	output, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}

	return output, nil
}

// handleAddPrompt handles adding a prompt to a context
func (w *MCPWorker) handleAddPrompt(ctx context.Context, task *models.MCPTask) ([]byte, error) {
	var input models.MCPAddPromptInput
	if err := json.Unmarshal(task.Input, &input); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input: %w", err)
	}

	// Convert to MCP client request
	req := mcp.PromptRequest{
		ContextID: input.ContextID,
		Prompt:    input.Prompt,
		PromptID:  input.PromptID,
		ParentID:  input.ParentID,
		Metadata:  input.Metadata,
		Stream:    input.Stream,
	}

	// Call MCP server
	resp, err := w.mcpClient.Prompt(ctx, input.ContextID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to add prompt: %w", err)
	}

	// Update context data in database
	contextResp, err := w.mcpClient.GetContext(ctx, input.ContextID)
	if err != nil {
		log.Printf("Warning: failed to get updated context data: %v", err)
	} else {
		contextData, err := json.Marshal(contextResp)
		if err != nil {
			log.Printf("Warning: failed to marshal context data: %v", err)
		} else {
			if err := w.mcpContextRepo.Store(ctx, input.ContextID, task.ModelID, task.UserID, contextData); err != nil {
				log.Printf("Warning: failed to store updated context data: %v", err)
			}
		}
	}

	// Return the response
	output, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}

	return output, nil
}

// handleAddNode handles adding a node to a context
func (w *MCPWorker) handleAddNode(ctx context.Context, task *models.MCPTask) ([]byte, error) {
	var input models.MCPAddNodeInput
	if err := json.Unmarshal(task.Input, &input); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input: %w", err)
	}

	// Convert to MCP client request
	node := mcp.ContextNode{
		ID:          input.Node.ID,
		Content:     input.Node.Content,
		ContentType: input.Node.ContentType,
		Metadata:    input.Node.Metadata,
		Parent:      input.Node.Parent,
		Children:    input.Node.Children,
	}

	// Call MCP server
	resp, err := w.mcpClient.AddNode(ctx, input.ContextID, node)
	if err != nil {
		return nil, fmt.Errorf("failed to add node: %w", err)
	}

	// Update context data in database
	contextResp, err := w.mcpClient.GetContext(ctx, input.ContextID)
	if err != nil {
		log.Printf("Warning: failed to get updated context data: %v", err)
	} else {
		contextData, err := json.Marshal(contextResp)
		if err != nil {
			log.Printf("Warning: failed to marshal context data: %v", err)
		} else {
			if err := w.mcpContextRepo.Store(ctx, input.ContextID, task.ModelID, task.UserID, contextData); err != nil {
				log.Printf("Warning: failed to store updated context data: %v", err)
			}
		}
	}

	// Return the response
	output, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}

	return output, nil
}

// handleDeleteNode handles deleting a node from a context
func (w *MCPWorker) handleDeleteNode(ctx context.Context, task *models.MCPTask) ([]byte, error) {
	var input models.MCPDeleteNodeInput
	if err := json.Unmarshal(task.Input, &input); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input: %w", err)
	}

	// Call MCP server
	resp, err := w.mcpClient.DeleteNode(ctx, input.ContextID, input.NodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete node: %w", err)
	}

	// Update context data in database if deletion was successful
	if resp.Deleted {
		contextResp, err := w.mcpClient.GetContext(ctx, input.ContextID)
		if err != nil {
			log.Printf("Warning: failed to get updated context data: %v", err)
		} else {
			contextData, err := json.Marshal(contextResp)
			if err != nil {
				log.Printf("Warning: failed to marshal context data: %v", err)
			} else {
				if err := w.mcpContextRepo.Store(ctx, input.ContextID, task.ModelID, task.UserID, contextData); err != nil {
					log.Printf("Warning: failed to store updated context data: %v", err)
				}
			}
		}
	}

	// Return the response
	output, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}

	return output, nil
}

// handleDeleteContext handles deleting a context
func (w *MCPWorker) handleDeleteContext(ctx context.Context, task *models.MCPTask) ([]byte, error) {
	var input models.MCPDeleteContextInput
	if err := json.Unmarshal(task.Input, &input); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input: %w", err)
	}

	// Call MCP server
	resp, err := w.mcpClient.DeleteContext(ctx, input.ContextID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete context: %w", err)
	}

	// Delete context from database if deletion was successful
	if resp.Deleted {
		if err := w.mcpContextRepo.Delete(ctx, input.ContextID); err != nil {
			log.Printf("Warning: failed to delete context from database: %v", err)
		}
	}

	// Return the response
	output, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}

	return output, nil
}

// Helper function to create time pointer
func (w *MCPWorker) CheckHealth(ctx context.Context) error {
	// Check last heartbeat time
	w.mu.RLock()
	lastHeartbeat := w.lastHeartbeat
	w.mu.RUnlock()

	if time.Since(lastHeartbeat) > 2*time.Minute {
		w.metrics.WorkerErrors.WithLabelValues("heartbeat_timeout").Inc()
		return fmt.Errorf("worker heartbeat timeout")
	}

	// Check resource usage
	if err := w.reportResourceUsage(ctx); err != nil {
		w.metrics.WorkerErrors.WithLabelValues("resource_check").Inc()
		return fmt.Errorf("resource check failed: %w", err)
	}

	return nil
}

// reportResourceUsage collects and reports worker resource usage
func (w *MCPWorker) reportResourceUsage(ctx context.Context) error {
	// TODO: Implement actual resource monitoring
	// For now report default values
	cpuUsage := 0.5
	memUsage := 1.0
	gpuUsage := 0.3

	w.metrics.ResourceUsage.WithLabelValues("cpu", "cores").Set(cpuUsage)
	w.metrics.ResourceUsage.WithLabelValues("memory", "gb").Set(memUsage)
	w.metrics.ResourceUsage.WithLabelValues("gpu", "percent").Set(gpuUsage)

	return nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}
