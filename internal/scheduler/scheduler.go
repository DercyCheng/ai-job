package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"ai-job/internal/database"
	"ai-job/internal/metrics"
	"ai-job/internal/models"
)

// WorkerResources tracks worker resource availability
type WorkerResources struct {
	CPU    float64
	Memory float64
	GPU    float64
	mu     sync.RWMutex
}

func (wr *WorkerResources) Update(cpu, memory, gpu float64) {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	wr.CPU = cpu
	wr.Memory = memory
	wr.GPU = gpu

	// Update metrics
	m := metrics.GetMetrics()
	m.ResourceUsage.WithLabelValues("cpu", "cores").Set(cpu)
	m.ResourceUsage.WithLabelValues("memory", "gb").Set(memory)
	m.ResourceUsage.WithLabelValues("gpu", "percent").Set(gpu)
}

func (wr *WorkerResources) Get() (float64, float64, float64) {
	wr.mu.RLock()
	defer wr.mu.RUnlock()
	return wr.CPU, wr.Memory, wr.GPU
}

// Config represents scheduler configuration
type Config struct {
	PollInterval      time.Duration
	MaxTasks          int
	HeartbeatInterval time.Duration
	TaskTimeout       time.Duration
}

// Scheduler is responsible for assigning tasks to workers
type Scheduler struct {
	taskRepo   *database.TaskRepository
	workerRepo *database.WorkerRepository
	config     Config
	stopCh     chan struct{}
	waitGroup  sync.WaitGroup
	resources  *WorkerResources
	metrics    *metrics.Metrics
}

// New creates a new scheduler
func New(taskRepo *database.TaskRepository, workerRepo *database.WorkerRepository, config Config) *Scheduler {
	return &Scheduler{
		taskRepo:   taskRepo,
		workerRepo: workerRepo,
		config:     config,
		stopCh:     make(chan struct{}),
		resources:  &WorkerResources{},
		metrics:    metrics.GetMetrics(),
	}
}

// Start starts the scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	s.waitGroup.Add(1)
	go s.schedulingLoop(ctx)

	log.Println("Scheduler started")
	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	close(s.stopCh)
	s.waitGroup.Wait()
	log.Println("Scheduler stopped")
}

// schedulingLoop runs the main scheduling loop
func (s *Scheduler) schedulingLoop(ctx context.Context) {
	defer s.waitGroup.Done()

	ticker := time.NewTicker(s.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			if err := s.processPendingTasks(ctx); err != nil {
				log.Printf("Error processing pending tasks: %v", err)
			}
		}
	}
}

