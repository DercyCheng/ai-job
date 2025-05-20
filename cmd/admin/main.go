package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"ai-job/internal/database"
	"ai-job/internal/models"
	"ai-job/pkg/config"

	"github.com/google/uuid"
)

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	Run         func(args []string) error
}

func main() {
	log.Println("AI Job Admin Tool")

	// Define available commands
	commands := []Command{
		{
			Name:        "create-task",
			Description: "Create a new task",
			Run:         createTask,
		},
		{
			Name:        "list-tasks",
			Description: "List all tasks",
			Run:         listTasks,
		},
		{
			Name:        "get-task",
			Description: "Get task details",
			Run:         getTask,
		},
		{
			Name:        "cancel-task",
			Description: "Cancel a task",
			Run:         cancelTask,
		},
		{
			Name:        "list-workers",
			Description: "List all workers",
			Run:         listWorkers,
		},
		{
			Name:        "init-db",
			Description: "Initialize the database",
			Run:         initDB,
		},
	}

	// Create a map of commands for easy lookup
	commandMap := make(map[string]Command)
	for _, cmd := range commands {
		commandMap[cmd.Name] = cmd
	}

	// Print usage if no arguments provided
	if len(os.Args) < 2 {
		printUsage(commands)
		os.Exit(1)
	}

	// Get the command name from arguments
	cmdName := os.Args[1]

	// Look up the command
	cmd, ok := commandMap[cmdName]
	if !ok {
		fmt.Printf("Unknown command: %s\n", cmdName)
		printUsage(commands)
		os.Exit(1)
	}

	// Run the command
	if err := cmd.Run(os.Args[2:]); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// printUsage prints the usage information
func printUsage(commands []Command) {
	fmt.Println("Usage: admin [command] [arguments]")
	fmt.Println("\nAvailable commands:")
	for _, cmd := range commands {
		fmt.Printf("  %-15s %s\n", cmd.Name, cmd.Description)
	}
}

// createTask creates a new task
func createTask(args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("create-task", flag.ExitOnError)
	name := fs.String("name", "", "Task name")
	desc := fs.String("desc", "", "Task description")
	model := fs.String("model", "", "Model name")
	priority := fs.Int("priority", int(models.TaskPriorityNormal), "Task priority (1=low, 2=normal, 3=high, 4=critical)")
	userID := fs.String("user", "admin", "User ID")
	input := fs.String("input", "", "Task input (JSON string)")
	inputFile := fs.String("input-file", "", "Task input file (JSON file)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Validate required flags
	if *name == "" {
		return fmt.Errorf("name is required")
	}
	if *model == "" {
		return fmt.Errorf("model is required")
	}

	// Get input data
	var inputData []byte
	if *inputFile != "" {
		// Read input from file
		var err error
		inputData, err = os.ReadFile(*inputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
	} else if *input != "" {
		// Use input string
		inputData = []byte(*input)
	} else {
		return fmt.Errorf("either input or input-file is required")
	}

	// Validate input data as JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal(inputData, &jsonData); err != nil {
		return fmt.Errorf("invalid JSON input: %w", err)
	}

	// Load configuration
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Connect to database
	db, err := connectToDatabase(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	// Create task repository
	taskRepo := database.NewTaskRepository(db)

	// Create the task
	task := models.NewTask(*name, *model, *userID, models.TaskPriority(*priority), inputData)
	task.Description = *desc

	// Save the task
	ctx := context.Background()
	if err := taskRepo.Create(ctx, task); err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	fmt.Printf("Task created with ID: %s\n", task.ID)
	return nil
}

// listTasks lists all tasks
func listTasks(args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("list-tasks", flag.ExitOnError)
	status := fs.String("status", "", "Filter by status")
	limit := fs.Int("limit", 10, "Maximum number of tasks to list")
	offset := fs.Int("offset", 0, "Offset for pagination")
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Load configuration
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Connect to database
	db, err := connectToDatabase(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	// Create task repository
	taskRepo := database.NewTaskRepository(db)

	// Get tasks
	ctx := context.Background()
	var statusFilter *models.TaskStatus
	if *status != "" {
		s := models.TaskStatus(*status)
		statusFilter = &s
	}

	tasks, err := taskRepo.List(ctx, statusFilter, *limit, *offset)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	// Print tasks
	fmt.Printf("Found %d tasks:\n", len(tasks))
	fmt.Println("ID\tName\tStatus\tModel\tCreated\tUpdated")
	for _, task := range tasks {
		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\n",
			task.ID,
			task.Name,
			task.Status,
			task.ModelName,
			task.CreatedAt.Format(time.RFC3339),
			task.UpdatedAt.Format(time.RFC3339),
		)
	}

	return nil
}

// getTask gets a task by ID
func getTask(args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("get-task", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Validate arguments
	if fs.NArg() < 1 {
		return fmt.Errorf("task ID is required")
	}
	taskID := fs.Arg(0)

	// Validate task ID
	if _, err := uuid.Parse(taskID); err != nil {
		return fmt.Errorf("invalid task ID: %s", taskID)
	}

	// Load configuration
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Connect to database
	db, err := connectToDatabase(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	// Create task repository
	taskRepo := database.NewTaskRepository(db)

	// Get the task
	ctx := context.Background()
	task, err := taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Print task details
	fmt.Printf("Task ID: %s\n", task.ID)
	fmt.Printf("Name: %s\n", task.Name)
	fmt.Printf("Description: %s\n", task.Description)
	fmt.Printf("Model: %s\n", task.ModelName)
	fmt.Printf("Status: %s\n", task.Status)
	fmt.Printf("Priority: %d\n", task.Priority)
	fmt.Printf("Created: %s\n", task.CreatedAt.Format(time.RFC3339))
	fmt.Printf("Updated: %s\n", task.UpdatedAt.Format(time.RFC3339))

	if task.StartedAt != nil {
		fmt.Printf("Started: %s\n", task.StartedAt.Format(time.RFC3339))
	}

	if task.CompletedAt != nil {
		fmt.Printf("Completed: %s\n", task.CompletedAt.Format(time.RFC3339))
	}

	if task.WorkerID != nil {
		fmt.Printf("Worker ID: %s\n", *task.WorkerID)
	}

	if task.Error != "" {
		fmt.Printf("Error: %s\n", task.Error)
	}

	// Print input/output if available
	if len(task.Input) > 0 {
		fmt.Println("\nInput:")
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, task.Input, "", "  "); err == nil {
			fmt.Println(prettyJSON.String())
		} else {
			fmt.Println(string(task.Input))
		}
	}

	if len(task.Output) > 0 {
		fmt.Println("\nOutput:")
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, task.Output, "", "  "); err == nil {
			fmt.Println(prettyJSON.String())
		} else {
			fmt.Println(string(task.Output))
		}
	}

	return nil
}

// cancelTask cancels a task
func cancelTask(args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("cancel-task", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Validate arguments
	if fs.NArg() < 1 {
		return fmt.Errorf("task ID is required")
	}
	taskID := fs.Arg(0)

	// Validate task ID
	if _, err := uuid.Parse(taskID); err != nil {
		return fmt.Errorf("invalid task ID: %s", taskID)
	}

	// Load configuration
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Connect to database
	db, err := connectToDatabase(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	// Create task repository
	taskRepo := database.NewTaskRepository(db)

	// Get the task
	ctx := context.Background()
	task, err := taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Only allow cancellation of pending or scheduled tasks
	if task.Status != models.TaskStatusPending && task.Status != models.TaskStatusScheduled {
		return fmt.Errorf("cannot cancel task in status: %s", task.Status)
	}

	// Update task status
	task.Status = models.TaskStatusCancelled
	task.UpdatedAt = time.Now()

	// Save the task
	if err := taskRepo.Update(ctx, task); err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}

	fmt.Printf("Task %s cancelled successfully\n", taskID)
	return nil
}

// listWorkers lists all workers
func listWorkers(args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("list-workers", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Load configuration
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Connect to database
	db, err := connectToDatabase(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	// Create worker repository
	workerRepo := database.NewWorkerRepository(db)

	// Get workers
	ctx := context.Background()
	workers, err := workerRepo.ListAvailable(ctx)
	if err != nil {
		return fmt.Errorf("failed to list workers: %w", err)
	}

	// Print workers
	fmt.Printf("Found %d workers:\n", len(workers))
	fmt.Println("ID\tName\tStatus\tCapabilities\tLast Heartbeat\tTasks Handled")
	for _, worker := range workers {
		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%d\n",
			worker.ID,
			worker.Name,
			worker.Status,
			strings.Join(worker.Capabilities, ","),
			worker.LastHeartbeat.Format(time.RFC3339),
			worker.TotalTasksHandled,
		)
	}

	return nil
}

// initDB initializes the database
func initDB(args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("init-db", flag.ExitOnError)
	schemaFile := fs.String("schema", "deployments/schema.sql", "Schema file path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Load configuration
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Connect to database
	db, err := connectToDatabase(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	// Check if schema file exists
	if _, err := os.Stat(*schemaFile); os.IsNotExist(err) {
		return fmt.Errorf("schema file does not exist: %s", *schemaFile)
	}

	// Execute schema
	fmt.Println("Schema file found. Please execute the schema manually using a PostgreSQL client:")
	fmt.Printf("psql -h %s -p %d -U %s -d %s -f %s\n",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Name, *schemaFile)

	fmt.Println("Database connection successful")
	return nil
}

// connectToDatabase connects to the database
func connectToDatabase(cfg *config.Config) (*database.Database, error) {
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
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Ping the database to ensure connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
