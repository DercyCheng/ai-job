package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"ai-job/internal/database"
	"ai-job/internal/models"
	"ai-job/pkg/mcp"

	"github.com/go-chi/chi/v5"
)

// MCPHandler handles MCP API endpoints
type MCPHandler struct {
	mcpTaskRepo    *database.MCPTaskRepository
	mcpContextRepo *database.MCPContextRepository
	mcpClient      *mcp.Client
}

// NewMCPHandler creates a new MCP handler
func NewMCPHandler(mcpTaskRepo *database.MCPTaskRepository, mcpContextRepo *database.MCPContextRepository, mcpServerURL string) *MCPHandler {
	return &MCPHandler{
		mcpTaskRepo:    mcpTaskRepo,
		mcpContextRepo: mcpContextRepo,
		mcpClient:      mcp.NewClient(mcpServerURL),
	}
}

// RegisterRoutes registers MCP API routes
func (h *MCPHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/mcp", func(r chi.Router) {
		r.Post("/contexts", h.createContext)
		r.Get("/contexts", h.listContexts)
		r.Get("/contexts/{contextID}", h.getContext)
		r.Delete("/contexts/{contextID}", h.deleteContext)
		r.Post("/contexts/{contextID}/prompt", h.addPrompt)
		r.Post("/contexts/{contextID}/nodes", h.addNode)
		r.Delete("/contexts/{contextID}/nodes/{nodeID}", h.deleteNode)
		r.Get("/tasks", h.listTasks)
		r.Get("/tasks/{taskID}", h.getTask)
		r.Get("/health", h.healthCheck)
		r.Get("/models", h.listModels)
	})
}

// CreateContextRequest represents a request to create a new context
type CreateContextRequest struct {
	ModelID       string                  `json:"model_id"`
	Nodes         []models.MCPContextNode `json:"nodes"`
	Metadata      map[string]interface{}  `json:"metadata"`
	ReturnContext bool                    `json:"return_context"`
	UserID        string                  `json:"user_id"`
	Priority      models.TaskPriority     `json:"priority"`
}

// createContext handles creating a new context
func (h *MCPHandler) createContext(w http.ResponseWriter, r *http.Request) {
	var req CreateContextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create input data for the task
	input := models.MCPCreateContextInput{
		ModelID:       req.ModelID,
		Nodes:         req.Nodes,
		Metadata:      req.Metadata,
		ReturnContext: req.ReturnContext,
	}

	// Marshal input to JSON
	inputBytes, err := json.Marshal(input)
	if err != nil {
		log.Printf("Error marshaling input: %v", err)
		http.Error(w, "Failed to process request", http.StatusInternalServerError)
		return
	}

	// Create a new task
	task := models.NewMCPTask(
		"Create MCP Context",
		models.MCPTaskTypeCreateContext,
		req.ModelID,
		req.UserID,
		req.Priority,
		inputBytes,
	)

	// Save the task
	if err := h.mcpTaskRepo.Create(r.Context(), task); err != nil {
		log.Printf("Error creating task: %v", err)
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	// Return the task ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"task_id": task.ID,
		"status":  string(task.Status),
	})
}

// PromptRequest represents a request to add a prompt to a context
type PromptRequest struct {
	Prompt   string                 `json:"prompt"`
	PromptID *string                `json:"prompt_id,omitempty"`
	ParentID *string                `json:"parent_id,omitempty"`
	Metadata map[string]interface{} `json:"metadata"`
	Stream   bool                   `json:"stream"`
	UserID   string                 `json:"user_id"`
	Priority models.TaskPriority    `json:"priority"`
}