// processPendingTasks processes pending tasks and assigns them to available workers
func (s *Scheduler) processPendingTasks(ctx context.Context) error {
	// Get pending tasks
	pendingTasks, err := s.taskRepo.GetPendingTasks(ctx, s.config.MaxTasks)
	if err != nil {
		return err
	}

	if len(pendingTasks) == 0 {
		return nil
	}

	// Get available workers
	availableWorkers, err := s.workerRepo.ListAvailable(ctx)
	if err != nil {
		return err
	}

	if len(availableWorkers) == 0 {
		return nil
	}

	// Update metrics
	s.metrics.TasksQueued.Set(float64(len(pendingTasks)))
	s.metrics.TasksInProgress.Set(float64(len(availableWorkers)))

	// Smart task assignment with resource awareness
	for _, task := range pendingTasks {
		assigned := false

		// Try to find best worker for the task
		for i, worker := range availableWorkers {
			if s.canWorkerHandleTask(worker, task) {
				// Update task status
				task.Status = models.TaskStatusScheduled
				task.WorkerID = &worker.ID
				task.UpdatedAt = time.Now()

				if err := s.taskRepo.Update(ctx, task); err != nil {
					log.Printf("Error updating task %s: %v", task.ID, err)
					continue
				}

				// Update worker status and resources
				worker.Status = "busy"
				worker.CurrentTaskID = &task.ID
				s.updateWorkerResources(worker, task)

				if err := s.workerRepo.Update(ctx, worker); err != nil {
					log.Printf("Error updating worker %s: %v", worker.ID, err)

					// Revert task status if worker update fails
					task.Status = models.TaskStatusPending
					task.WorkerID = nil
					if err := s.taskRepo.Update(ctx, task); err != nil {
						log.Printf("Error reverting task %s: %v", task.ID, err)
					}
					continue
				}

				log.Printf("Assigned task %s to worker %s", task.ID, worker.ID)
				s.metrics.TasksCompleted.WithLabelValues("scheduled").Inc()

				// Remove assigned worker from available list
				availableWorkers = append(availableWorkers[:i], availableWorkers[i+1:]...)
				assigned = true
				break
			}
		}

		if !assigned && len(availableWorkers) > 0 {
			// Fallback to simple round-robin if no suitable worker found
			worker := availableWorkers[0]

			// Update task status
			task.Status = models.TaskStatusScheduled
			task.WorkerID = &worker.ID
			task.UpdatedAt = time.Now()

			if err := s.taskRepo.Update(ctx, task); err != nil {
				log.Printf("Error updating task %s: %v", task.ID, err)
				continue
			}

			// Update worker status
			worker.Status = "busy"
			worker.CurrentTaskID = &task.ID
			s.updateWorkerResources(worker, task)

			if err := s.workerRepo.Update(ctx, worker); err != nil {
				log.Printf("Error updating worker %s: %v", worker.ID, err)

				// Revert task status if worker update fails
				task.Status = models.TaskStatusPending
				task.WorkerID = nil
				if err := s.taskRepo.Update(ctx, task); err != nil {
					log.Printf("Error reverting task %s: %v", task.ID, err)
				}
				continue
			}

			log.Printf("Assigned task %s to worker %s (fallback)", task.ID, worker.ID)
			s.metrics.TasksCompleted.WithLabelValues("scheduled").Inc()

			// Remove assigned worker from available list
			availableWorkers = availableWorkers[1:]
		}
	}

	return nil
}

// checkTaskTimeouts checks for and handles timed-out tasks
func (s *Scheduler) checkTaskTimeouts(ctx context.Context) error {
	runningStatus := models.TaskStatusRunning
	runningTasks, err := s.taskRepo.List(ctx, &runningStatus, 100, 0)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, task := range runningTasks {
		if task.StartedAt == nil {
			continue
		}

		taskDuration := now.Sub(*task.StartedAt)
		if taskDuration > s.config.TaskTimeout {
			log.Printf("Task %s has timed out after %.2f seconds", task.ID, taskDuration.Seconds())

			// Update task status
			task.Status = models.TaskStatusFailed
			task.Error = "Task timed out"
			task.UpdatedAt = now

			if err := s.taskRepo.Update(ctx, task); err != nil {
				log.Printf("Error updating timed-out task %s: %v", task.ID, err)
				continue
			}

			// If there's a worker assigned, update its status
			if task.WorkerID != nil {
				worker, err := s.workerRepo.GetByID(ctx, *task.WorkerID)
				if err != nil {
					log.Printf("Error fetching worker %s: %v", *task.WorkerID, err)
					continue
				}

				worker.Status = "available"
				worker.CurrentTaskID = nil

				if err := s.workerRepo.Update(ctx, worker); err != nil {
					log.Printf("Error updating worker %s after task timeout: %v", worker.ID, err)
				}
			}
		}
	}

	return nil
}

