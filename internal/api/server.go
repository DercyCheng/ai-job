package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"ai-job/internal/database"
	"ai-job/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Config represents API server configuration
type Config struct {
	Host           string
	Port           string
	Timeout        time.Duration
	MaxRequestSize int64
	MCPServerURL   string // URL for the MCP server
}

// Server represents the API server
type Server struct {
	router         *chi.Mux
	taskRepo       *database.TaskRepository
	workerRepo     *database.WorkerRepository
	mcpTaskRepo    *database.MCPTaskRepository
	mcpContextRepo *database.MCPContextRepository
	mcpHandler     *MCPHandler
	config         Config
}

// New creates a new API server
func New(taskRepo *database.TaskRepository, workerRepo *database.WorkerRepository,
	mcpTaskRepo *database.MCPTaskRepository, mcpContextRepo *database.MCPContextRepository,
	config Config) *Server {
	s := &Server{
		router:         chi.NewRouter(),
		taskRepo:       taskRepo,
		workerRepo:     workerRepo,
		mcpTaskRepo:    mcpTaskRepo,
		mcpContextRepo: mcpContextRepo,
		config:         config,
	}

	// Create MCP handler if enabled
	if mcpTaskRepo != nil && mcpContextRepo != nil {
		if config.MCPServerURL != "" {
			s.mcpHandler = NewMCPHandler(mcpTaskRepo, mcpContextRepo, config.MCPServerURL)
			log.Printf("MCP handler initialized with server URL: %s", config.MCPServerURL)
		}
	}

	s.setupRoutes()
	return s
}

// Start starts the API server
func (s *Server) Start() error {
	addr := s.config.Host + ":" + s.config.Port
	log.Printf("API server starting on %s", addr)
	return http.ListenAndServe(addr, s.router)
}

// setupRoutes configures the API routes
func (s *Server) setupRoutes() {
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(s.config.Timeout))

	s.router.Route("/api/v1", func(r chi.Router) {
		r.Route("/tasks", func(r chi.Router) {
			r.Get("/", s.listTasks)
			r.Post("/", s.createTask)
			r.Get("/{id}", s.getTask)
			r.Delete("/{id}", s.cancelTask)
		})

		r.Route("/workers", func(r chi.Router) {
			r.Get("/", s.listWorkers)
			r.Post("/", s.registerWorker)
			r.Put("/{id}/heartbeat", s.workerHeartbeat)
			r.Put("/{id}/status", s.updateWorkerStatus)
		})
	})

	// Register MCP routes if the handler is available
	if s.mcpHandler != nil {
		s.mcpHandler.RegisterRoutes(s.router)
	}
}

// CreateTaskRequest represents a request to create a new task
type CreateTaskRequest struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	ModelName   string              `json:"model_name"`
	Priority    models.TaskPriority `json:"priority"`
	Input       json.RawMessage     `json:"input"`
	UserID      string              `json:"user_id"`
	Timeout     int                 `json:"timeout,omitempty"`
	MaxRetries  int                 `json:"max_retries,omitempty"`
}

// createTask handles the creation of a new task
func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	task := models.NewTask(req.Name, req.ModelName, req.UserID, req.Priority, []byte(req.Input))
	task.Description = req.Description

	if req.Timeout > 0 {
		task.Timeout = req.Timeout
	}

	if req.MaxRetries > 0 {
		task.MaxRetries = req.MaxRetries
	}

	if err := s.taskRepo.Create(r.Context(), task); err != nil {
		log.Printf("Error creating task: %v", err)
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// getTask handles retrieving a task by ID
func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing task ID", http.StatusBadRequest)
		return
	}

	task, err := s.taskRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// listTasks handles listing all tasks