// addPrompt handles adding a prompt to a context
func (h *MCPHandler) addPrompt(w http.ResponseWriter, r *http.Request) {
	contextID := chi.URLParam(r, "contextID")
	if contextID == "" {
		http.Error(w, "Missing context ID", http.StatusBadRequest)
		return
	}

	var req PromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Handle streaming if requested
	if req.Stream {
		h.handleStreamingPrompt(w, r, contextID, req)
		return
	}

	// Create input data for the task
	input := models.MCPAddPromptInput{
		ContextID: contextID,
		Prompt:    req.Prompt,
		PromptID:  req.PromptID,
		ParentID:  req.ParentID,
		Metadata:  req.Metadata,
		Stream:    req.Stream,
	}

	// Marshal input to JSON
	inputBytes, err := json.Marshal(input)
	if err != nil {
		log.Printf("Error marshaling input: %v", err)
		http.Error(w, "Failed to process request", http.StatusInternalServerError)
		return
	}

	// Get the model ID from the context
	context, err := h.mcpContextRepo.Get(r.Context(), contextID)
	if err != nil {
		log.Printf("Error retrieving context: %v", err)
		http.Error(w, "Context not found", http.StatusNotFound)
		return
	}

	// Create a new task
	task := models.NewMCPTask(
		"Add Prompt to MCP Context",
		models.MCPTaskTypeAddPrompt,
		context.ModelID,
		req.UserID,
		req.Priority,
		inputBytes,
	)
	task.ContextID = &contextID

	// Save the task
	if err := h.mcpTaskRepo.Create(r.Context(), task); err != nil {
		log.Printf("Error creating task: %v", err)
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	// Return the task ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"task_id": task.ID,
		"status":  string(task.Status),
	})
}

// handleStreamingPrompt handles streaming prompt requests directly to the MCP server
func (h *MCPHandler) handleStreamingPrompt(w http.ResponseWriter, r *http.Request, contextID string, req PromptRequest) {
	// Set headers for streaming response
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create the prompt request for the MCP client
	promptReq := mcp.AddPromptRequest{
		Prompt:   req.Prompt,
		Metadata: req.Metadata,
		Stream:   true,
	}

	if req.PromptID != nil {
		promptReq.PromptID = *req.PromptID
	}

	if req.ParentID != nil {
		promptReq.ParentID = *req.ParentID
	}

	// Get a streaming response from the MCP client
	stream, err := h.mcpClient.AddPromptStream(r.Context(), contextID, promptReq)
	if err != nil {
		log.Printf("Error starting stream: %v", err)
		http.Error(w, "Failed to start stream", http.StatusInternalServerError)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Printf("Error: ResponseWriter does not support flushing")
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Create a closure to handle cancellation
	ctx := r.Context()

	// Create an MCP context record for tracking
	mcpContext, err := h.mcpContextRepo.Get(ctx, contextID)
	if err != nil {
		log.Printf("Error retrieving context: %v", err)
		http.Error(w, "Context not found", http.StatusNotFound)
		return
	}

	// Create a streaming task record
	streamTask := models.NewMCPTask(
		"Streaming Prompt",
		models.MCPTaskTypeAddPrompt,
		mcpContext.ModelID,
		req.UserID,
		req.Priority,
		[]byte(req.Prompt),
	)
	streamTask.ContextID = &contextID
	streamTask.Status = models.TaskStatusRunning
	startTime := time.Now()
	streamTask.StartedAt = &startTime

	if err := h.mcpTaskRepo.Create(ctx, streamTask); err != nil {
		log.Printf("Error creating stream task record: %v", err)
		// Continue anyway, this is just for tracking
	}

	// Stream each chunk to the client
	var fullCompletion strings.Builder
	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			log.Printf("Client disconnected from stream for task %s", streamTask.ID)
			return
		default:
			chunk, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					// Stream completed normally
					log.Printf("Stream completed for task %s", streamTask.ID)

					// Update the task record
					streamTask.Status = models.TaskStatusCompleted
					endTime := time.Now()
					streamTask.CompletedAt = &endTime
					streamTask.Output = []byte(fullCompletion.String())

					if err := h.mcpTaskRepo.Update(context.Background(), streamTask); err != nil {
						log.Printf("Error updating stream task record: %v", err)
					}

					return
				}

				// Handle stream error
				log.Printf("Error receiving stream chunk: %v", err)

				// Update the task record
				streamTask.Status = models.TaskStatusFailed
				streamTask.Error = err.Error()
				endTime := time.Now()
				streamTask.CompletedAt = &endTime

				if err := h.mcpTaskRepo.Update(context.Background(), streamTask); err != nil {
					log.Printf("Error updating stream task record: %v", err)
				}

				return
			}

			// Format the chunk as a server-sent event
			data, err := json.Marshal(chunk)
			if err != nil {
				log.Printf("Error marshaling chunk: %v", err)
				continue
			}

			// Append to full completion
			if chunk.Completion != "" {
				fullCompletion.WriteString(chunk.Completion)
			}

			// Write the event to the response
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

