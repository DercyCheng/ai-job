package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"ai-job/pkg/config"
)

func main() {
	log.Println("Starting AI Job Worker Manager")

	// Load configuration
	_, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Get the root directory of the application
	rootDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// Path to the Python worker script
	pythonWorkerPath := filepath.Join(rootDir, "scripts", "python", "worker.py")
	configPath := filepath.Join(rootDir, "config", "config.yaml")

	// Set up signal handling for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Command to run the Python worker
	cmd := exec.Command(
		filepath.Join(rootDir, "venv", "bin", "python"),
		pythonWorkerPath,
		"--config", configPath,
	)

	// Set the current working directory
	cmd.Dir = rootDir

	// Redirect stdout and stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the worker process
	log.Println("Starting Python worker process")
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start worker process: %v", err)
	}

	// Handle graceful shutdown
	go func() {
		<-quit
		log.Println("Shutting down...")

		// Send SIGTERM to the worker process
		if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
			log.Printf("Failed to send SIGTERM to worker process: %v", err)
			cmd.Process.Kill()
		}
	}()

	// Wait for the worker process to finish
	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			log.Printf("Worker process exited with code %d", exitErr.ExitCode())
		} else {
			log.Printf("Worker process error: %v", err)
		}
	}

	log.Println("Worker manager stopped")
}
