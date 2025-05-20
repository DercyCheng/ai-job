package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// PromptStreamResponse represents a streaming response chunk from a prompt
type PromptStreamResponse struct {
	Completion string                 `json:"completion"`
	PromptID   string                 `json:"prompt_id,omitempty"`
	IsFinal    bool                   `json:"is_final,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// PromptStream represents a stream of prompt responses
type PromptStream struct {
	ctx    context.Context
	reader io.ReadCloser
	client *http.Client
}

// NewPromptStream creates a new prompt stream
func NewPromptStream(ctx context.Context, reader io.ReadCloser, client *http.Client) *PromptStream {
	return &PromptStream{
		ctx:    ctx,
		reader: reader,
		client: client,
	}
}

// Recv receives the next chunk of the prompt response
func (s *PromptStream) Recv() (*PromptStreamResponse, error) {
	// Check if context is canceled
	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	default:
	}

	// Read one SSE event (up to a blank line)
	var buffer []byte
	var eventBuffer []byte
	isEnd := false

	for !isEnd {
		temp := make([]byte, 1)
		_, err := s.reader.Read(temp)
		if err != nil {
			if err == io.EOF {
				if len(buffer) > 0 {
					// Process final buffer before returning EOF
					eventBuffer = append(eventBuffer, buffer...)
					isEnd = true
					break
				}
				// Regular EOF
				return nil, io.EOF
			}
			return nil, err
		}

		// Check for end of line
		if temp[0] == '\n' {
			if len(buffer) == 0 {
				// Blank line - end of event
				isEnd = true
				break
			}

			// Append line to event buffer
			eventBuffer = append(eventBuffer, buffer...)
			eventBuffer = append(eventBuffer, '\n')
			buffer = []byte{}
		} else {
			buffer = append(buffer, temp[0])
		}
	}

	// Process the event data
	event := string(eventBuffer)
	var data string
	for _, line := range splitLines(event) {
		if len(line) > 5 && line[:5] == "data:" {
			data = line[5:]
			// Remove leading space if present
			if len(data) > 0 && data[0] == ' ' {
				data = data[1:]
			}
			break
		}
	}

	if data == "" {
		// No data field found, try again
		return s.Recv()
	}

	// Parse the JSON data
	var response PromptStreamResponse
	if err := json.Unmarshal([]byte(data), &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stream response: %w", err)
	}

	return &response, nil
}

// Close closes the stream
func (s *PromptStream) Close() error {
	return s.reader.Close()
}

// Helper function to split a string by newlines
func splitLines(s string) []string {
	var lines []string
	var line []byte

	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, string(line))
			line = []byte{}
		} else {
			line = append(line, s[i])
		}
	}

	if len(line) > 0 {
		lines = append(lines, string(line))
	}

	return lines
}

// AddPromptStream adds a prompt to a context and returns a streaming response
func (c *Client) AddPromptStream(ctx context.Context, contextID string, req AddPromptRequest) (*PromptStream, error) {
	// Require stream mode
	req.Stream = true

	// Encode the request body
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create the HTTP request
	httpReq, err := c.newRequest(ctx, http.MethodPost, fmt.Sprintf("/v1/contexts/%s/prompt", contextID), reqBody)
	if err != nil {
		return nil, err
	}

	// Set headers for server-sent events
	httpReq.Header.Set("Accept", "text/event-stream")

	// Execute the request
	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Check the response status
	if httpResp.StatusCode != http.StatusOK {
		httpResp.Body.Close()
		return nil, c.handleErrorResponse(httpResp)
	}

	// Create and return the prompt stream
	return NewPromptStream(ctx, httpResp.Body, c.client), nil
}
