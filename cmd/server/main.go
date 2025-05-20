package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ai-job/internal/api"
	"ai-job/internal/database"
	"ai-job/internal/scheduler"
	"ai-job/pkg/config"
)

func main() {
	log.Println("Starting AI Job Scheduler Server")

	// Load configuration
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := database.New(database.Config{
		Driver:          cfg.Database.Driver,
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		Name:            cfg.Database.Name,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create repositories
	taskRepo := database.NewTaskRepository(db)
	workerRepo := database.NewWorkerRepository(db)
	mcpTaskRepo := database.NewMCPTaskRepository(db)
	mcpContextRepo := database.NewMCPContextRepository(db)

	// Create and start scheduler
	schedulerSvc := scheduler.New(taskRepo, workerRepo, scheduler.Config{
		PollInterval:      cfg.Worker.PollInterval,
		MaxTasks:          cfg.Worker.MaxWorkers,
		HeartbeatInterval: cfg.Worker.HeartbeatInterval,
		TaskTimeout:       cfg.Worker.TaskTimeout,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := schedulerSvc.Start(ctx); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}

	// Create and start API server
	server := api.New(taskRepo, workerRepo, mcpTaskRepo, mcpContextRepo, api.Config{
		Host:           cfg.Server.Host,
		Port:           cfg.Server.Port,
		Timeout:        cfg.Server.Timeout,
		MaxRequestSize: cfg.Server.MaxRequestSize,
		MCPServerURL:   cfg.MCP.ServerURL,
	})

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start API server: %v", err)
		}
	}()

	log.Println("System is ready to accept requests")

	<-quit
	log.Println("Shutting down...")

	// Stop the scheduler
	schedulerSvc.Stop()

	log.Println("Server gracefully stopped")
}