// AddNodeRequest represents a request to add a node to a context
type AddNodeRequest struct {
	Node     models.MCPContextNode `json:"node"`
	UserID   string                `json:"user_id"`
	Priority models.TaskPriority   `json:"priority"`
}

// addNode handles adding a node to a context
func (h *MCPHandler) addNode(w http.ResponseWriter, r *http.Request) {
	contextID := chi.URLParam(r, "contextID")
	if contextID == "" {
		http.Error(w, "Missing context ID", http.StatusBadRequest)
		return
	}

	var req AddNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create input data for the task
	input := models.MCPAddNodeInput{
		ContextID: contextID,
		Node:      req.Node,
	}

	// Marshal input to JSON
	inputBytes, err := json.Marshal(input)
	if err != nil {
		log.Printf("Error marshaling input: %v", err)
		http.Error(w, "Failed to process request", http.StatusInternalServerError)
		return
	}

	// Get the model ID from the context
	context, err := h.mcpContextRepo.Get(r.Context(), contextID)
	if err != nil {
		log.Printf("Error retrieving context: %v", err)
		http.Error(w, "Context not found", http.StatusNotFound)
		return
	}

	// Create a new task
	task := models.NewMCPTask(
		"Add Node to MCP Context",
		models.MCPTaskTypeAddNode,
		context.ModelID,
		req.UserID,
		req.Priority,
		inputBytes,
	)
	task.ContextID = &contextID

	// Save the task
	if err := h.mcpTaskRepo.Create(r.Context(), task); err != nil {
		log.Printf("Error creating task: %v", err)
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	// Return the task ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"task_id": task.ID,
		"status":  string(task.Status),
	})
}

// deleteNode handles deleting a node from a context
func (h *MCPHandler) deleteNode(w http.ResponseWriter, r *http.Request) {
	contextID := chi.URLParam(r, "contextID")
	nodeID := chi.URLParam(r, "nodeID")

	if contextID == "" || nodeID == "" {
		http.Error(w, "Missing context ID or node ID", http.StatusBadRequest)
		return
	}

	// Get userID and priority from query parameters
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	priorityStr := r.URL.Query().Get("priority")
	priority := models.TaskPriorityNormal
	if priorityStr != "" {
		// Parse priority - simplified for this example
		switch priorityStr {
		case "high":
			priority = models.TaskPriorityHigh
		case "critical":
			priority = models.TaskPriorityCritical
		case "low":
			priority = models.TaskPriorityLow
		}
	}

	// Create input data for the task
	input := models.MCPDeleteNodeInput{
		ContextID: contextID,
		NodeID:    nodeID,
	}

	// Marshal input to JSON
	inputBytes, err := json.Marshal(input)
	if err != nil {
		log.Printf("Error marshaling input: %v", err)
		http.Error(w, "Failed to process request", http.StatusInternalServerError)
		return
	}

	// Get the model ID from the context
	context, err := h.mcpContextRepo.Get(r.Context(), contextID)
	if err != nil {
		log.Printf("Error retrieving context: %v", err)
		http.Error(w, "Context not found", http.StatusNotFound)
		return
	}

	// Create a new task
	task := models.NewMCPTask(
		"Delete Node from MCP Context",
		models.MCPTaskTypeDeleteNode,
		context.ModelID,
		userID,
		priority,
		inputBytes,
	)
	task.ContextID = &contextID

	// Save the task
	if err := h.mcpTaskRepo.Create(r.Context(), task); err != nil {
		log.Printf("Error creating task: %v", err)
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	// Return the task ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"task_id": task.ID,
		"status":  string(task.Status),
	})
}