// handleFailedWorkers checks for workers that haven't sent heartbeats recently
func (s *Scheduler) handleFailedWorkers(ctx context.Context) error {
	workers, err := s.workerRepo.ListAvailable(ctx)
	if err != nil {
		return err
	}

	heartbeatThreshold := time.Now().Add(-2 * s.config.HeartbeatInterval)

	for _, worker := range workers {
		if worker.LastHeartbeat.Before(heartbeatThreshold) {
			log.Printf("Worker %s appears to be offline (last heartbeat: %s)", worker.ID, worker.LastHeartbeat)

			// If the worker has a task assigned, mark it for retry
			if worker.CurrentTaskID != nil {
				task, err := s.taskRepo.GetByID(ctx, *worker.CurrentTaskID)
				if err != nil {
					log.Printf("Error fetching task %s: %v", *worker.CurrentTaskID, err)
					continue
				}

				// Increment retry count if under max retries
				if task.RetryCount < task.MaxRetries {
					task.Status = models.TaskStatusPending
					task.WorkerID = nil
					task.RetryCount++
					task.UpdatedAt = time.Now()

					if err := s.taskRepo.Update(ctx, task); err != nil {
						log.Printf("Error updating task %s for retry: %v", task.ID, err)
					} else {
						log.Printf("Task %s marked for retry (attempt %d of %d)", task.ID, task.RetryCount, task.MaxRetries)
					}
				} else {
					// Max retries reached, mark as failed
					task.Status = models.TaskStatusFailed
					task.Error = "Task failed after maximum retry attempts"
					task.UpdatedAt = time.Now()

					if err := s.taskRepo.Update(ctx, task); err != nil {
						log.Printf("Error marking task %s as failed: %v", task.ID, err)
					}
				}
			}

			// Mark worker as offline
			worker.Status = "offline"
			worker.CurrentTaskID = nil

			if err := s.workerRepo.Update(ctx, worker); err != nil {
				log.Printf("Error marking worker %s as offline: %v", worker.ID, err)
			}
		}
	}

	return nil
}

// canWorkerHandleTask checks if worker can handle the task
func (s *Scheduler) canWorkerHandleTask(worker *models.Worker, task *models.Task) bool {
	// Basic checks:
	// 1. Worker must be available
	// 2. Worker must not already have a task assigned
	if worker.Status != "available" || worker.CurrentTaskID != nil {
		return false
	}

	// Check if worker supports the task model
	if task.ModelName != "" {
		supported := false
		for _, capability := range worker.Capabilities {
			if capability == task.ModelName {
				supported = true
				break
			}
		}
		if !supported {
			return false
		}
	}

	// Get current resource usage
	cpu, memory, gpu := s.resources.Get()

	// Check resource requirements
	requiredCPU := 0.5 // Default CPU requirement
	requiredMem := 1.0 // Default memory requirement in GB
	requiredGPU := 0.3 // Default GPU requirement

	if task.ModelName != "" {
		// Adjust for model-specific requirements
		requiredCPU = 1.0
		requiredMem = 2.0
		requiredGPU = 0.7
	}

	// Check if worker has sufficient resources
	// Convert memory from bytes to GB for comparison
	availableMemGB := float64(worker.AvailableMemory) / 1024 / 1024 / 1024
	if cpu+requiredCPU > worker.AvailableCPU ||
		memory+requiredMem > availableMemGB ||
		gpu+requiredGPU > worker.AvailableGPU {
		return false
	}

	return true
}

// updateWorkerResources updates tracked resources after task assignment
func (s *Scheduler) updateWorkerResources(worker *models.Worker, task *models.Task) {
	// Calculate resource requirements (must match canWorkerHandleTask)
	cpuReq := 0.5 // Default CPU requirement
	memReq := 1.0 // Default memory requirement in GB
	gpuReq := 0.3 // Default GPU requirement

	if task.ModelName != "" {
		// Model-specific requirements
		cpuReq = 1.0
		memReq = 2.0
		gpuReq = 0.7
	}

	// Update metrics
	s.metrics.ResourceUsage.WithLabelValues("assigned_tasks", "count").Inc()
	s.metrics.ResourceRequests.WithLabelValues("cpu", "cores").Set(cpuReq)
	s.metrics.ResourceRequests.WithLabelValues("memory", "gb").Set(memReq)
	s.metrics.ResourceRequests.WithLabelValues("gpu", "percent").Set(gpuReq)

	// Update global resource tracking
	s.resources.Update(cpuReq, memReq, gpuReq)

	// Update worker's available resources (in-memory only)
	worker.AvailableCPU -= cpuReq
	worker.AvailableMemory -= int64(memReq * 1024 * 1024 * 1024) // Convert GB to bytes
	worker.AvailableGPU -= gpuReq
}
