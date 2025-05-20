package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a client for the Model Context Protocol API
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new MCP client
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ClientWithTimeout creates a new MCP client with a custom timeout
func NewClientWithTimeout(baseURL string, timeout time.Duration) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Model represents model information in MCP
type Model struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Provider     string                 `json:"provider"`
	Capabilities []string               `json:"capabilities"`
	Config       map[string]interface{} `json:"config"`
}

// ContextNode represents a node in the MCP context tree
type ContextNode struct {
	ID          string                 `json:"id"`
	Content     string                 `json:"content"`
	ContentType string                 `json:"content_type"`
	Metadata    map[string]interface{} `json:"metadata"`
	Parent      *string                `json:"parent"`
	Children    []string               `json:"children"`
}

// CreateContextRequest represents a request to create a new context
type CreateContextRequest struct {
	ModelID       string                 `json:"model_id"`
	Nodes         []ContextNode          `json:"nodes"`
	Metadata      map[string]interface{} `json:"metadata"`
	ReturnContext bool                   `json:"return_context"`
}

// CreateContextResponse represents a response after creating a context
type CreateContextResponse struct {
	ContextID string                 `json:"context_id"`
	Model     Model                  `json:"model"`
	Nodes     []ContextNode          `json:"nodes"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// PromptRequest represents a request to add a prompt to a context
type PromptRequest struct {
	ContextID string                 `json:"context_id"`
	Prompt    string                 `json:"prompt"`
	PromptID  *string                `json:"prompt_id,omitempty"`
	ParentID  *string                `json:"parent_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
	Stream    bool                   `json:"stream"`
}

// AddPromptRequest represents a request to add a prompt to a context (client version)
type AddPromptRequest struct {
	Prompt   string                 `json:"prompt"`
	PromptID string                 `json:"prompt_id,omitempty"`
	ParentID string                 `json:"parent_id,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Stream   bool                   `json:"stream"`
}

// AddPromptResponse represents a response to an add prompt request
type AddPromptResponse struct {
	PromptID   string                 `json:"prompt_id"`
	Completion string                 `json:"completion"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// AddNodeRequest represents a request to add a node to a context
type AddNodeRequest struct {
	ContextID string      `json:"context_id"`
	Node      ContextNode `json:"node"`
}

// AddNodeResponse represents a response after adding a node
type AddNodeResponse struct {
	ContextID string      `json:"context_id"`
	Node      ContextNode `json:"node"`
}

// DeleteNodeResponse represents a response after deleting a node
type DeleteNodeResponse struct {
	ContextID string `json:"context_id"`
	Deleted   bool   `json:"deleted"`
}

// ListContextsResponse represents a response with list of contexts
type ListContextsResponse struct {
	Contexts []string `json:"contexts"`
}

// GetContextResponse represents a response with context details
type GetContextResponse struct {
	ContextID string                 `json:"context_id"`
	Model     Model                  `json:"model"`
	Nodes     []ContextNode          `json:"nodes"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// DeleteContextResponse represents a response after deleting a context
type DeleteContextResponse struct {
	ContextID string `json:"context_id"`
	Deleted   bool   `json:"deleted"`
}

// CreateContext creates a new MCP context
func (c *Client) CreateContext(ctx context.Context, req CreateContextRequest) (*CreateContextResponse, error) {
	url := fmt.Sprintf("%s/v1/contexts", c.BaseURL)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result CreateContextResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// Prompt sends a prompt to an existing context
func (c *Client) Prompt(ctx context.Context, contextID string, req PromptRequest) (*PromptResponse, error) {
	url := fmt.Sprintf("%s/v1/contexts/%s/prompt", c.BaseURL, contextID)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result PromptResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// AddNode adds a node to an existing context
func (c *Client) AddNode(ctx context.Context, contextID string, node ContextNode) (*AddNodeResponse, error) {
	url := fmt.Sprintf("%s/v1/contexts/%s/nodes", c.BaseURL, contextID)

	req := AddNodeRequest{
		ContextID: contextID,
		Node:      node,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result AddNodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// DeleteNode deletes a node from an existing context
func (c *Client) DeleteNode(ctx context.Context, contextID, nodeID string) (*DeleteNodeResponse, error) {
	url := fmt.Sprintf("%s/v1/contexts/%s/nodes/%s", c.BaseURL, contextID, nodeID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result DeleteNodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// ListContexts lists all active contexts
func (c *Client) ListContexts(ctx context.Context) (*ListContextsResponse, error) {
	url := fmt.Sprintf("%s/v1/contexts", c.BaseURL)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result ListContextsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetContext gets a context by ID
func (c *Client) GetContext(ctx context.Context, contextID string) (*GetContextResponse, error) {
	url := fmt.Sprintf("%s/v1/contexts/%s", c.BaseURL, contextID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result GetContextResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// DeleteContext deletes a context by ID
func (c *Client) DeleteContext(ctx context.Context, contextID string) (*DeleteContextResponse, error) {
	url := fmt.Sprintf("%s/v1/contexts/%s", c.BaseURL, contextID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result DeleteContextResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// CheckHealth checks if the MCP server is healthy
func (c *Client) CheckHealth(ctx context.Context) (bool, error) {
	url := fmt.Sprintf("%s/health", c.BaseURL)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// ListModels lists all loaded models
func (c *Client) ListModels(ctx context.Context) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/models", c.BaseURL)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
