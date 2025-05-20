# AI Job Scheduling System

A distributed task scheduling system for large language model inference and training tasks, built with Go and Python.

## System Architecture

- **Go Backend**: Handles API, scheduling, and worker management
- **Python Workers**: Execute model inference and training tasks
- **Database**: Store task status, results, and configurations

![Architecture Diagram](docs/architecture.png)

## Components

- **API Server**: RESTful API for task submission, status checking, and management
- **Scheduler**: Distributes tasks to available workers based on priority and resource requirements
- **Worker Pool**: Manages Python worker processes that execute model tasks
- **Database**: Stores task information, configurations, and results
- **Monitoring**: Track system performance, worker health, and task status
- **LLM Server**: Python FastAPI server for model inference

## Features

- **Task Scheduling**: Priority-based scheduling of LLM tasks
- **Resource Management**: Tracks available resources on worker nodes
- **Fault Tolerance**: Worker health monitoring and task retries
- **Task Queuing**: Handles workload spikes with configurable queues
- **Model Variety**: Supports both API-based (OpenAI) and local models
- **Scalability**: Can scale horizontally by adding more workers
- **Monitoring**: Tracks task progress and worker health

## Getting Started

### Prerequisites

- Go 1.20 or later
- Python 3.8 or later
- PostgreSQL 14 or later
- Docker and Docker Compose (for containerized deployment)

### Installation

#### Option 1: Local Development

1. Clone the repository
2. Install Go dependencies:
   ```
   go mod tidy
   ```
3. Set up Python environment:
   ```
   cd scripts/python
   python -m venv venv
   source venv/bin/activate  # Or venv\Scripts\activate on Windows
   pip install -r requirements.txt
   ```
4. Configure database connection in `config/config.yaml`
5. Initialize the database:
   ```
   psql -h localhost -U postgres -f deployments/schema.sql
   ```
6. Run the server:
   ```
   go run cmd/server/main.go
   ```
7. Run the worker:
   ```
   go run cmd/worker/main.go
   ```

#### Option 2: Docker Deployment

1. Clone the repository
2. Build and start the containers:
   ```
   docker-compose up -d
   ```

### Configuration

The system is configured through `config/config.yaml`. Key configuration options include:

- **Server settings**: Host, port, timeouts
- **Database connection**: Credentials, pool sizes
- **Worker settings**: Number of workers, polling intervals
- **LLM settings**: Available models and their requirements

### CLI Usage

The system includes a CLI tool for administrative tasks:

```
# Create a task
go run cmd/admin/main.go create-task --name "Text Generation" --model "gpt-3.5-turbo" --input '{"prompt":"Hello, world!"}'

# List all tasks
go run cmd/admin/main.go list-tasks

# Get task details
go run cmd/admin/main.go get-task <task-id>

# Cancel a task
go run cmd/admin/main.go cancel-task <task-id>

# List workers
go run cmd/admin/main.go list-workers
```

## API Endpoints

The system exposes a RESTful API:

- `POST /api/v1/tasks`: Submit a new task
- `GET /api/v1/tasks/{id}`: Get task status
- `GET /api/v1/tasks`: List all tasks
- `DELETE /api/v1/tasks/{id}`: Cancel a task
- `POST /api/v1/workers`: Register a new worker
- `PUT /api/v1/workers/{id}/heartbeat`: Update worker heartbeat
- `PUT /api/v1/workers/{id}/status`: Update worker status

## Project Structure

```
.
├── cmd/
│   ├── admin/       # Admin CLI tool
│   ├── server/      # API server
│   └── worker/      # Worker manager
├── config/          # Configuration files
├── deployments/     # Deployment files (SQL schemas, etc.)
├── docs/            # Documentation
├── internal/
│   ├── api/         # API server implementation
│   ├── database/    # Database access layer
│   ├── models/      # Data models
│   ├── scheduler/   # Task scheduler
│   └── worker/      # Worker implementation
├── pkg/
│   ├── config/      # Configuration utilities
│   ├── llm/         # LLM client
│   ├── queue/       # Queue implementation
│   └── utils/       # Common utilities
└── scripts/
    └── python/      # Python scripts for model inference
```

## License

MIT