func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	limit := 100
	offset := 0

	var statusFilter *models.TaskStatus
	statusParam := r.URL.Query().Get("status")
	if statusParam != "" {
		status := models.TaskStatus(statusParam)
		statusFilter = &status
	}

	tasks, err := s.taskRepo.List(r.Context(), statusFilter, limit, offset)
	if err != nil {
		log.Printf("Error listing tasks: %v", err)
		http.Error(w, "Failed to list tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// cancelTask handles cancelling a task
func (s *Server) cancelTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing task ID", http.StatusBadRequest)
		return
	}

	task, err := s.taskRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Only allow cancellation of pending or scheduled tasks
	if task.Status != models.TaskStatusPending && task.Status != models.TaskStatusScheduled {
		http.Error(w, "Cannot cancel task in status: "+string(task.Status), http.StatusBadRequest)
		return
	}

	task.Status = models.TaskStatusCancelled
	task.UpdatedAt = time.Now()

	if err := s.taskRepo.Update(r.Context(), task); err != nil {
		log.Printf("Error cancelling task: %v", err)
		http.Error(w, "Failed to cancel task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// RegisterWorkerRequest represents a request to register a new worker
type RegisterWorkerRequest struct {
	Name            string   `json:"name"`
	Capabilities    []string `json:"capabilities"`
	AvailableMemory int64    `json:"available_memory"`
	AvailableCPU    float64  `json:"available_cpu"`
	AvailableGPU    float64  `json:"available_gpu"`
}

// registerWorker handles worker registration
func (s *Server) registerWorker(w http.ResponseWriter, r *http.Request) {
	var req RegisterWorkerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	worker := models.NewWorker(req.Name, req.Capabilities)
	worker.AvailableMemory = req.AvailableMemory
	worker.AvailableCPU = req.AvailableCPU
	worker.AvailableGPU = req.AvailableGPU

	if err := s.workerRepo.Create(r.Context(), worker); err != nil {
		log.Printf("Error registering worker: %v", err)
		http.Error(w, "Failed to register worker", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(worker)
}

// workerHeartbeat handles worker heartbeat updates
func (s *Server) workerHeartbeat(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing worker ID", http.StatusBadRequest)
		return
	}

	if err := s.workerRepo.UpdateHeartbeat(r.Context(), id); err != nil {
		http.Error(w, "Worker not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// UpdateWorkerStatusRequest represents a request to update worker status
type UpdateWorkerStatusRequest struct {
	Status          string  `json:"status"`
	CurrentTaskID   *string `json:"current_task_id"`
	TaskStatus      string  `json:"task_status,omitempty"`
	TaskOutput      []byte  `json:"task_output,omitempty"`
	TaskError       string  `json:"task_error,omitempty"`
	AvailableMemory int64   `json:"available_memory"`
	AvailableCPU    float64 `json:"available_cpu"`
	AvailableGPU    float64 `json:"available_gpu"`
}

// updateWorkerStatus handles updating worker status and task results
func (s *Server) updateWorkerStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing worker ID", http.StatusBadRequest)
		return
	}

	var req UpdateWorkerStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	worker, err := s.workerRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Worker not found", http.StatusNotFound)
		return
	}

	// Update worker status
	worker.Status = req.Status
	worker.CurrentTaskID = req.CurrentTaskID
	worker.LastHeartbeat = time.Now()
	worker.AvailableMemory = req.AvailableMemory
	worker.AvailableCPU = req.AvailableCPU
	worker.AvailableGPU = req.AvailableGPU

	// Update worker in database
	if err := s.workerRepo.Update(r.Context(), worker); err != nil {
		log.Printf("Error updating worker: %v", err)
		http.Error(w, "Failed to update worker", http.StatusInternalServerError)
		return
	}

	// Handle task status update if provided
	if req.TaskStatus != "" && worker.CurrentTaskID != nil {
		if err := s.updateTaskStatus(r.Context(), *worker.CurrentTaskID, req); err != nil {
			log.Printf("Error updating task: %v", err)
			http.Error(w, "Failed to update task", http.StatusInternalServerError)
			return
		}

		// Increment task count if task is completed
		if req.TaskStatus == string(models.TaskStatusCompleted) || req.TaskStatus == string(models.TaskStatusFailed) {
			worker.TotalTasksHandled++
			worker.CurrentTaskID = nil
			worker.Status = "available"

			if err := s.workerRepo.Update(r.Context(), worker); err != nil {
				log.Printf("Error updating worker stats: %v", err)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(worker)
}

// updateTaskStatus updates a task status based on worker update
func (s *Server) updateTaskStatus(ctx context.Context, taskID string, req UpdateWorkerStatusRequest) error {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	task.Status = models.TaskStatus(req.TaskStatus)
	task.UpdatedAt = time.Now()

	if req.TaskStatus == string(models.TaskStatusRunning) && task.StartedAt == nil {
		now := time.Now()
		task.StartedAt = &now
	}

	if req.TaskStatus == string(models.TaskStatusCompleted) || req.TaskStatus == string(models.TaskStatusFailed) {
		now := time.Now()
		task.CompletedAt = &now

		if req.TaskOutput != nil {
			task.Output = req.TaskOutput
		}

		if req.TaskError != "" {
			task.Error = req.TaskError
		}
	}

	return s.taskRepo.Update(ctx, task)
}

// listWorkers handles listing all workers
func (s *Server) listWorkers(w http.ResponseWriter, r *http.Request) {
	workers, err := s.workerRepo.ListAvailable(r.Context())
	if err != nil {
		log.Printf("Error listing workers: %v", err)
		http.Error(w, "Failed to list workers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workers)
}
