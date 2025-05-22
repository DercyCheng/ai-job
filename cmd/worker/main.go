package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"ai-job/pkg/config"
	"ai-job/pkg/logger"

	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration first
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Then initialize logger with config
	logCfg := logger.ConvertConfig(cfg.Logging)
	logger, err := logger.NewLogger(logCfg)
	if err != nil {
		logrus.Fatalf("Failed to initialize logger: %v", err)
	}
	logrus.SetOutput(logger.Writer())
	logrus.Info("Starting AI Job Worker Manager")

	// Get the root directory of the application
	rootDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// Path to the Python worker scripts
	pythonWorkerPath := filepath.Join(rootDir, "scripts", "python", "worker.py")
	mcpWorkerPath := filepath.Join(rootDir, "scripts", "python", "mcp_worker.py")
	configPath := filepath.Join(rootDir, "config", "config.yaml")

	// Set up signal handling for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Command to run the Python worker
	workerCmd := exec.Command(
		filepath.Join(rootDir, "venv", "bin", "python"),
		pythonWorkerPath,
		"--config", configPath,
	)

	// Set the current working directory
	workerCmd.Dir = rootDir

	// Redirect stdout and stderr
	workerCmd.Stdout = os.Stdout
	workerCmd.Stderr = os.Stderr

	// Start the worker process
	log.Println("Starting Python worker process")
	if err := workerCmd.Start(); err != nil {
		log.Fatalf("Failed to start worker process: %v", err)
	}

	// Variable to hold MCP worker process
	var mcpWorkerCmd *exec.Cmd

	// Start MCP worker if enabled
	if cfg.MCP.Enabled {
		mcpWorkerCmd = exec.Command(
			filepath.Join(rootDir, "venv", "bin", "python"),
			mcpWorkerPath,
			"--config", configPath,
		)

		// Set the current working directory
		mcpWorkerCmd.Dir = rootDir

		// Redirect stdout and stderr
		mcpWorkerCmd.Stdout = os.Stdout
		mcpWorkerCmd.Stderr = os.Stderr

		// Start the MCP worker process
		log.Println("Starting MCP worker process")
		if err := mcpWorkerCmd.Start(); err != nil {
			log.Fatalf("Failed to start MCP worker process: %v", err)
		}
	}

	// Handle graceful shutdown
	go func() {
		<-quit
		log.Println("Shutting down...")

		// Send SIGTERM to the worker processes
		if err := workerCmd.Process.Signal(syscall.SIGTERM); err != nil {
			log.Printf("Failed to send SIGTERM to worker process: %v", err)
			workerCmd.Process.Kill()
		}

		// If MCP worker is running, shut it down
		if mcpWorkerCmd != nil && mcpWorkerCmd.Process != nil {
			if err := mcpWorkerCmd.Process.Signal(syscall.SIGTERM); err != nil {
				log.Printf("Failed to send SIGTERM to MCP worker process: %v", err)
				mcpWorkerCmd.Process.Kill()
			}
		}
	}()

	// Wait for the worker process to finish
	if err := workerCmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			log.Printf("Worker process exited with code %d", exitErr.ExitCode())
		} else {
			log.Printf("Worker process error: %v", err)
		}
	}

	// If MCP worker is running, wait for it to finish
	if mcpWorkerCmd != nil && mcpWorkerCmd.Process != nil {
		if err := mcpWorkerCmd.Wait(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				log.Printf("MCP worker process exited with code %d", exitErr.ExitCode())
			} else {
				log.Printf("MCP worker process error: %v", err)
			}
		}
	}

	log.Println("Worker manager stopped")
}
