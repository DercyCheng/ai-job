-- Schema for AI Job Scheduler system with MCP support

-- Create extension for UUID type
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Tasks table
CREATE TABLE IF NOT EXISTS tasks (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    model_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    priority INT NOT NULL,
    input BYTEA,
    output BYTEA,
    error TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    worker_id VARCHAR(36),
    user_id VARCHAR(36) NOT NULL,
    timeout INT NOT NULL,
    retry_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 3
);

-- Create index on task status for faster queries
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_worker_id ON tasks(worker_id);
CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_tasks_priority_created ON tasks(priority DESC, created_at ASC);

-- Workers table
CREATE TABLE IF NOT EXISTS workers (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL,
    capabilities VARCHAR[] NOT NULL,
    current_task_id VARCHAR(36),
    last_heartbeat TIMESTAMP NOT NULL,
    registered_at TIMESTAMP NOT NULL,
    available_memory BIGINT,
    available_cpu FLOAT,
    available_gpu FLOAT,
    total_tasks_handled INT NOT NULL DEFAULT 0
);

-- Create index on worker status
CREATE INDEX IF NOT EXISTS idx_workers_status ON workers(status);
CREATE INDEX IF NOT EXISTS idx_workers_last_heartbeat ON workers(last_heartbeat);

-- Models table
CREATE TABLE IF NOT EXISTS models (
    name VARCHAR(100) PRIMARY KEY,
    provider VARCHAR(100) NOT NULL,
    model_path VARCHAR(255),
    max_context_length INT NOT NULL,
    quantization VARCHAR(20),
    required_memory BIGINT,
    requires_gpu BOOLEAN NOT NULL DEFAULT TRUE
);

-- Add foreign key constraints
ALTER TABLE tasks
    ADD CONSTRAINT fk_tasks_worker
    FOREIGN KEY (worker_id)
    REFERENCES workers(id)
    ON DELETE SET NULL;

ALTER TABLE workers
    ADD CONSTRAINT fk_workers_current_task
    FOREIGN KEY (current_task_id)
    REFERENCES tasks(id)
    ON DELETE SET NULL;

-- Insert some default models
INSERT INTO models (name, provider, model_path, max_context_length, quantization, required_memory, requires_gpu)
VALUES
    ('gpt-3.5-turbo', 'openai', NULL, 4096, NULL, 0, FALSE),
    ('llama-2-70b', 'local', '/models/llama-2-70b', 4096, 'int8', 20000000000, TRUE),
    ('mistral-7b', 'local', '/models/mistral-7b', 8192, 'int4', 8000000000, TRUE)
ON CONFLICT (name) DO NOTHING;

-- MCP Support Tables

-- Create MCP tasks table for Model Context Protocol
CREATE TABLE IF NOT EXISTS mcp_tasks (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    model_id VARCHAR(255) NOT NULL,
    context_id VARCHAR(36),
    type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    priority INT NOT NULL DEFAULT 2,
    input BYTEA,
    output BYTEA,
    error TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    worker_id VARCHAR(36),
    user_id VARCHAR(36) NOT NULL,
    timeout INT NOT NULL DEFAULT 1800,
    retry_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 3
);

-- Create MCP contexts table
CREATE TABLE IF NOT EXISTS mcp_contexts (
    id VARCHAR(36) PRIMARY KEY,
    model_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    data BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Create indexes for MCP tables
CREATE INDEX IF NOT EXISTS idx_mcp_tasks_status ON mcp_tasks(status);
CREATE INDEX IF NOT EXISTS idx_mcp_tasks_type ON mcp_tasks(type);
CREATE INDEX IF NOT EXISTS idx_mcp_tasks_context_id ON mcp_tasks(context_id);
CREATE INDEX IF NOT EXISTS idx_mcp_tasks_worker_id ON mcp_tasks(worker_id);
CREATE INDEX IF NOT EXISTS idx_mcp_tasks_user_id ON mcp_tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_mcp_tasks_created_at ON mcp_tasks(created_at);
CREATE INDEX IF NOT EXISTS idx_mcp_tasks_priority_created_at ON mcp_tasks(priority, created_at);
CREATE INDEX IF NOT EXISTS idx_mcp_tasks_model_id ON mcp_tasks(model_id);

CREATE INDEX IF NOT EXISTS idx_mcp_contexts_user_id ON mcp_contexts(user_id);
CREATE INDEX IF NOT EXISTS idx_mcp_contexts_model_id ON mcp_contexts(model_id);
CREATE INDEX IF NOT EXISTS idx_mcp_contexts_updated_at ON mcp_contexts(updated_at);

-- Add foreign key constraints for MCP tables
ALTER TABLE mcp_tasks
    ADD CONSTRAINT fk_mcp_tasks_worker
    FOREIGN KEY (worker_id)
    REFERENCES workers(id)
    ON DELETE SET NULL;

ALTER TABLE mcp_tasks
    ADD CONSTRAINT fk_mcp_tasks_context
    FOREIGN KEY (context_id)
    REFERENCES mcp_contexts(id)
    ON DELETE CASCADE;
