package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// AddPrompt adds a prompt to a context and returns the completion
func (c *Client) AddPrompt(ctx context.Context, contextID string, req AddPromptRequest) (AddPromptResponse, error) {
	// Create request body
	reqBody, err := json.Marshal(req)
	if err != nil {
		return AddPromptResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create the HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/v1/contexts/%s/prompt", c.BaseURL, contextID),
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return AddPromptResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute the request
	httpResp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return AddPromptResponse{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer httpResp.Body.Close()

	// Check for errors
	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		return AddPromptResponse{}, fmt.Errorf("unexpected status code: %d, body: %s", httpResp.StatusCode, string(body))
	}

	// Parse the response
	var respObj AddPromptResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&respObj); err != nil {
		return AddPromptResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return respObj, nil
}
