package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ai-job/internal/models"
	"ai-job/pkg/mcp"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestMCPClient tests the MCP client functionality
func TestMCPClient(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/health":
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		case "/v1/contexts":
			if r.Method == http.MethodPost {
				// Create context
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"context_id": uuid.New().String(),
					"model_id":   "test-model",
					"metadata":   map[string]interface{}{},
				})
			} else {
				// List contexts
				json.NewEncoder(w).Encode([]map[string]interface{}{
					{
						"context_id": uuid.New().String(),
						"model_id":   "test-model",
						"metadata":   map[string]interface{}{},
					},
				})
			}
		case "/v1/models":
			json.NewEncoder(w).Encode([]map[string]interface{}{
				{
					"id":          "test-model",
					"name":        "Test Model",
					"description": "Model for testing",
					"metadata":    map[string]interface{}{},
				},
			})
		default:
			// Handle context/{id} paths
			if r.Method == http.MethodGet && len(r.URL.Path) > 13 && r.URL.Path[:13] == "/v1/contexts/" {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"context_id": r.URL.Path[13:],
					"model_id":   "test-model",
					"metadata":   map[string]interface{}{},
					"nodes":      []interface{}{},
				})
			} else if r.Method == http.MethodDelete && len(r.URL.Path) > 13 && r.URL.Path[:13] == "/v1/contexts/" {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
				})
			} else if r.Method == http.MethodPost && len(r.URL.Path) > 20 && r.URL.Path[len(r.URL.Path)-7:] == "/prompt" {
				// Handle prompt endpoint
				json.NewEncoder(w).Encode(map[string]interface{}{
					"completion": "This is a test completion",
					"prompt_id":  uuid.New().String(),
					"metadata":   map[string]interface{}{},
				})
			} else if r.Method == http.MethodPost && len(r.URL.Path) > 19 && r.URL.Path[len(r.URL.Path)-6:] == "/nodes" {
				// Handle add node endpoint
				json.NewEncoder(w).Encode(map[string]interface{}{
					"id":      uuid.New().String(),
					"success": true,
				})
			} else if r.Method == http.MethodDelete && len(r.URL.Path) > 20 && r.URL.Path[13:19] == "/nodes" {
				// Handle delete node endpoint
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
				})
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}
	}))
	defer server.Close()

	// Create a client pointing to the test server
	client := mcp.NewClient(server.URL)

	// Test health check
	t.Run("HealthCheck", func(t *testing.T) {
		healthy, err := client.CheckHealth(context.Background())
		assert.NoError(t, err)
		assert.True(t, healthy)
	})

	// Test list models
	t.Run("ListModels", func(t *testing.T) {
		models, err := client.ListModels(context.Background())
		assert.NoError(t, err)
		assert.Len(t, models, 1)
		assert.Equal(t, "test-model", models[0].ID)
	})

	// Test create context
	t.Run("CreateContext", func(t *testing.T) {
		req := mcp.CreateContextRequest{
			ModelID:  "test-model",
			Metadata: map[string]interface{}{"test": "value"},
		}
		resp, err := client.CreateContext(context.Background(), req)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.ContextID)
		assert.Equal(t, "test-model", resp.ModelID)
	})

	// Test get context
	t.Run("GetContext", func(t *testing.T) {
		contextID := uuid.New().String()
		context, err := client.GetContext(context.Background(), contextID)
		assert.NoError(t, err)
		assert.Equal(t, contextID, context.ContextID)
		assert.Equal(t, "test-model", context.ModelID)
	})

	// Test delete context
	t.Run("DeleteContext", func(t *testing.T) {
		contextID := uuid.New().String()
		success, err := client.DeleteContext(context.Background(), contextID)
		assert.NoError(t, err)
		assert.True(t, success)
	})

	// Test add prompt
	t.Run("AddPrompt", func(t *testing.T) {
		contextID := uuid.New().String()
		req := mcp.AddPromptRequest{
			Prompt:   "This is a test prompt",
			Metadata: map[string]interface{}{"test": "value"},
		}
		resp, err := client.AddPrompt(context.Background(), contextID, req)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.PromptID)
		assert.NotEmpty(t, resp.Completion)
	})

	// Test add node
	t.Run("AddNode", func(t *testing.T) {
		contextID := uuid.New().String()
		node := mcp.ContextNode{
			Content:     "This is a test node",
			ContentType: "text",
			Metadata:    map[string]interface{}{"test": "value"},
		}
		nodeID, err := client.AddNode(context.Background(), contextID, node)
		assert.NoError(t, err)
		assert.NotEmpty(t, nodeID)
	})

	// Test delete node
	t.Run("DeleteNode", func(t *testing.T) {
		contextID := uuid.New().String()
		nodeID := uuid.New().String()
		success, err := client.DeleteNode(context.Background(), contextID, nodeID)
		assert.NoError(t, err)
		assert.True(t, success)
	})
}

