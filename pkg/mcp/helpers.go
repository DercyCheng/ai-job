package mcp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

// newRequest is a helper function to create a new request with context
func (c *Client) newRequest(ctx context.Context, method, path string, body []byte) (*http.Request, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, path)
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewBuffer(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

// handleErrorResponse is a helper function to handle error responses
func (c *Client) handleErrorResponse(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
}
