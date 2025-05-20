# Model Context Protocol (MCP) Integration

This document describes the integration of the Model Context Protocol (MCP) with the AI-Job scheduling system.

## Overview

The Model Context Protocol (MCP) is a standardized way of interacting with large language models (LLMs). It provides a unified interface for managing context, generating completions, and handling LLM tasks in a consistent manner.

Our system now fully supports MCP, allowing clients to:

1. Create and manage contexts for LLM interactions
2. Add prompts to contexts and receive completions
3. Manage context nodes (add/remove)
4. Delete contexts when no longer needed

## Architecture

The MCP integration consists of the following components:

1. **MCP Server (Python)**: A FastAPI server that implements the MCP specification and interfaces with LLMs
2. **MCP Client (Go)**: A Go client library for interacting with the MCP server
3. **MCP Handler (Go)**: An API handler for exposing MCP functionality through our main API
4. **MCP Worker (Python)**: A worker for processing MCP tasks
5. **Database Integration**: Tables and repositories for storing MCP tasks and contexts

## API Endpoints

The following MCP endpoints are available:

- `POST /api/v1/mcp/contexts`: Create a new context
- `GET /api/v1/mcp/contexts`: List all contexts
- `GET /api/v1/mcp/contexts/{contextID}`: Get a specific context
- `DELETE /api/v1/mcp/contexts/{contextID}`: Delete a context
- `POST /api/v1/mcp/contexts/{contextID}/prompt`: Add a prompt to a context
- `POST /api/v1/mcp/contexts/{contextID}/nodes`: Add a node to a context
- `DELETE /api/v1/mcp/contexts/{contextID}/nodes/{nodeID}`: Delete a node from a context
- `GET /api/v1/mcp/tasks`: List MCP tasks
- `GET /api/v1/mcp/tasks/{taskID}`: Get a specific MCP task
- `GET /api/v1/mcp/health`: Check MCP server health
- `GET /api/v1/mcp/models`: List available models

## Configuration

MCP can be configured in the `config.yaml` file:

```yaml
mcp:
  enabled: true                       # Enable/disable MCP support
  server_url: "http://localhost:8001" # URL of the MCP server
  api_version: "v1"                   # MCP API version
  max_contexts: 100                   # Maximum number of contexts to store
  timeout: "60s"                      # Timeout for MCP requests
```

## Usage Examples

### Creating a Context

```bash
curl -X POST http://localhost:8080/api/v1/mcp/contexts \
  -H "Content-Type: application/json" \
  -d '{
    "model_id": "gpt-3.5-turbo",
    "nodes": [
      {
        "content": "You are a helpful assistant.",
        "content_type": "text",
        "metadata": {"role": "system"}
      }
    ],
    "metadata": {"session": "user123"},
    "return_context": true,
    "user_id": "user123",
    "priority": "medium"
  }'
```

### Adding a Prompt

```bash
curl -X POST http://localhost:8080/api/v1/mcp/contexts/{contextID}/prompt \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "What is the capital of France?",
    "metadata": {"conversation_id": "12345"},
    "stream": false,
    "user_id": "user123",
    "priority": "high"
  }'
```

### Deleting a Context

```bash
curl -X DELETE http://localhost:8080/api/v1/mcp/contexts/{contextID}
```

## MCP Task Types

The system supports the following MCP task types:

1. `mcp.create_context`: Create a new context
2. `mcp.add_prompt`: Add a prompt to a context
3. `mcp.add_node`: Add a node to a context
4. `mcp.delete_node`: Delete a node from a context
5. `mcp.delete_context`: Delete a context

## Worker Process

The MCP Worker processes MCP tasks as follows:

1. Worker polls for scheduled MCP tasks
2. When a task is found, it updates the task status to "running"
3. It processes the task based on its type (create context, add prompt, etc.)
4. It communicates with the MCP server to perform the actual operation
5. It updates the task status to "completed" or "failed" based on the result
6. The task output or error is stored in the database

## Streaming Support

MCP supports streaming responses for long completions. When the `stream` parameter is set to `true` in an add prompt request, the completion will be streamed back to the client.

## Client Rate Limiting

The MCP client implements rate limiting and retry logic to handle transient failures and avoid overloading the MCP server.

## Monitoring and Metrics

The system collects the following metrics for MCP operations:

- Task processing time
- Success/failure rates
- Response times
- Queue depths
- Context usage statistics

These metrics are available through the monitoring dashboard.

## Troubleshooting

Common issues and their solutions:

1. **MCP Server Not Available**: Ensure the MCP server is running and the `server_url` in the configuration is correct.
2. **Model Not Found**: Check that the requested model is available in the MCP server.
3. **Context Not Found**: The context may have expired or been deleted. Create a new context.
4. **Rate Limiting**: The client may be sending too many requests. Implement backoff strategy.
5. **Worker Not Processing Tasks**: Check that the MCP worker is running and properly registered.

For more information, refer to the MCP specification documentation.