// TestMCPWorker tests the MCP worker functionality
func TestMCPWorker(t *testing.T) {
	// Create a mock repository
	mcpTaskRepo := &MockMCPTaskRepository{}
	mcpContextRepo := &MockMCPContextRepository{}

	// Create a mock MCP client
	mockClient := &MockMCPClient{}

	// Create a worker with the mock components
	worker := NewMCPWorker(mcpTaskRepo, mcpContextRepo, mockClient, "test-worker-id")

	// Create a test MCP task
	task := &models.MCPTask{
		ID:        uuid.New().String(),
		Name:      "Test MCP Task",
		Type:      models.MCPTaskTypeCreateContext,
		Status:    models.TaskStatusScheduled,
		ModelID:   "test-model",
		UserID:    "test-user",
		Priority:  models.TaskPriorityMedium,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Input:     []byte(`{"model_id":"test-model","nodes":[],"metadata":{},"return_context":true}`),
	}

	// Process the task
	err := worker.ProcessTask(context.Background(), task)

	// Verify the task was processed
	assert.NoError(t, err)
	assert.Equal(t, models.TaskStatusCompleted, task.Status)
	assert.NotNil(t, task.Output)
	assert.NotNil(t, task.CompletedAt)
}

// Mock implementations for testing

type MockMCPTaskRepository struct{}

func (m *MockMCPTaskRepository) Create(ctx context.Context, task *models.MCPTask) error {
	return nil
}

func (m *MockMCPTaskRepository) GetByID(ctx context.Context, id string) (*models.MCPTask, error) {
	return &models.MCPTask{ID: id}, nil
}

func (m *MockMCPTaskRepository) Update(ctx context.Context, task *models.MCPTask) error {
	return nil
}

func (m *MockMCPTaskRepository) List(ctx context.Context, status *models.TaskStatus, limit, offset int) ([]*models.MCPTask, error) {
	return []*models.MCPTask{}, nil
}

func (m *MockMCPTaskRepository) GetPending(ctx context.Context, limit int) ([]*models.MCPTask, error) {
	return []*models.MCPTask{}, nil
}

type MockMCPContextRepository struct{}

func (m *MockMCPContextRepository) Create(ctx context.Context, context *models.MCPContext) error {
	return nil
}

func (m *MockMCPContextRepository) Get(ctx context.Context, id string) (*models.MCPContext, error) {
	return &models.MCPContext{ID: id}, nil
}

func (m *MockMCPContextRepository) Update(ctx context.Context, context *models.MCPContext) error {
	return nil
}

func (m *MockMCPContextRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *MockMCPContextRepository) List(ctx context.Context, limit, offset int) ([]*models.MCPContext, error) {
	return []*models.MCPContext{}, nil
}

type MockMCPClient struct{}

func (m *MockMCPClient) CreateContext(ctx context.Context, req mcp.CreateContextRequest) (mcp.CreateContextResponse, error) {
	return mcp.CreateContextResponse{
		ContextID: uuid.New().String(),
		ModelID:   req.ModelID,
		Metadata:  req.Metadata,
	}, nil
}

func (m *MockMCPClient) GetContext(ctx context.Context, contextID string) (mcp.GetContextResponse, error) {
	return mcp.GetContextResponse{
		ContextID: contextID,
		ModelID:   "test-model",
		Metadata:  map[string]interface{}{},
		Nodes:     []mcp.ContextNode{},
	}, nil
}

func (m *MockMCPClient) DeleteContext(ctx context.Context, contextID string) (bool, error) {
	return true, nil
}

func (m *MockMCPClient) ListContexts(ctx context.Context) ([]mcp.GetContextResponse, error) {
	return []mcp.GetContextResponse{}, nil
}

func (m *MockMCPClient) AddPrompt(ctx context.Context, contextID string, req mcp.AddPromptRequest) (mcp.AddPromptResponse, error) {
	return mcp.AddPromptResponse{
		PromptID:   uuid.New().String(),
		Completion: "Test completion",
		Metadata:   map[string]interface{}{},
	}, nil
}

func (m *MockMCPClient) AddNode(ctx context.Context, contextID string, node mcp.ContextNode) (string, error) {
	return uuid.New().String(), nil
}

func (m *MockMCPClient) DeleteNode(ctx context.Context, contextID, nodeID string) (bool, error) {
	return true, nil
}

func (m *MockMCPClient) ListModels(ctx context.Context) ([]mcp.Model, error) {
	return []mcp.Model{
		{
			ID:   "test-model",
			Name: "Test Model",
		},
	}, nil
}

func (m *MockMCPClient) CheckHealth(ctx context.Context) (bool, error) {
	return true, nil
}