// deleteContext handles deleting a context
func (h *MCPHandler) deleteContext(w http.ResponseWriter, r *http.Request) {
	contextID := chi.URLParam(r, "contextID")
	if contextID == "" {
		http.Error(w, "Missing context ID", http.StatusBadRequest)
		return
	}

	// Get userID and priority from query parameters
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	priorityStr := r.URL.Query().Get("priority")
	priority := models.TaskPriorityNormal
	if priorityStr != "" {
		// Parse priority - simplified for this example
		switch priorityStr {
		case "high":
			priority = models.TaskPriorityHigh
		case "critical":
			priority = models.TaskPriorityCritical
		case "low":
			priority = models.TaskPriorityLow
		}
	}

	// Create input data for the task
	input := models.MCPDeleteContextInput{
		ContextID: contextID,
	}

	// Marshal input to JSON
	inputBytes, err := json.Marshal(input)
	if err != nil {
		log.Printf("Error marshaling input: %v", err)
		http.Error(w, "Failed to process request", http.StatusInternalServerError)
		return
	}

	// Get the model ID from the context
	context, err := h.mcpContextRepo.Get(r.Context(), contextID)
	if err != nil {
		log.Printf("Error retrieving context: %v", err)
		http.Error(w, "Context not found", http.StatusNotFound)
		return
	}

	// Create a new task
	task := models.NewMCPTask(
		"Delete MCP Context",
		models.MCPTaskTypeDeleteContext,
		context.ModelID,
		userID,
		priority,
		inputBytes,
	)
	task.ContextID = &contextID

	// Save the task
	if err := h.mcpTaskRepo.Create(r.Context(), task); err != nil {
		log.Printf("Error creating task: %v", err)
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	// Return the task ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"task_id": task.ID,
		"status":  string(task.Status),
	})
}

// listContexts handles listing all contexts
func (h *MCPHandler) listContexts(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	limit := 100
	offset := 0

	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}

	contexts, err := h.mcpContextRepo.List(r.Context(), userIDPtr, limit, offset)
	if err != nil {
		log.Printf("Error listing contexts: %v", err)
		http.Error(w, "Failed to list contexts", http.StatusInternalServerError)
		return
	}

	// Format the response
	type contextInfo struct {
		ID        string    `json:"id"`
		ModelID   string    `json:"model_id"`
		UserID    string    `json:"user_id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	response := make([]contextInfo, len(contexts))
	for i, ctx := range contexts {
		response[i] = contextInfo{
			ID:        ctx.ID,
			ModelID:   ctx.ModelID,
			UserID:    ctx.UserID,
			CreatedAt: ctx.CreatedAt,
			UpdatedAt: ctx.UpdatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getContext handles retrieving a context by ID
func (h *MCPHandler) getContext(w http.ResponseWriter, r *http.Request) {
	contextID := chi.URLParam(r, "contextID")
	if contextID == "" {
		http.Error(w, "Missing context ID", http.StatusBadRequest)
		return
	}

	context, err := h.mcpContextRepo.Get(r.Context(), contextID)
	if err != nil {
		log.Printf("Error retrieving context: %v", err)
		http.Error(w, "Context not found", http.StatusNotFound)
		return
	}

	// Parse the context data
	var contextData map[string]interface{}
	if err := json.Unmarshal(context.Data, &contextData); err != nil {
		log.Printf("Error parsing context data: %v", err)
		http.Error(w, "Failed to parse context data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contextData)
}

// listTasks handles listing MCP tasks
func (h *MCPHandler) listTasks(w http.ResponseWriter, r *http.Request) {
	statusParam := r.URL.Query().Get("status")
	typeParam := r.URL.Query().Get("type")
	limit := 100
	offset := 0

	var statusPtr *models.TaskStatus
	if statusParam != "" {
		status := models.TaskStatus(statusParam)
		statusPtr = &status
	}

	var typePtr *models.MCPTaskType
	if typeParam != "" {
		taskType := models.MCPTaskType(typeParam)
		typePtr = &taskType
	}

	tasks, err := h.mcpTaskRepo.List(r.Context(), statusPtr, typePtr, limit, offset)
	if err != nil {
		log.Printf("Error listing tasks: %v", err)
		http.Error(w, "Failed to list tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// getTask handles retrieving a task by ID
func (h *MCPHandler) getTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	if taskID == "" {
		http.Error(w, "Missing task ID", http.StatusBadRequest)
		return
	}

	task, err := h.mcpTaskRepo.GetByID(r.Context(), taskID)
	if err != nil {
		log.Printf("Error retrieving task: %v", err)
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// healthCheck handles checking the health of the MCP server
func (h *MCPHandler) healthCheck(w http.ResponseWriter, r *http.Request) {
	healthy, err := h.mcpClient.CheckHealth(r.Context())
	if err != nil || !healthy {
		log.Printf("MCP server health check failed: %v", err)
		http.Error(w, "MCP server is not healthy", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// listModels handles listing available models from the MCP server
func (h *MCPHandler) listModels(w http.ResponseWriter, r *http.Request) {
	models, err := h.mcpClient.ListModels(r.Context())
	if err != nil {
		log.Printf("Error listing models: %v", err)
		http.Error(w, "Failed to list models", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
}
